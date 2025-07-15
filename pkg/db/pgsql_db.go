package dbpkg

// pgsql_db.go contains the implementation of the PostgreSQL database connection and operations
// I am using the pgx library for PostgreSQL connections and operations.

import (
	"context"
	"fmt"
	"log"
	"strings"
	"targetad/pkg/redisstream"

	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// DBConfig holds database configuration values
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
}

// Dbconn holds the database connection pool
type Dbconn struct {
	Db     *pgxpool.Pool
	Config DBConfig
}

var (
	dbOnce sync.Once
	dbConn *Dbconn
	dbErr  error
)

// loadEnv loads environment variables from the .env file
// NOTE: for this demonstation I am using .env file for secret managements instead of using a secret manager. If production system
// I would definitely use a secret manager like AWS Secrets Manager or HashiCorp Vault.
func loadEnv() DBConfig {
	_ = godotenv.Load()

	return DBConfig{
		User:     getEnv("PG_USER", "postgres"),
		Password: getEnv("PG_PASSWORD", "postgres"),
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnv("PG_PORT", "5432"),
		Name:     getEnv("PG_DB", "random"),
		SSLMode:  getEnv("PG_SSLMODE", "disable"),
	}

}

// InitDB initializes the database connection. using sync.Once to ensure it is only called once every other call will return the same instance
// and will not reinitialize the connection.
func InitDB() (*Dbconn, error) {
	dbOnce.Do(func() {
		config := loadEnv()

		defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.SSLMode)
		targetDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.Name, config.SSLMode)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check if database exists
		conn, err := pgx.Connect(ctx, defaultDSN)
		if err != nil {
			dbErr = fmt.Errorf("failed to connect to PostgreSQL: %v", err)
			return
		}
		defer conn.Close(ctx)

		exists, err := databaseExists(ctx, conn, config.Name)
		if err != nil {
			dbErr = fmt.Errorf("failed to check database existence: %v", err)
			return
		}

		if !exists {
			log.Printf("Database %s does not exist. Creating it...", config.Name)
			if err := createDatabase(ctx, conn, config.Name); err != nil {

				dbErr = fmt.Errorf("failed to create database: %v", err)
				log.Fatalf("Error creating database: %v", dbErr)
				return
			}
			log.Println("Database created successfully!")
		}

		// Setup connection pool
		poolConfig, err := pgxpool.ParseConfig(targetDSN)
		if err != nil {
			dbErr = fmt.Errorf("failed to parse database config: %v", err)
			return
		}

		poolConfig.MaxConns = viper.GetInt32("pgsql.maxconn")
		poolConfig.MinConns = viper.GetInt32("pgsql.minconn")
		poolConfig.HealthCheckPeriod = time.Duration(viper.GetInt32("pgsql.healthCheckPeriod")) * time.Second
		poolConfig.MaxConnLifetime = time.Duration(viper.GetInt32("pgsql.maxConnLifetime")) * time.Minute
		poolConfig.MaxConnIdleTime = time.Duration(viper.GetInt32("pgsql.maxConnIdleTime")) * time.Minute

		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			dbErr = fmt.Errorf("failed to connect to target database: %v", err)
			return
		}

		log.Println("Connected to PostgreSQL database successfully.")
		dbConn = &Dbconn{Db: pool, Config: config}
	})

	return dbConn, dbErr
}

func GetConn() *Dbconn {
	return dbConn
}

// listenForNewDataInPgsql listens for new data's <tablename:primarykey> in PostgreSQL and once new data lands on our 2 tables
// it will write into our redis stream which all the microservices can listen to and update their cache with the latest data
func ListenForNewDataInPgsql(ctx context.Context) {
	for {
		err := listen(ctx)
		log.Printf("listener stopped: %v. Reconnecting in 2s...", err)
		time.Sleep(2 * time.Second)
	}
}

// I am creating a new db connection here because if the connection is lost in our infinite loop
// I will reconnect it. I am not adding this as part of the InitDB pool function because
// I want to keep the connection pool alive and not close it unlike init db's pool which closes and the connections in the pool
func listen(ctx context.Context) error {
	config := loadEnv()

	targetDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Name, config.SSLMode)

	conn, err := pgx.Connect(ctx, targetDSN)
	if err != nil {
		dbErr = fmt.Errorf("failed to connect to PostgreSQL: %v", err)
		return dbErr
	}
	_, err = conn.Exec(ctx, "LISTEN table_changes;")
	if err != nil {
		return fmt.Errorf("listen exec: %w", err)
	}

	log.Println("Connected and listening on 'table_changes'...")

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("wait failed: %w", err)
		}

		log.Printf("Change received:")
		table, id, isdeleted, err := parsePgsqlNotificationPayload(notification.Payload)
		if err != nil {
			log.Printf("Error parsing notification payload: %v", err)
			continue
		}
		err = redisstream.PushToRedisStream(table, id, isdeleted)
		if err != nil { // TODO: if it fails to push to redis stream we need to store the data in a queue and retry later
			log.Printf("Error pushing to Redis stream: %v", err)
			continue
		}

		log.Printf("Pushed to Redis stream: %s:%s", table, id)
	}
}

func parsePgsqlNotificationPayload(payload string) (string, string, bool, error) {
	parts := strings.Split(payload, ":")
	if len(parts) != 3 {
		return "", "", false, fmt.Errorf("invalid payload format: %s", payload)
	}
	tableName := parts[0]
	primaryKey := parts[1]
	isdeleted := parts[2] == "TRUE"
	return tableName, primaryKey, isdeleted, nil
}

func databaseExists(ctx context.Context, conn *pgx.Conn, name string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists)
	return exists, err
}

func createDatabase(ctx context.Context, conn *pgx.Conn, name string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", name))
	return err
}

func (dbconn *Dbconn) GetDB() *pgxpool.Pool {
	return dbconn.Db
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}

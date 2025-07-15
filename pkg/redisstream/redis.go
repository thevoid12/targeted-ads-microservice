package redisstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	dbpkg "targetad/pkg/db"
	"targetad/pkg/target"

	// "targetad/pkg/target"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

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
	config := dbpkg.LoadEnv()

	targetDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Name, config.SSLMode)

	conn, err := pgx.Connect(ctx, targetDSN)
	if err != nil {
		dbErr := fmt.Errorf("failed to connect to PostgreSQL: %v", err)
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
		err = PushToRedisStream(table, id, isdeleted)
		if err != nil { // TODO: if it fails to push to redis stream we need to store the data in a queue and retry later
			log.Printf("Error pushing to Redis stream: %v", err)
			continue
		}

		log.Printf("Pushed to Redis stream: %s:%s", table, id)
	}
}

func parsePgsqlNotificationPayload(payload string) (string, string, bool, error) {
	parts := strings.Split(payload, ":") // i am very sure the payload will be in the format of <table_name:primary_key:is_deleted>
	if len(parts) != 3 {
		return "", "", false, fmt.Errorf("invalid payload format: %s", payload)
	}
	tableName := parts[0]
	primaryKey := parts[1]
	isdeleted := parts[2] == "TRUE"
	return tableName, primaryKey, isdeleted, nil
}

func InitRedis(addr string) error {

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // I am not setting any password for now, but you can set it if needed
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConns"),
		ReadTimeout:  time.Duration(viper.GetInt("redis.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("redis.writeTimeout")) * time.Second,
		DialTimeout:  time.Duration(viper.GetInt("redis.dialTimeout")) * time.Second, // wait for 5 seconds to establish a dns connection
	})

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	if err := initRedisGroup(); err != nil {

	}
	return nil
}

func initRedisGroup() error {
	//$ to indicate that we want to read the last message in the stream
	// if the stream does not exist, it will be created automatically
	// if the group already exists, it will not be created again
	err := RedisClient.XGroupCreateMkStream(ctx, viper.GetString("redis.redisStream.streamName"), viper.GetString("redis.redisStream.consumerGroup"), "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func PushToRedisStream(table string, id string, isDeleted bool) error {
	_, err := RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: viper.GetString("redis.redisStream.streamName"),
		Values: map[string]interface{}{
			"table":      table,
			"id":         id,
			"is_deleted": isDeleted,
			"ts":         time.Now().UnixMilli(),
		},
	}).Result()
	return err
}

func StartRedisStreamListener(ctx context.Context) {
	for {
		streams, err := RedisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    viper.GetString("redis.redisStream.consumerGroup"),
			Consumer: viper.GetString("redis.redisStream.consumerName"),
			Streams:  []string{viper.GetString("redis.redisStream.streamName"), ">"},               // " > " = only get new messages not yet seen by this consumer. ones added after you start listening
			Block:    time.Duration(viper.GetInt("redis.redisStream.consumerBlock")) * time.Second, // time i should wait for new messages before returning an empty result
			Count:    viper.GetInt64("redis.redisStream.consumerCount"),                            // number of records to read in one go
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue // no new messages
			}
			log.Printf("Error reading from Redis stream: %v", err)
			time.Sleep(2 * time.Second) // wait before retrying
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				table := message.Values["table"].(string)
				id := message.Values["id"].(string)
				isDeleted := message.Values["is_deleted"].(bool)
				log.Printf("Received message from stream %s: table=%s, id=%s, is_deleted =%t", stream.Stream, table, id, isDeleted)
				// Process the message received from the stream
				err = target.ProcessRedisStreamDataService(ctx, table, id, isDeleted)
				if err != nil {
					log.Printf("Error processing Redis stream data: %v", err)
				} else {
					// Acknowledge the message after processing only after this acknowledgement
					// the message will be removed from the stream
					// if you do not acknowledge the message, it will be reprocessed again and again this is the reason why
					// I am using redis stream instead of redis pub/sub
					// because in pub/sub you cannot acknowledge the message
					// and it will be reprocessed again and again
					if err := RedisClient.XAck(ctx, viper.GetString("redis.redisStream.streamName"), viper.GetString("redis.redisStream.consumerGroup"), message.ID).Err(); err != nil {
						log.Printf("Error acknowledging message: %v", err)
					}
				}
			}
		}
	}
}

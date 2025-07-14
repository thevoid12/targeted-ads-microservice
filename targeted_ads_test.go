package main

// this file contains all the tests for this microservice
import (
	"log"
	dbpkg "targetad/pkg/db"
	"testing"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// loadEnvAndConfig loads environment variables and configuration from viper
// This function is used to set up the environment for testing
func loadEnvAndConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("there is a error in the path of config file", err)
		} else {
			log.Println("error laoding config file from viper", err)
		}
		return err
	}

	err = godotenv.Load(viper.GetString("app.env"))
	if err != nil {
		log.Println("there is a error loading environment variables", err)
		return err
	}

	return nil
}

// TestCheckPostgresSetup tests the PostgreSQL database connection setup
// It ensures that the database connection can be established successfully
func TestCheckPostgresSetup(t *testing.T) {
	err := loadEnvAndConfig()
	if err != nil {
		t.Fatalf("Failed to load environment variables: %v", err)
	}

	conn, err := dbpkg.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database connection: %v", err)
	}
	if conn == nil {
		t.Fatal("Database connection is nil")
	}
	t.Logf("Database connection established test successful: %v", conn)
}

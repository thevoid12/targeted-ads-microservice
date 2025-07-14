package main

import (
	"context"
	"fmt"
	"log"
	dbpkg "targetad/pkg/db"
	"targetad/pkg/target"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {

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
	}

	err = godotenv.Load(viper.GetString("app.env"))
	if err != nil {
		log.Println("there is a error loading environment variables", err)
		return
	}

	conn, err := dbpkg.InitDB()
	if err != nil {
		log.Println("error initializing database connection", err)
		return
	}

	if conn == nil {
		log.Println("database connection is nil")
		return
	}

	target.InitCache(context.TODO())
	fmt.Println(conn)
}

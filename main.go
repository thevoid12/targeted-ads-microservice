package main

import (
	"context"
	"log"
	"net/http"
	dbpkg "targetad/pkg/db"
	"targetad/pkg/target"
	"targetad/transport"

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
	// not all microservices need to listen for new data in pgsql and push it to redis stream
	// the others will just listen to the redis stream for new data and update its cache
	if viper.GetBool("app.isNotifyableMicroservice") {
		log.Println("this is a notifyable microservice, listening for new data in pgsql")
		go dbpkg.ListenForNewDataInPgsql(context.TODO())
	}

	handler := transport.NewHTTPHandler()
	log.Fatal(http.ListenAndServe(":8080", handler))

	// select {}
}

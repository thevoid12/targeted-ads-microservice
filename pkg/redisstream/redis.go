package redisstream

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	ctx         = context.Background()
	RedisClient *redis.Client
)

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
				log.Printf("Received message from stream %s: table=%s, id=%s", stream.Stream, table, id)
				// Process the message received from the stream

			}
		}
	}
}

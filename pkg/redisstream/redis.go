package redisstream

import (
	"context"
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

	return nil
}

func PushToRedisStream(table string, id string) error {
	_, err := RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "*",
		Values: map[string]interface{}{
			"table": table,
			"id":    id,
			"ts":    time.Now().UnixMilli(),
		},
	}).Result()
	return err
}

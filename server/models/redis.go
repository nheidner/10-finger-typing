package models

import (
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisHost     = "redis"
	redisPassword = ""
	redisDbname   = 0
	redisPort     = "6379"
)

var RedisClient *redis.Client

func init() {
	connectRedis()
}

func connectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisHost + ":" + redisPort,
		Password:     redisPassword,
		DB:           redisDbname,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	})
}

package models

import (
	"github.com/google/uuid"
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
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       redisDbname,
	})
}

func getUnstartedGamesKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String() + ":unstarted_games"
}

func getGameUserIdsKey(gameId uuid.UUID) string {
	return "games:" + gameId.String() + ":user_ids"
}

func getUserDataKey(gameId uuid.UUID, userId string) string {
	return "games:" + gameId.String() + ":user_data:" + userId
}

func getGameStatusKey(gameId uuid.UUID) string {
	return "games:" + gameId.String() + ":status"
}

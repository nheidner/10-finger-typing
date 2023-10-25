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

func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

func getRoomSubscriberConnectionsKey(roomId, userid uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userid) + ":conns"
}

func getRoomStreamKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":stream"
}

func getCurrentGameKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":current_game"
}

func getCurrentGameUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":user_ids"
}

func getTextIdsKey() string {
	return "text_ids"
}

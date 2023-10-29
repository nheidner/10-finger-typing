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

const (
	gameStatusField           = "status"
	subscriberStatusField     = "status"
	subscriberGameStatusField = "game_status"
	roomAdminIdField          = "admin_id"
	roomCreatedAtField        = "created_at"
	roomUpdatedAtField        = "updated_at"
	currentGameGameId         = "game_id"
	currentGameTextId         = "text_id"
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

// rooms:[room_id] hash: roomAdminId, createdAt, updatedAt
func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

// rooms:[room_id]:subscribers_ids set: user ids
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

// rooms:[room_id]:subscribers:[user_id] hash: status
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

// rooms:[room_id]:stream stream: action: "terminate/..", data: message stringified json
func getRoomStreamKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":stream"
}

// rooms:[room_id]:current_game hash: id, text_id, status
func getCurrentGameKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":current_game"
}

// rooms:[room_id]:subscribers:[user_id]:conns set: connection ids
func getRoomSubscriberConnectionsKey(roomId, userid uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userid) + ":conns"
}

// rooms:[room_id]:current_game:user_ids set: game user ids
func getCurrentGameUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":user_ids"
}

// text_ids set: text ids
func getTextIdsKey() string {
	return "text_ids"
}

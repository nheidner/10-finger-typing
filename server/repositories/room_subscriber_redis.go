package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	roomSubscriberStatusField     = "status"
	roomSubscriberGameStatusField = "game_status"
	roomSubscriberUsernameField   = "username"
)

type SubscriberStatus int

const (
	NilSubscriberStatus SubscriberStatus = iota
	InactiveSubscriberStatus
	ActiveSubscriberStatus
)

func (s *SubscriberStatus) String() string {
	return []string{"undefined", "inactive", "active"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type SubscriberGameStatus int

const (
	NilSubscriberGameStatus SubscriberGameStatus = iota
	UnstartedSubscriberGameStatus
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// rooms:[room_id]:subscribers_ids set: user ids
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

// rooms:[room_id]:subscribers:[user_id] hash: status, username, game_status
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

// rooms:[room_id]:subscribers:[user_id]:conns set: connection ids
func getRoomSubscriberConnectionsKey(roomId, userid uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userid) + ":conns"
}

type RoomSubscriberRedisRepository struct {
	redisClient *redis.Client
}

func NewRoomSubscriberRedisRepository(redisClient *redis.Client) *RoomSubscriberRedisRepository {
	return &RoomSubscriberRedisRepository{redisClient}
}

func (rr *RoomSubscriberRedisRepository) RemoveRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	err := rr.redisClient.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		return err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	return rr.redisClient.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err()
}

package repositories

import (
	"10-typing/models"
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	roomSubscriberStatusField     = "status"
	roomSubscriberGameStatusField = "game_status"
	roomSubscriberUsernameField   = "username"
)

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

func (rr *RoomSubscriberRedisRepository) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rr.redisClient.SAdd(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (rr *RoomSubscriberRedisRepository) RemoveRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rr.redisClient.SRem(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (rr *RoomSubscriberRedisRepository) SetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rr.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err()
}

func (rr *RoomSubscriberRedisRepository) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rr.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	if err != nil {
		return models.NilSubscriberStatus, err
	}

	return models.SubscriberStatus(status), nil
}

func (rr *RoomSubscriberRedisRepository) SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rr.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err()
}

func (rr *RoomSubscriberRedisRepository) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberGameStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rr.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberGameStatusField).Int()
	if err != nil {
		return models.NilSubscriberGameStatus, err
	}

	return models.SubscriberGameStatus(status), nil
}

func (rr *RoomSubscriberRedisRepository) GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]models.RoomSubscriber, error) {
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)

	r, err := rr.redisClient.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	roomSubscribers := make([]models.RoomSubscriber, 0, len(r))
	for _, roomSubscriberIdStr := range r {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, err
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		r, err := rr.redisClient.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		status := models.NilSubscriberStatus
		statusStr, ok := r[roomSubscriberStatusField]
		if ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err != nil {
				return nil, err
			}
			status = models.SubscriberStatus(statusInt)
		}

		subscriberGameStatus := models.NilSubscriberGameStatus
		subscriberGameStatusStr, ok := r[roomSubscriberGameStatusField]
		if ok {
			subscriberGameStatusInt, err := strconv.Atoi(subscriberGameStatusStr)
			if err != nil {
				return nil, err
			}

			subscriberGameStatus = models.SubscriberGameStatus(subscriberGameStatusInt)
		}

		username := r[roomSubscriberUsernameField]

		roomSubscribers = append(roomSubscribers, models.RoomSubscriber{
			UserId:     roomSubscriberId,
			Status:     status,
			GameStatus: subscriberGameStatus,
			Username:   username,
		})
	}

	return roomSubscribers, nil
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

package redis_repo

import (
	"10-typing/models"
	"context"
	"strconv"

	"github.com/google/uuid"
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

func (repo *RedisRepository) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	if err != nil {
		return models.InactiveSubscriberStatus, err
	}

	return models.SubscriberStatus(status), nil
}

func (repo *RedisRepository) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberGameStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberGameStatusField).Int()
	if err != nil {
		return models.UnstartedSubscriberGameStatus, err
	}

	return models.SubscriberGameStatus(status), nil
}

func (repo *RedisRepository) GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]models.RoomSubscriber, error) {
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)

	r, err := repo.redisClient.SMembers(ctx, roomSubscriberIdsKey).Result()
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

		r, err := repo.redisClient.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		status := models.InactiveSubscriberStatus
		statusStr, ok := r[roomSubscriberStatusField]
		if ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err != nil {
				return nil, err
			}
			status = models.SubscriberStatus(statusInt)
		}

		subscriberGameStatus := models.UnstartedSubscriberGameStatus
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

func (repo *RedisRepository) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return repo.redisClient.SAdd(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (repo *RedisRepository) SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err()
}

func (repo *RedisRepository) SetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err()
}

func (repo *RedisRepository) DeleteRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	err := repo.redisClient.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		return err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	return repo.redisClient.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err()
}

func (repo *RedisRepository) DeleteRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return repo.redisClient.SRem(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

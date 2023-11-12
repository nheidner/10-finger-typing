package redis_repo

import (
	"10-typing/models"
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	roomSubscriberStatusField     = "status"
	roomSubscriberGameStatusField = "game_status"
	roomSubscriberUsernameField   = "username"
	connectionExpirationMilli     = 1000 * 10
)

// rooms:[room_id]:subscribers_ids set: user ids
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

// rooms:[room_id]:subscribers:[user_id] hash: status, username, game_status
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

func getRoomSubscriberConnectionKey(roomId, userId uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userId) + ":conns"
}

func (repo *RedisRepository) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (status models.SubscriberStatus, roomSubscriberStatusHasBeenUpdated bool, err error) {
	numberRoomSubscriberConns, err := repo.getNumberRoomSubscriberConnections(ctx, roomId, userId)
	if err != nil {
		return models.InactiveSubscriberStatus, false, err
	}

	status, err = repo.getRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return models.InactiveSubscriberStatus, false, err
	}

	// have to check all possibilities because deleted connection might have been expired
	switch {
	case numberRoomSubscriberConns > 0 && status == models.ActiveSubscriberStatus:
		return models.ActiveSubscriberStatus, false, nil
	case numberRoomSubscriberConns == 0 && status == models.InactiveSubscriberStatus:
		return models.InactiveSubscriberStatus, false, nil
	case numberRoomSubscriberConns > 0 && status == models.InactiveSubscriberStatus:
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.ActiveSubscriberStatus); err != nil {
			return models.InactiveSubscriberStatus, false, err
		}

		return models.ActiveSubscriberStatus, true, nil
	default: // numberRoomSubscriberConns == 0 && status == models.ActiveSubscriberStatus
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.InactiveSubscriberStatus); err != nil {
			return models.InactiveSubscriberStatus, false, err
		}

		return models.InactiveSubscriberStatus, true, nil
	}
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

func (repo *RedisRepository) SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err()
}

func (repo *RedisRepository) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, newConnectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	status, err := repo.getRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return false, err
	}

	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)
	expirationTime := time.Now().Add(connectionExpirationMilli * time.Millisecond).UnixMilli()

	if err = repo.redisClient.ZAdd(ctx, roomSubscriberConnectionKey, redis.Z{
		Score:  float64(expirationTime),
		Member: newConnectionId.String(),
	}).Err(); err != nil {
		return false, err
	}

	if status == models.InactiveSubscriberStatus {
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.ActiveSubscriberStatus); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
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

func (repo *RedisRepository) DeleteRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)

	if err = repo.redisClient.ZRem(ctx, roomSubscriberConnectionKey, connectionId.String()).Err(); err != nil {
		return false, err
	}

	numberRoomSubscriberConns, err := repo.getNumberRoomSubscriberConnections(ctx, roomId, userId)
	if err != nil {
		return false, err
	}

	status, err := repo.getRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return false, err
	}

	if numberRoomSubscriberConns == 0 && status == models.ActiveSubscriberStatus {
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.InactiveSubscriberStatus); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (repo *RedisRepository) getRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	if err != nil {
		return models.InactiveSubscriberStatus, err
	}

	return models.SubscriberStatus(status), nil
}

func (repo *RedisRepository) setRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err()
}

func (repo *RedisRepository) getNumberRoomSubscriberConnections(ctx context.Context, roomId, userId uuid.UUID) (int64, error) {
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)

	nowMilli := time.Now().UnixMilli()
	nowMilliStr := strconv.Itoa(int(nowMilli))

	if err := repo.redisClient.ZRangeByScore(ctx, roomSubscriberConnectionKey, &redis.ZRangeBy{Min: "0", Max: nowMilliStr}).Err(); err != nil {
		return 0, err
	}

	return repo.redisClient.ZCard(ctx, roomSubscriberConnectionKey).Result()
}

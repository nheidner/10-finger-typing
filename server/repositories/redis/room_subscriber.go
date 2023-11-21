package redis_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
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
	connectionExpirationMilli     = 1000 * 60 * 10
)

// rooms:[room_id]:subscribers_ids set: user ids
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

// rooms:[room_id]:subscribers:[user_id] hash: status, username, game_status
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

// rooms:[room_id]:subscribers:[user_id]:conns
func getRoomSubscriberConnectionKey(roomId, userId uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userId) + ":conns"
}

func (repo *RedisRepository) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (numberRoomSubscriberConns int64, roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.GetRoomSubscriberStatus"

	numberRoomSubscriberConns, err = repo.getNumberRoomSubscriberConnections(ctx, roomId, userId)
	if err != nil {
		return 0, false, errors.E(op, err)
	}

	status, err := repo.getRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return 0, false, errors.E(op, err)
	}

	if numberRoomSubscriberConns == 0 && status == models.ActiveSubscriberStatus {
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.InactiveSubscriberStatus); err != nil {
			return 0, false, errors.E(op, err)
		}

		return numberRoomSubscriberConns, true, nil
	}

	return numberRoomSubscriberConns, false, nil
}

func (repo *RedisRepository) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberGameStatus, error) {
	const op errors.Op = "redis_repo.GetRoomSubscriberGameStatus"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberGameStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.UnstartedSubscriberGameStatus, errors.E(op, repositories.ErrNotFound)
	case err != nil:
		return models.UnstartedSubscriberGameStatus, err
	}

	return models.SubscriberGameStatus(status), nil
}

func (repo *RedisRepository) GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]models.RoomSubscriber, error) {
	const op errors.Op = "redis_repo.GetRoomSubscribers"
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)

	r, err := repo.redisClient.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	roomSubscribers := make([]models.RoomSubscriber, 0, len(r))
	for _, roomSubscriberIdStr := range r {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		r, err := repo.redisClient.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, errors.E(op, err)
		}

		status := models.InactiveSubscriberStatus
		statusStr, ok := r[roomSubscriberStatusField]
		if ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err != nil {
				return nil, errors.E(op, err)
			}
			status = models.SubscriberStatus(statusInt)
		}

		subscriberGameStatus := models.UnstartedSubscriberGameStatus
		subscriberGameStatusStr, ok := r[roomSubscriberGameStatusField]
		if ok {
			subscriberGameStatusInt, err := strconv.Atoi(subscriberGameStatusStr)
			if err != nil {
				return nil, errors.E(op, err)
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
	const op errors.Op = "redis_repo.SetRoomSubscriberGameStatus"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	if err := repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// adds room subscriber connection to rooms:[room_id]:subscribers:[user_id]:conns and adapts room subscribe status to active if necessary
func (repo *RedisRepository) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, newConnectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.SetRoomSubscriberConnection"

	status, err := repo.getRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return false, errors.E(op, err)
	}

	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)
	expirationTime := time.Now().Add(connectionExpirationMilli * time.Millisecond).UnixMilli()

	if err = repo.redisClient.ZAdd(ctx, roomSubscriberConnectionKey, redis.Z{
		Score:  float64(expirationTime),
		Member: newConnectionId.String(),
	}).Err(); err != nil {
		return false, errors.E(op, err)
	}

	if status == models.InactiveSubscriberStatus {
		if err = repo.setRoomSubscriberStatus(ctx, roomId, userId, models.ActiveSubscriberStatus); err != nil {
			return false, errors.E(op, err)
		}

		return true, nil
	}

	return false, nil
}

func (repo *RedisRepository) DeleteRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.DeleteRoomSubscriber"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	err := repo.redisClient.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		return errors.E(op, err)
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	if err := repo.redisClient.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.DeleteRoomSubscriberConnection"
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)

	if err = repo.redisClient.ZRem(ctx, roomSubscriberConnectionKey, connectionId.String()).Err(); err != nil {
		return false, errors.E(op, err)
	}

	_, roomSubscriberStatusHasBeenUpdated, err = repo.GetRoomSubscriberStatus(ctx, roomId, userId)
	if err != nil {
		return false, errors.E(op, err)
	}

	return roomSubscriberStatusHasBeenUpdated, nil
}

func (repo *RedisRepository) getRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberStatus, error) {
	const op errors.Op = "redis_repo.getRoomSubscriberStatus"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.InactiveSubscriberStatus, errors.E(op, repositories.ErrNotFound)
	case err != nil:
		return models.InactiveSubscriberStatus, errors.E(op, err)
	}

	return models.SubscriberStatus(status), nil
}

func (repo *RedisRepository) setRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberStatus) error {
	const op errors.Op = "redis_repo.setRoomSubscriberStatus"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	if err := repo.redisClient.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) getNumberRoomSubscriberConnections(ctx context.Context, roomId, userId uuid.UUID) (numberRoomSubscriberConnections int64, err error) {
	const op errors.Op = "redis_repo.getNumberRoomSubscriberConnections"
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)

	nowMilli := time.Now().UnixMilli()
	nowMilliStr := strconv.Itoa(int(nowMilli))

	if err := repo.redisClient.ZRemRangeByScore(ctx, roomSubscriberConnectionKey, "0", nowMilliStr).Err(); err != nil {
		return 0, errors.E(op, err)
	}

	numberRoomSubscriberConnections, err = repo.redisClient.ZCard(ctx, roomSubscriberConnectionKey).Result()
	if err != nil {
		return 0, errors.E(op, err)
	}

	return numberRoomSubscriberConnections, nil
}

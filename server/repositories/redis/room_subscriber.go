package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"10-typing/utils"
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const connectionExpirationMilli = 1000 * 60 * 10

// GetRoomSubscriberStatus queries the current number of connections from the key rooms:[room_id]:subscribers:[user_id]:conns,
// queries the room subscriber's (rooms:[room_id]:subscribers:[user_id]) status and sets the status to inactive if there were no room connections returned.
// It does this by using a transaction with WATCH and starts the whole transaction again after the previous transaction was discarded
// when f.e. one of the watched keys' (rooms:[room_id]:subscribers:[user_id], rooms:[room_id]:subscribers:[user_id]:conns) values changed.
func (repo *RedisRepository) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (numberRoomSubscriberConns int64, roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.RedisRepository.GetRoomSubscriberStatus"
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	for i := 0; i < retries; i++ {
		// watches roomSubscriber key and roomSubscriberConnection key and conditionally sets new room subscriber status
		err = repo.redisClient.Watch(ctx, func(_ *redis.Tx) error {
			numberRoomSubscriberConns, err = repo.getNumberRoomSubscriberConnections(ctx, roomSubscriberConnectionKey, roomId, userId)
			if err != nil {
				return errors.E(op, err)
			}

			status, err := repo.getRoomSubscriberStatus(ctx, roomSubscriberKey, roomId, userId)
			if err != nil {
				return errors.E(op, err)
			}

			tx := repo.BeginTx()
			if numberRoomSubscriberConns == 0 && status == models.ActiveSubscriberStatus {
				if err = repo.setRoomSubscriberStatus(ctx, tx, roomId, userId, models.InactiveSubscriberStatus); err != nil {
					err := errors.E(op, err)
					return utils.RollbackAndErr(op, err, tx)
				}

				roomSubscriberStatusHasBeenUpdated = true
			}

			return tx.Commit(ctx)
		}, roomSubscriberKey, roomSubscriberConnectionKey)
		switch err {
		case redis.TxFailedErr:
			roomSubscriberStatusHasBeenUpdated = false
			continue
		case nil:
			return numberRoomSubscriberConns, roomSubscriberStatusHasBeenUpdated, nil
		default:
			return 0, false, errors.E(op, err)
		}
	}

	return numberRoomSubscriberConns, false, nil
}

// func (repo *RedisRepository) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberGameStatus, error) {
// 	const op errors.Op = "redis_repo.RedisRepository.GetRoomSubscriberGameStatus"
// 	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

// 	status, err := repo.redisClient.HGet(ctx, roomSubscriberKey, roomSubscriberGameStatusField).Int()
// 	switch {
// 	case err == redis.Nil:
// 		return models.UnstartedSubscriberGameStatus, errors.E(op, common.ErrNotFound)
// 	case err != nil:
// 		return models.UnstartedSubscriberGameStatus, err
// 	}

// 	return models.SubscriberGameStatus(status), nil
// }

func (repo *RedisRepository) GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]models.RoomSubscriber, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetRoomSubscribers"
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

func (repo *RedisRepository) SetRoomSubscriberGameStatus(ctx context.Context, tx common.Transaction, roomId, userId uuid.UUID, status models.SubscriberGameStatus) error {
	const op errors.Op = "redis_repo.RedisRepository.SetRoomSubscriberGameStatus"
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	cmd := repo.cmdable(tx)

	if err := cmd.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// SetRoomSubscriberConnection adds a new room subscriber connection to rooms:[room_id]:subscribers:[user_id]:conns key
// and sets room subscriber's key (rooms:[room_id]:subscribers:[user_id]) status field value to active if it was previously set to inactive.
// It does this by using a transaction with WATCH and starts the whole transaction again after the previous transaction was discarded
// when f.e. one of the watched keys' (rooms:[room_id]:subscribers:[user_id], rooms:[room_id]:subscribers:[user_id]:conns) values changed.
func (repo *RedisRepository) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, newConnectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.RedisRepository.SetRoomSubscriberConnection"

	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	roomSubscriberConnectionKey := getRoomSubscriberConnectionKey(roomId, userId)

	for i := 0; i < retries; i++ {
		err := repo.redisClient.Watch(ctx, func(_ *redis.Tx) error {
			status, err := repo.getRoomSubscriberStatus(ctx, roomSubscriberKey, roomId, userId)
			if err != nil {
				return err
			}

			tx := repo.BeginTx()
			cmd := repo.cmdable(tx)
			expirationTime := time.Now().Add(connectionExpirationMilli * time.Millisecond).UnixMilli()
			if err = cmd.ZAdd(ctx, roomSubscriberConnectionKey, redis.Z{
				Score:  float64(expirationTime),
				Member: newConnectionId.String(),
			}).Err(); err != nil {
				return err
			}

			if status == models.InactiveSubscriberStatus {
				if err = repo.setRoomSubscriberStatus(ctx, tx, roomId, userId, models.ActiveSubscriberStatus); err != nil {
					return err
				}

				roomSubscriberStatusHasBeenUpdated = true
			}

			return tx.Commit(ctx)
		}, roomSubscriberKey, roomSubscriberConnectionKey)
		switch err {
		case redis.TxFailedErr:
			roomSubscriberStatusHasBeenUpdated = false
			continue
		case nil:
			return roomSubscriberStatusHasBeenUpdated, nil
		default:
			return false, errors.E(op, err)
		}
	}

	return roomSubscriberStatusHasBeenUpdated, nil
}

// DeleteRoomSubscriber deletes the rooms:[room_id]:subscribers:[user_id] key and the user id from the rooms:[room_id]:subscribers_ids key's set value.
// It does so by using a MULTI/EXEC transaction.
func (repo *RedisRepository) DeleteRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteRoomSubscriber"
	var cmd = repo.redisClient

	tx := repo.BeginTx()

	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	err := cmd.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	if err := cmd.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err(); err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error) {
	const op errors.Op = "redis_repo.RedisRepository.DeleteRoomSubscriberConnection"
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

func (repo *RedisRepository) getRoomSubscriberStatus(ctx context.Context, roomSubscriberKey string, roomId, userId uuid.UUID) (models.SubscriberStatus, error) {
	const op errors.Op = "redis_repo.RedisRepository.getRoomSubscriberStatus"
	var cmd = repo.redisClient

	status, err := cmd.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.InactiveSubscriberStatus, errors.E(op, common.ErrNotFound)
	case err != nil:
		return models.InactiveSubscriberStatus, errors.E(op, err)
	}

	return models.SubscriberStatus(status), nil
}

func (repo *RedisRepository) setRoomSubscriberStatus(ctx context.Context, tx common.Transaction, roomId, userId uuid.UUID, status models.SubscriberStatus) error {
	const op errors.Op = "redis_repo.RedisRepository.setRoomSubscriberStatus"
	cmd := repo.cmdable(tx)
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	if err := cmd.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// getNumberRoomSubscriberConnections deletes expired connections from the roomSubscriberConnection key's sorted set and returns the number of connections remaining for this room subscriber.
func (repo *RedisRepository) getNumberRoomSubscriberConnections(ctx context.Context, roomSubscriberConnectionKey string, roomId, userId uuid.UUID) (numberRoomSubscriberConnections int64, err error) {
	const op errors.Op = "redis_repo.RedisRepository.getNumberRoomSubscriberConnections"
	var cmd = repo.redisClient

	nowMilli := time.Now().UnixMilli()
	nowMilliStr := strconv.Itoa(int(nowMilli))

	if err := cmd.ZRemRangeByScore(ctx, roomSubscriberConnectionKey, "0", nowMilliStr).Err(); err != nil {
		return 0, errors.E(op, err)
	}

	numberRoomSubscriberConnections, err = cmd.ZCard(ctx, roomSubscriberConnectionKey).Result()
	if err != nil {
		return 0, errors.E(op, err)
	}

	return numberRoomSubscriberConnections, nil
}

// MULTIOPERATIONS
func (repo *RedisRepository) SetRoomSubscriberGameStatusForAllRoomSubscribers(ctx context.Context, roomId uuid.UUID, newSubscriberGameStatus models.SubscriberGameStatus) error {
	const op errors.Op = "redis_repo.RedisRepository.SetGameStatusForAllRoomSubscribers"

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)

	for i := 0; i < retries; i++ {
		err := repo.redisClient.Watch(ctx, func(_ *redis.Tx) error {
			currentGameUserIds, err := repo.getRoomSubscribersIds(ctx, roomId)
			if err != nil {
				return err
			}

			tx := repo.BeginTx()

			for _, currentGameUserId := range currentGameUserIds {
				if err := repo.SetRoomSubscriberGameStatus(ctx, tx, roomId, currentGameUserId, newSubscriberGameStatus); err != nil {
					return err
				}
			}

			return tx.Commit(ctx)
		}, roomSubscriberIdsKey)
		switch err {
		case redis.TxFailedErr:
			continue
		case nil:
			return nil
		default:
			return errors.E(op, err)
		}
	}

	return nil
}

func (repo *RedisRepository) getRoomSubscribersIds(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetRoomSubscribersIds"
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	var cmd = repo.redisClient

	r, err := cmd.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	roomSubscriberIds, err := stringsToUuids(r)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return roomSubscriberIds, nil
}

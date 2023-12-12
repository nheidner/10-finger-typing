package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"

	"10-typing/utils"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (repo *RedisRepository) GetRoomInCacheOrDb(ctx context.Context, dbRepo common.DBRepository, roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetRoomInCacheOrDb"

	room, err := repo.getRoom(ctx, roomId)
	switch {
	case err != nil && errors.Is(err, common.ErrNotFound):
		room, err = dbRepo.FindRoomWithUsers(ctx, nil, roomId)
		if err != nil {
			return nil, errors.E(op, err)
		}

		if err = repo.SetRoom(ctx, nil, *room); err != nil {
			// no error should be returned
			log.Print(errors.E(op, err))
		}
	case err != nil:
		return nil, errors.E(op, err)
	}

	return room, nil
}

func (repo *RedisRepository) GetRoomGameDurationSec(ctx context.Context, roomId uuid.UUID) (gameDurationSec int, err error) {
	const op errors.Op = "redis_repo.RedisRepository.GetRoomGameDurationSec"
	var roomKey = getRoomKey(roomId)
	var cmd redis.Cmdable = repo.redisClient

	gameDurationSec, err = cmd.HGet(ctx, roomKey, roomGameDurationSecField).Int()
	switch {
	case err == redis.Nil:
		return 0, errors.E(op, common.ErrNotFound)
	case err != nil:
		return 0, errors.E(op, err)
	}

	return gameDurationSec, nil
}

func (repo *RedisRepository) SetRoom(ctx context.Context, tx common.Transaction, room models.Room) error {
	const op errors.Op = "redis_repo.RedisRepository.SetRoom"
	var roomKey = getRoomKey(room.ID)

	// PIPELINE start if no outer pipeline exists
	cmd, innerTx := repo.beginPipelineIfNoOuterTransactionExists(tx)

	// add room
	roomValue := map[string]any{
		roomAdminIdField:         room.AdminId.String(),
		roomCreatedAtField:       room.CreatedAt.UnixMilli(),
		roomUpdatedAtField:       room.UpdatedAt.UnixMilli(),
		roomGameDurationSecField: room.GameDurationSec,
	}
	cmd.HSet(ctx, roomKey, roomValue)

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(room.ID)
	roomSubscriberIdsValue := make([]string, 0, len(room.Users))
	for _, subscriber := range room.Users {
		roomSubscriberIdsValue = append(roomSubscriberIdsValue, subscriber.ID.String())

		// add room subscribers
		roomSubscriberKey := getRoomSubscriberKey(room.ID, subscriber.ID)
		roomSubscriberValue := map[string]any{
			roomSubscriberUsernameField:   subscriber.Username,
			roomSubscriberStatusField:     strconv.Itoa(int(models.InactiveSubscriberStatus)),
			roomSubscriberGameStatusField: strconv.Itoa(int(models.InactiveSubscriberStatus)),
		}

		cmd.HSet(ctx, roomSubscriberKey, roomSubscriberValue)
	}

	// add room subscriber ids
	cmd.SAdd(ctx, roomSubscriberIdsKey, roomSubscriberIdsValue)

	// PIPELINE commit
	if innerTx != nil {
		if err := innerTx.Commit(ctx); err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (repo *RedisRepository) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.RoomHasAdmin"
	var roomKey = getRoomKey(roomId)
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.HGet(ctx, roomKey, roomAdminIdField).Result()
	switch {
	case err == redis.Nil:
		return false, errors.E(op, common.ErrNotFound)
	case err != nil:
		return false, err
	}

	return r == adminId.String(), nil
}

func (repo *RedisRepository) RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.RoomHasSubscribers"
	var roomSubscriberIdsKey = getRoomSubscriberIdsKey(roomId)
	var tempUserIdsKey = "temp:" + uuid.New().String()
	var cmd redis.Cmdable = repo.redisClient

	if len(userIds) == 0 {
		err := fmt.Errorf("at least one user id must be specified")
		return false, errors.E(op, err)
	}

	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}

	if err := cmd.SAdd(ctx, tempUserIdsKey, userIdStrs...).Err(); err != nil {
		return false, errors.E(op, err)
	}

	r, err := cmd.SInter(ctx, roomSubscriberIdsKey, tempUserIdsKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	if err := cmd.Del(ctx, tempUserIdsKey).Err(); err != nil {
		return false, errors.E(op, err)
	}

	return len(r) == len(userIds), nil
}

func (repo *RedisRepository) RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.RoomExists"
	var roomKey = getRoomKey(roomId)
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.Exists(ctx, roomKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r > 0, nil
}

func (repo *RedisRepository) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteRoom"
	var roomKey = getRoomKey(roomId)
	var pattern = roomKey + "*"

	if err := deleteKeysByPattern(ctx, repo, pattern); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllRooms(ctx context.Context) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteAllRooms"
	var pattern = "rooms:*"

	if err := deleteKeysByPattern(ctx, repo, pattern); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) getRoom(ctx context.Context, roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "redis_repo.RedisRepository.getRoom"
	var roomKey = getRoomKey(roomId)
	var cmd redis.Cmdable = repo.redisClient

	roomData, err := cmd.HGetAll(ctx, roomKey).Result()
	switch {
	case err != nil:
		return nil, err
	case len(roomData) == 0:
		return nil, errors.E(op, common.ErrNotFound)
	}

	createdAt, err := utils.StringToTime(roomData[roomCreatedAtField])
	if err != nil {
		return nil, errors.E(op, err)
	}
	updatedAt, err := utils.StringToTime(roomData[roomUpdatedAtField])
	if err != nil {
		return nil, errors.E(op, err)
	}
	adminId, err := uuid.Parse(roomData[roomAdminIdField])
	if err != nil {
		return nil, errors.E(op, err)
	}
	gameDurationSec, err := strconv.Atoi(roomData[roomGameDurationSecField])
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &models.Room{
		ID:              roomId,
		AdminId:         adminId,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		GameDurationSec: gameDurationSec,
	}, nil
}

package redis_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	roomAdminIdField         = "admin_id"
	roomCreatedAtField       = "created_at"
	roomUpdatedAtField       = "updated_at"
	roomGameDurationSecField = "game_duration"
)

// rooms:[room_id] hash: roomAdminId, createdAt, updatedAt
func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

func (repo *RedisRepository) GetRoom(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "redis_repo.GetRoom"
	roomKey := getRoomKey(roomId)

	roomData, err := repo.redisClient.HGetAll(ctx, roomKey).Result()
	switch {
	case err != nil:
		return nil, err
	case len(roomData) == 0:
		return nil, errors.E(op, repositories.ErrNotFound)
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	roomSubscriberIds, err := repo.redisClient.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	userIdStr := userId.String()
	if !utils.SliceContains[string](roomSubscriberIds, userIdStr) {
		err := fmt.Errorf("user with user id %s is not subscribed to room with room id %s", userIdStr, roomId.String())
		return nil, errors.E(op, err)
	}

	roomSubscribers := make([]models.RoomSubscriber, 0, len(roomSubscriberIds))
	for _, roomSubscriberIdStr := range roomSubscriberIds {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		roomSubscriber, err := repo.redisClient.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, errors.E(op, err)
		}

		username := roomSubscriber[roomSubscriberUsernameField]

		subscriber := models.RoomSubscriber{
			UserId:   roomSubscriberId,
			Username: username,
		}

		roomSubscribers = append(roomSubscribers, subscriber)
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
		Subscribers:     roomSubscribers,
		GameDurationSec: gameDurationSec,
	}, nil
}

func (repo *RedisRepository) GetRoomGameDurationSec(ctx context.Context, roomId uuid.UUID) (gameDurationSec int, err error) {
	const op errors.Op = "redis_repo.GetRoomGameDurationSec"
	roomKey := getRoomKey(roomId)

	gameDurationSec, err = repo.redisClient.HGet(ctx, roomKey, roomGameDurationSecField).Int()
	switch {
	case err == redis.Nil:
		return 0, errors.E(op, repositories.ErrNotFound)
	case err != nil:
		return 0, errors.E(op, err)
	}

	return gameDurationSec, nil
}

func (repo *RedisRepository) SetRoom(ctx context.Context, room models.Room) error {
	const op errors.Op = "redis_repo.SetRoom"
	// add room
	roomKey := getRoomKey(room.ID)
	roomValue := map[string]any{
		roomAdminIdField:         room.AdminId.String(),
		roomCreatedAtField:       room.CreatedAt.UnixMilli(),
		roomUpdatedAtField:       room.UpdatedAt.UnixMilli(),
		roomGameDurationSecField: room.GameDurationSec,
	}
	if err := repo.redisClient.HSet(ctx, roomKey, roomValue).Err(); err != nil {
		return errors.E(op, err)
	}

	// add room subscriber ids
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(room.ID)
	roomSubscriberIdsValue := make([]string, 0, len(room.Users))
	for _, subscriber := range room.Users {
		roomSubscriberIdsValue = append(roomSubscriberIdsValue, subscriber.ID.String())
	}

	if err := repo.redisClient.SAdd(ctx, roomSubscriberIdsKey, roomSubscriberIdsValue).Err(); err != nil {
		return errors.E(op, err)
	}

	// add room subscribers
	for _, subscriber := range room.Users {
		roomSubscriberKey := getRoomSubscriberKey(room.ID, subscriber.ID)
		roomSubscriberValue := map[string]any{
			roomSubscriberUsernameField:   subscriber.Username,
			roomSubscriberStatusField:     strconv.Itoa(int(models.InactiveSubscriberStatus)),
			roomSubscriberGameStatusField: strconv.Itoa(int(models.InactiveSubscriberStatus)),
		}

		if err := repo.redisClient.HSet(ctx, roomSubscriberKey, roomSubscriberValue).Err(); err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (repo *RedisRepository) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RoomHasAdmin"
	roomKey := getRoomKey(roomId)

	r, err := repo.redisClient.HGet(ctx, roomKey, roomAdminIdField).Result()
	switch {
	case err == redis.Nil:
		return false, errors.E(op, repositories.ErrNotFound)
	case err != nil:
		return false, err
	}

	return r == adminId.String(), nil
}

func (repo *RedisRepository) RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RoomHasSubscribers"

	if len(userIds) == 0 {
		err := fmt.Errorf("at least one user id must be specified")
		return false, errors.E(op, err)
	}

	roomSubscriberIds := getRoomSubscriberIdsKey(roomId)
	tempUserIdsKey := "temp:" + uuid.New().String()

	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}

	if err := repo.redisClient.SAdd(ctx, tempUserIdsKey, userIdStrs...).Err(); err != nil {
		return false, errors.E(op, err)
	}

	r, err := repo.redisClient.SInter(ctx, roomSubscriberIds, tempUserIdsKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	if err := repo.redisClient.Del(ctx, tempUserIdsKey).Err(); err != nil {
		return false, errors.E(op, err)
	}

	return len(r) == len(userIds), nil
}

func (repo *RedisRepository) RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RoomExists"
	roomKey := getRoomKey(roomId)

	r, err := repo.redisClient.Exists(ctx, roomKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r > 0, nil
}

func (repo *RedisRepository) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "redis_repo.DeleteRoom"
	roomKey := getRoomKey(roomId)
	pattern := roomKey + "*"

	if err := deleteKeysByPattern(ctx, repo, pattern); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllRooms(ctx context.Context) error {
	const op errors.Op = "redis_repo.DeleteAllRooms"
	pattern := "rooms:*"

	if err := deleteKeysByPattern(ctx, repo, pattern); err != nil {
		return errors.E(op, err)
	}

	return nil
}

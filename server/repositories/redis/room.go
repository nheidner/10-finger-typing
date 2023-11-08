package redis_repo

import (
	"10-typing/models"
	"10-typing/utils"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"
)

const (
	roomAdminIdField   = "admin_id"
	roomCreatedAtField = "created_at"
	roomUpdatedAtField = "updated_at"
)

// rooms:[room_id] hash: roomAdminId, createdAt, updatedAt
func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

func (repo *RedisRepository) GetRoom(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	roomKey := getRoomKey(roomId)

	roomData, err := repo.redisClient.HGetAll(ctx, roomKey).Result()
	switch {
	case err != nil:
		return nil, err
	case len(roomData) == 0:
		return nil, nil
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	roomSubscriberIds, err := repo.redisClient.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	userIdStr := userId.String()
	if !utils.SliceContains[string](roomSubscriberIds, userIdStr) {
		return nil, fmt.Errorf("user is not subscribed to room")
	}

	roomSubscribers := make([]models.RoomSubscriber, 0, len(roomSubscriberIds))
	for _, roomSubscriberIdStr := range roomSubscriberIds {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, err
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		roomSubscriber, err := repo.redisClient.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
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
		return nil, err
	}
	updatedAt, err := utils.StringToTime(roomData[roomUpdatedAtField])
	if err != nil {
		return nil, err
	}
	adminId, err := uuid.Parse(roomData[roomAdminIdField])
	if err != nil {
		return nil, err
	}

	return &models.Room{
		ID:          roomId,
		AdminId:     adminId,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Subscribers: roomSubscribers,
	}, nil
}

func (repo *RedisRepository) SetRoom(ctx context.Context, room models.Room) error {
	// add room
	roomKey := getRoomKey(room.ID)
	roomValue := map[string]any{
		roomAdminIdField:   room.AdminId.String(),
		roomCreatedAtField: room.CreatedAt.UnixMilli(),
		roomUpdatedAtField: room.UpdatedAt.UnixMilli(),
	}
	if err := repo.redisClient.HSet(ctx, roomKey, roomValue).Err(); err != nil {
		return err
	}

	// add room subscriber ids
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(room.ID)
	roomSubscriberIdsValue := make([]string, 0, len(room.Users))
	for _, subscriber := range room.Users {
		roomSubscriberIdsValue = append(roomSubscriberIdsValue, subscriber.ID.String())
	}

	if err := repo.redisClient.SAdd(ctx, roomSubscriberIdsKey, roomSubscriberIdsValue).Err(); err != nil {
		return err
	}

	// add room subscribers
	for _, subscriber := range room.Users {
		roomSubscriberKey := getRoomSubscriberKey(room.ID, subscriber.ID)
		roomSubscriberValue := map[string]any{
			roomSubscriberUsernameField:   subscriber.Username,
			roomSubscriberStatusField:     strconv.Itoa(int(models.NilSubscriberStatus)),
			roomSubscriberGameStatusField: strconv.Itoa(int(models.NilSubscriberGameStatus)),
		}

		if err := repo.redisClient.HSet(ctx, roomSubscriberKey, roomSubscriberValue).Err(); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (repo *RedisRepository) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := repo.redisClient.HGet(ctx, roomKey, roomAdminIdField).Result()
	if err != nil {
		return false, err
	}

	return r == adminId.String(), nil
}

func (repo *RedisRepository) RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error) {
	if len(userIds) == 0 {
		return false, fmt.Errorf("at least one user id must be specified")
	}

	roomSubscriberIds := getRoomSubscriberIdsKey(roomId)
	tempUserIdsKey := "temp:" + uuid.New().String()

	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}

	if err := repo.redisClient.SAdd(ctx, tempUserIdsKey, userIdStrs...).Err(); err != nil {
		return false, err
	}

	r, err := repo.redisClient.SInter(ctx, roomSubscriberIds, tempUserIdsKey).Result()
	if err != nil {
		return false, err
	}

	if err := repo.redisClient.Del(ctx, tempUserIdsKey).Err(); err != nil {
		return false, err
	}

	return len(r) == len(userIds), nil
}

func (repo *RedisRepository) RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := repo.redisClient.Exists(ctx, roomKey).Result()
	if err != nil {
		return false, err
	}

	return r > 0, nil
}

func (repo *RedisRepository) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	roomKey := getRoomKey(roomId)

	pattern := roomKey + "*"
	iter := repo.redisClient.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		repo.redisClient.Del(ctx, key)
	}

	return iter.Err()
}

func (repo *RedisRepository) DeleteAllRooms(ctx context.Context) error {
	pattern := "rooms:*"

	iter := repo.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		repo.redisClient.Del(ctx, key)

		return iter.Err()
	}

	return nil
}

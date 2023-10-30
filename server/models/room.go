package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RoomService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func (rs *RoomService) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := rs.RDB.HGet(ctx, roomKey, roomAdminIdField).Result()
	if err != nil {
		return false, err
	}

	return r == adminId.String(), nil
}

func (rs *RoomService) RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error) {
	if len(userIds) == 0 {
		return false, fmt.Errorf("at least one user id must be specified")
	}

	roomSubscriberIds := getRoomSubscriberIdsKey(roomId)
	tempUserIdsKey := "temp:" + uuid.New().String()

	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}

	if err := rs.RDB.SAdd(ctx, tempUserIdsKey, userIdStrs...).Err(); err != nil {
		return false, err
	}

	r, err := rs.RDB.SInter(ctx, roomSubscriberIds, tempUserIdsKey).Result()
	if err != nil {
		return false, err
	}

	if err := rs.RDB.Del(ctx, tempUserIdsKey).Err(); err != nil {
		return false, err
	}

	return len(r) == len(userIds), nil
}

func (rs *RoomService) RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := rs.RDB.Exists(ctx, roomKey).Result()
	if err != nil {
		return false, err
	}

	return r > 0, nil
}

func (rs *RoomService) PublishMessage(ctx context.Context, roomId uuid.UUID, msg WSMessage) error {
	roomStreamKey := getRoomStreamKey(roomId)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return rs.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{"data": data},
	}).Err()
}

func (rs *RoomService) PublishAction(ctx context.Context, roomId uuid.UUID, action StreamAction) error {
	roomStreamKey := getRoomStreamKey(roomId)

	return rs.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]string{"action": strconv.Itoa(int(action))},
	}).Err()
}

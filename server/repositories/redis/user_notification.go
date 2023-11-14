package redis_repo

import (
	"10-typing/models"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	userNotificationStreamMaxlen = 10
)

func getUserNotificationStreamKey(userId uuid.UUID) string {
	return "users:" + userId.String() + ":notifications"
}

func (repo *RedisRepository) PublishUserNotification(ctx context.Context, userId uuid.UUID, userNotification models.UserNotification) error {
	userNotificationStreamKey := getUserNotificationStreamKey(userId)
	userNotificationData, err := json.Marshal(userNotification)
	if err != nil {
		return err
	}

	return repo.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: userNotificationStreamKey,
		MaxLen: userNotificationStreamMaxlen,
		Values: map[string]any{
			streamEntryMessageField: userNotificationData,
		},
	}).Err()
}

func (repo *RedisRepository) GetUserNotification(ctx context.Context, userId uuid.UUID, startId string) (*models.UserNotification, error) {
	userNotificationStreamKey := getUserNotificationStreamKey(userId)

	id := "$"
	if startId != "" {
		id = startId
	}

	r, err := repo.redisClient.XRead(ctx, &redis.XReadArgs{
		Streams: []string{userNotificationStreamKey, id},
		Count:   1,
		Block:   0,
	}).Result()
	if err != nil {
		return nil, err
	}

	userNotificationId := r[0].Messages[0].ID
	message, ok := r[0].Messages[0].Values[streamEntryMessageField]
	if !ok {
		return nil, errors.New("no " + streamEntryMessageField + " field in stream entry")
	}
	messageStr, ok := message.(string)
	if !ok {
		return nil, errors.New("underlying type of " + streamEntryMessageField + " stream entry field is not string")
	}

	var userNotification models.UserNotification
	if err := json.Unmarshal([]byte(messageStr), &userNotification); err != nil {
		return nil, err
	}

	userNotification.Id = userNotificationId

	return &userNotification, nil
}

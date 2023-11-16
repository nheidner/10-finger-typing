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

func (repo *RedisRepository) GetUserNotification(ctx context.Context, userId uuid.UUID, startId string) chan models.StreamSubscriptionResult[*models.UserNotification] {
	userNotificationStreamKey := getUserNotificationStreamKey(userId)

	return getStreamEntry[*models.UserNotification](ctx, repo, userNotificationStreamKey, startId, func(values map[string]interface{}, entryId string) (*models.UserNotification, error) {
		message, ok := values[streamEntryMessageField]
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

		userNotification.Id = entryId

		return &userNotification, nil
	})
}

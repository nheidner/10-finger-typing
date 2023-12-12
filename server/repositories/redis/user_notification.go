package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	userNotificationStreamMaxlen = 10
)

func (repo *RedisRepository) PublishUserNotification(ctx context.Context, tx common.Transaction, userId uuid.UUID, userNotification models.UserNotification) error {
	const op errors.Op = "redis_repo.RedisRepository.PublishUserNotification"
	var cmd = repo.cmdable(tx)

	userNotificationStreamKey := getUserNotificationStreamKey(userId)
	userNotificationData, err := json.Marshal(userNotification)
	if err != nil {
		return errors.E(op, err)
	}

	if err := cmd.XAdd(ctx, &redis.XAddArgs{
		Stream: userNotificationStreamKey,
		MaxLen: userNotificationStreamMaxlen,
		Values: map[string]any{
			streamEntryMessageField: userNotificationData,
		},
	}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) GetUserNotification(ctx context.Context, userId uuid.UUID, startId string) chan models.StreamSubscriptionResult[*models.UserNotification] {
	const op errors.Op = "redis_repo.RedisRepository.GetUserNotification"
	userNotificationStreamKey := getUserNotificationStreamKey(userId)

	return getStreamEntry[*models.UserNotification](ctx, repo, userNotificationStreamKey, startId, func(values map[string]interface{}, entryId string) (*models.UserNotification, error) {
		message, ok := values[streamEntryMessageField]
		if !ok {
			err := fmt.Errorf("%s key not found in %s map", streamEntryMessageField, values)
			return nil, errors.E(op, err)
		}

		messageStr, ok := message.(string)
		if !ok {
			err := fmt.Errorf("underlying type of message is not string")
			return nil, errors.E(op, err)
		}

		var userNotification models.UserNotification
		if err := json.Unmarshal([]byte(messageStr), &userNotification); err != nil {
			return nil, errors.E(op, err)
		}

		userNotification.Id = entryId

		return &userNotification, nil
	})
}

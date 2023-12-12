package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (repo *RedisRepository) PublishPushMessage(ctx context.Context, tx common.Transaction, roomId uuid.UUID, pushMessage models.PushMessage) error {
	const op errors.Op = "redis_repo.RedisRepository.PublishPushMessage"
	var roomStreamKey = getRoomStreamKey(roomId)
	var cmd = repo.cmdable(tx)

	pushMessageData, err := json.Marshal(pushMessage)
	if err != nil {
		return errors.E(op, err)
	}

	if err := cmd.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{
			streamEntryTypeField:    strconv.Itoa(int(models.PushMessageStreamEntryType)),
			streamEntryMessageField: pushMessageData,
		},
	}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) PublishAction(ctx context.Context, tx common.Transaction, roomId uuid.UUID, action models.StreamActionType) error {
	const op errors.Op = "redis_repo.RedisRepository.PublishAction"
	var roomStreamKey = getRoomStreamKey(roomId)
	var cmd = repo.cmdable(tx)

	if err := cmd.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]string{
			streamEntryTypeField:   strconv.Itoa(int(models.ActionStreamEntryType)),
			streamEntryActionField: strconv.Itoa(int(action)),
		},
	}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) GetPushMessages(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[[]byte] {
	const op errors.Op = "redis_repo.RedisRepository.GetPushMessages"
	var roomStreamKey = getRoomStreamKey(roomId)

	startId := ""
	if (startTime != time.Time{}) {
		startId = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	return getStreamEntry[[]byte](ctx, repo, roomStreamKey, startId, func(values map[string]interface{}, entryId string) ([]byte, error) {
		streamEntryType, err := getStreamEntryTypeFromMap(values)
		if err != nil {
			return nil, errors.E(op, err)
		}

		switch streamEntryType {
		case models.PushMessageStreamEntryType:
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

			return []byte(messageStr), nil
		case models.ActionStreamEntryType:
			if values[streamEntryActionField] == strconv.Itoa(int(models.TerminateAction)) {
				return nil, errors.E(op, errReceivedStreamTerminationAction)
			}

			return nil, errors.E(op, errIsIgnoredStreamEntry)
		default:
			return nil, errors.E(op, errIsIgnoredStreamEntry)
		}
	})
}

func (repo *RedisRepository) GetAction(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[models.StreamActionType] {
	const op errors.Op = "redis_repo.RedisRepository.GetAction"
	var roomStreamKey = getRoomStreamKey(roomId)

	startId := ""
	if (startTime != time.Time{}) {
		startId = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	return getStreamEntry[models.StreamActionType](ctx, repo, roomStreamKey, startId, func(values map[string]interface{}, entryId string) (models.StreamActionType, error) {
		streamEntryType, err := getStreamEntryTypeFromMap(values)
		if err != nil {
			return models.TerminateAction, errors.E(op, err)
		}

		switch streamEntryType {
		case models.ActionStreamEntryType:
			action, ok := values[streamEntryActionField]
			if !ok {
				err := fmt.Errorf("%s key not found in %s map", streamEntryActionField, values)
				return models.TerminateAction, errors.E(op, err)
			}

			if action == strconv.Itoa(int(models.TerminateAction)) {
				return models.TerminateAction, errors.E(op, errReceivedStreamTerminationAction)
			}

			actionStr, ok := action.(string)
			if !ok {
				err := fmt.Errorf("underlying type of action is not string")
				return models.TerminateAction, errors.E(op, err)
			}

			actionInt, err := strconv.Atoi(actionStr)
			if err != nil {
				return models.TerminateAction, errors.E(op, err)
			}

			return models.StreamActionType(actionInt), nil
		default:
			return models.TerminateAction, errors.E(op, errIsIgnoredStreamEntry)
		}
	})
}

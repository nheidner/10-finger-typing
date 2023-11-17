package redis_repo

import (
	"10-typing/models"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	streamEntryTypeField    = "type"
	streamEntryMessageField = "message"
	streamEntryActionField  = "action"
)

// rooms:[room_id]:stream stream: action: "terminate/..", data: message stringified json
func getRoomStreamKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":stream"
}

func (repo *RedisRepository) PublishPushMessage(ctx context.Context, roomId uuid.UUID, pushMessage models.PushMessage) error {
	roomStreamKey := getRoomStreamKey(roomId)
	pushMessageData, err := json.Marshal(pushMessage)
	if err != nil {
		return err
	}

	return repo.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{
			streamEntryTypeField:    strconv.Itoa(int(models.PushMessageStreamEntryType)),
			streamEntryMessageField: pushMessageData,
		},
	}).Err()
}

func (repo *RedisRepository) PublishAction(ctx context.Context, roomId uuid.UUID, action models.StreamActionType) error {
	roomStreamKey := getRoomStreamKey(roomId)

	return repo.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]string{
			streamEntryTypeField:   strconv.Itoa(int(models.ActionStreamEntryType)),
			streamEntryActionField: strconv.Itoa(int(action)),
		},
	}).Err()
}

func (repo *RedisRepository) GetPushMessages(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[[]byte] {
	roomStreamKey := getRoomStreamKey(roomId)
	startId := ""
	if (startTime != time.Time{}) {
		startId = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	return getStreamEntry[[]byte](ctx, repo, roomStreamKey, startId, func(values map[string]interface{}, entryId string) ([]byte, error) {
		streamEntryType, err := getStreamEntryTypeFromMap(values)
		if err != nil {
			return nil, err
		}

		switch streamEntryType {
		case models.PushMessageStreamEntryType:
			message, ok := values[streamEntryMessageField]
			if !ok {
				return nil, errors.New("no " + streamEntryMessageField + " field in stream entry")
			}

			messageStr, ok := message.(string)
			if !ok {
				return nil, errors.New("underlying type of " + streamEntryMessageField + " stream entry field is not string")
			}

			return []byte(messageStr), nil
		case models.ActionStreamEntryType:
			if values[streamEntryActionField] == strconv.Itoa(int(models.TerminateAction)) {
				return nil, errReceivedStreamTerminationAction
			}

			return nil, errIsIgnoredStreamEntry
		default:
			return nil, errIsIgnoredStreamEntry
		}
	})
}

func (repo *RedisRepository) GetAction(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[models.StreamActionType] {
	roomStreamKey := getRoomStreamKey(roomId)
	startId := ""
	if (startTime != time.Time{}) {
		startId = strconv.FormatInt(startTime.UnixMilli(), 10)
	}

	return getStreamEntry[models.StreamActionType](ctx, repo, roomStreamKey, startId, func(values map[string]interface{}, entryId string) (models.StreamActionType, error) {
		streamEntryType, err := getStreamEntryTypeFromMap(values)
		if err != nil {
			return models.TerminateAction, err
		}

		switch streamEntryType {
		case models.ActionStreamEntryType:
			action, ok := values[streamEntryActionField]
			if !ok {
				return models.TerminateAction, errors.New("no " + streamEntryActionField + " field in stream entry")
			}

			if action == strconv.Itoa(int(models.TerminateAction)) {
				return models.TerminateAction, errReceivedStreamTerminationAction
			}

			actionStr, ok := action.(string)
			if !ok {
				return models.TerminateAction, errors.New("underlying type of " + streamEntryActionField + " stream entry field is not string")
			}

			actionInt, err := strconv.Atoi(actionStr)
			if err != nil {
				return models.TerminateAction, errors.New("action cannot be converted to an integer: " + err.Error())
			}

			return models.StreamActionType(actionInt), nil
		default:
			return models.TerminateAction, errIsIgnoredStreamEntry
		}
	})
}

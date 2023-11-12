package redis_repo

import (
	"10-typing/models"
	"context"
	"encoding/json"
	"errors"
	"log"
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
			streamEntryTypeField:   strconv.Itoa(int(models.PushMessageStreamEntryType)),
			streamEntryActionField: strconv.Itoa(int(action)),
		},
	}).Err()
}

func (repo *RedisRepository) GetPushMessages(ctx context.Context, roomId uuid.UUID, startTime time.Time) (<-chan []byte, <-chan error) {
	out := make(chan []byte)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		roomStreamKey := getRoomStreamKey(roomId)
		id := "$"
		if (startTime != time.Time{}) {
			id = strconv.FormatInt(startTime.UnixMilli(), 10)
		}

		for {
			r, err := repo.redisClient.XRead(ctx, &redis.XReadArgs{
				Streams: []string{roomStreamKey, id},
				Count:   1,
				Block:   0,
			}).Result()
			if err != nil {
				errCh <- err
				return
			}

			id = r[0].Messages[0].ID
			values := r[0].Messages[0].Values

			streamEntryType, ok := values[streamEntryTypeField]
			if !ok {
				errCh <- errors.New("no " + streamEntryTypeField + " field in stream entry")
				return
			}

			switch streamEntryType {
			case strconv.Itoa(int(models.ActionStreamEntryType)):
				if values[streamEntryActionField] == strconv.Itoa(int(models.TerminateAction)) {
					log.Println("stream consumer is terminated")
					// TODO: shouldn't there an error be returned through the error channel
					return
				}
			case strconv.Itoa(int(models.PushMessageStreamEntryType)):
				messsage, ok := values[streamEntryMessageField]
				if !ok {
					errCh <- errors.New("no " + streamEntryMessageField + " field in stream entry")
					return
				}

				messsageStr, ok := messsage.(string)
				if !ok {
					errCh <- errors.New("underlying type of " + streamEntryMessageField + " stream entry field is not string")
					return
				}

				out <- []byte(messsageStr)
			default:
				errCh <- errors.New(streamEntryTypeField + " has not a correct value in stream entry")
				return
			}
		}
	}()

	return out, errCh
}

func (repo *RedisRepository) GetAction(ctx context.Context, roomId uuid.UUID, startTime time.Time) (<-chan models.StreamActionType, <-chan error) {
	out := make(chan models.StreamActionType)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		roomStreamKey := getRoomStreamKey(roomId)
		id := "$"
		if (startTime != time.Time{}) {
			id = strconv.FormatInt(startTime.UnixMilli(), 10)
		}

		for {
			select {
			// this does not really work because repo.redisClient.XRead is blocking and not a channel operation
			case <-ctx.Done():
				return
			default:
				r, err := repo.redisClient.XRead(ctx, &redis.XReadArgs{
					Streams: []string{roomStreamKey, id},
					Count:   1,
					Block:   0,
				}).Result()
				if err != nil {
					errCh <- err
					return
				}

				id = r[0].Messages[0].ID
				values := r[0].Messages[0].Values

				streamEntryType, ok := values[streamEntryTypeField]
				if !ok {
					errCh <- errors.New("no " + streamEntryTypeField + " field in stream entry")
					return
				}

				switch streamEntryType {
				case strconv.Itoa(int(models.ActionStreamEntryType)):
					action, ok := values[streamEntryActionField]
					if !ok {
						errCh <- errors.New("no " + streamEntryActionField + " field in stream entry")
						return
					}

					if action == strconv.Itoa(int(models.TerminateAction)) {
						log.Println("stream consumer is terminated")
						return
					}

					actionStr, ok := action.(string)
					if !ok {
						errCh <- errors.New("underlying type of " + streamEntryActionField + " stream entry field is not string")
						return
					}

					actionInt, err := strconv.Atoi(actionStr)
					if err != nil {
						errCh <- errors.New("action cannot be converted to an integer")
						return
					}

					out <- models.StreamActionType(actionInt)
				case strconv.Itoa(int(models.PushMessageStreamEntryType)):
				default:
					errCh <- errors.New(streamEntryTypeField + " has not a correct value in stream entry")
					return
				}
			}
		}

	}()

	return out, errCh
}

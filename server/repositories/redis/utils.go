package redis_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const xreadBlockingDurationSecs = 5

var (
	errReceivedStreamTerminationAction = errors.New("received stream termination action")
	errIsIgnoredStreamEntry            = errors.New("stream entry is ignored")
)

func getStreamEntry[T []byte | models.StreamActionType | *models.UserNotification](
	ctx context.Context,
	repo *RedisRepository,
	streamKey, startId string,
	processStreamEntry func(values map[string]interface{}, entryId string,
	) (T, error),
) chan models.StreamSubscriptionResult[T] {
	const op errors.Op = "redis_repo.getStreamEntry"
	var cmd redis.Cmdable = repo.redisClient

	out := make(chan models.StreamSubscriptionResult[T])

	go func() {
		defer close(out)

		id := "$"
		if startId != "" {
			id = startId
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				r, err := cmd.XRead(ctx, &redis.XReadArgs{
					Streams: []string{streamKey, id},
					Count:   1,
					Block:   xreadBlockingDurationSecs * time.Second,
				}).Result()
				if err == redis.Nil {
					continue
				}
				if err != nil {
					sendErrorResult[T](ctx, out, errors.E(op, err))
					return
				}

				id = r[0].Messages[0].ID
				values := r[0].Messages[0].Values

				v, err := processStreamEntry(values, id)
				switch {
				case errors.Is(err, errReceivedStreamTerminationAction):
					return
				case errors.Is(err, errIsIgnoredStreamEntry):
					continue
				case err != nil:
					sendErrorResult[T](ctx, out, errors.E(op, err))
					return
				}

				result := models.StreamSubscriptionResult[T]{
					Value: v,
				}

				select {
				case out <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

func sendErrorResult[T []byte | models.StreamActionType | *models.UserNotification](ctx context.Context, outCh chan<- models.StreamSubscriptionResult[T], err error) {
	result := models.StreamSubscriptionResult[T]{
		Error: err,
	}

	select {
	case outCh <- result:
	case <-ctx.Done():
	}
}

func getStreamEntryTypeFromMap(values map[string]any) (models.StreamEntryType, error) {
	const op errors.Op = "redis_repo.getStreamEntryTypeFromMap"

	streamEntryTypeAny, ok := values[streamEntryTypeField]
	if !ok {
		err := fmt.Errorf("%s key not found in %s map", streamEntryTypeField, values)
		return models.ActionStreamEntryType, errors.E(op, err)
	}

	streamEntryTypeStr, ok := streamEntryTypeAny.(string)
	if !ok {
		err := fmt.Errorf("streamEntryTypeAny's underlying value is not of type string")
		return models.ActionStreamEntryType, errors.E(op, err)
	}
	streamEntryTypeInt, err := strconv.Atoi(streamEntryTypeStr)
	if err != nil {
		return models.ActionStreamEntryType, errors.E(op, err)
	}

	return models.StreamEntryType(streamEntryTypeInt), nil
}

func deleteKeysByPattern(ctx context.Context, repo *RedisRepository, pattern string) error {
	const op errors.Op = "redis_repo.deleteKeysByPattern"
	var cmd redis.Cmdable = repo.redisClient

	iter := cmd.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		cmd.Del(ctx, key)

		if err := iter.Err(); err != nil {
			return errors.E(op)
		}
	}

	return nil
}

func stringsToUuids(uuidStrings []string) ([]uuid.UUID, error) {
	const op errors.Op = "redis_repo.stringsToUuids"

	uuids := make([]uuid.UUID, 0, len(uuidStrings))
	for _, gameUserIdStr := range uuidStrings {
		gameUserId, err := uuid.Parse(gameUserIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}

		uuids = append(uuids, gameUserId)
	}

	return uuids, nil
}

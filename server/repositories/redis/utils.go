package redis_repo

import (
	"10-typing/models"
	"context"
	"errors"
	"strconv"
	"time"

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
				r, err := repo.redisClient.XRead(ctx, &redis.XReadArgs{
					Streams: []string{streamKey, id},
					Count:   1,
					Block:   xreadBlockingDurationSecs * time.Second,
				}).Result()
				if err == redis.Nil {
					continue
				}
				if err != nil {
					sendErrorResult[T](ctx, out, err)
					return
				}

				id = r[0].Messages[0].ID
				values := r[0].Messages[0].Values

				v, err := processStreamEntry(values, id)
				if errors.Is(err, errReceivedStreamTerminationAction) {
					return
				}
				if errors.Is(err, errIsIgnoredStreamEntry) {
					continue
				}
				if err != nil {
					sendErrorResult[T](ctx, out, err)
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
	streamEntryTypeAny, ok := values[streamEntryTypeField]
	if !ok {
		err := errors.New("no " + streamEntryTypeField + " field in stream entry")
		return models.ActionStreamEntryType, err
	}

	streamEntryTypeStr, ok := streamEntryTypeAny.(string)
	if !ok {
		err := errors.New("streamEntryTypeAny is not of type string")
		return models.ActionStreamEntryType, err
	}
	streamEntryTypeInt, err := strconv.Atoi(streamEntryTypeStr)
	if err != nil {
		return models.ActionStreamEntryType, err
	}

	return models.StreamEntryType(streamEntryTypeInt), nil
}

func deleteKeysByPattern(ctx context.Context, repo *RedisRepository, pattern string) error {
	iter := repo.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		repo.redisClient.Del(ctx, key)

		return iter.Err()
	}

	return nil
}

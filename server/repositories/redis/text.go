package redis_repo

import (
	"context"

	"github.com/google/uuid"
)

const (
	// text_ids set: text ids
	textIdsKey = "text_ids"
)

func (repo *RedisRepository) SetTextId(ctx context.Context, textIds ...uuid.UUID) error {
	textIdsStr := make([]any, 0, len(textIds))
	for _, textId := range textIds {
		textIdsStr = append(textIdsStr, textId.String())
	}

	return repo.redisClient.SAdd(ctx, textIdsKey, textIdsStr...).Err()
}

func (repo *RedisRepository) TextIdsKeyExists(ctx context.Context) (bool, error) {
	r, err := repo.redisClient.Exists(ctx, textIdsKey).Result()

	return r != 0, err
}

func (repo *RedisRepository) TextIdExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	r, err := repo.redisClient.SMIsMember(ctx, textIdsKey, textId.String()).Result()
	if err != nil {
		return false, err
	}

	return r[0], nil
}

func (repo *RedisRepository) DeleteTextIdsKey(ctx context.Context) error {
	return repo.redisClient.Del(ctx, textIdsKey).Err()
}

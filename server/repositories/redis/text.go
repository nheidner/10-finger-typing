package redis_repo

import (
	"10-typing/errors"
	"context"

	"github.com/google/uuid"
)

const (
	// text_ids set: text ids
	textIdsKey = "text_ids"
)

func (repo *RedisRepository) SetTextId(ctx context.Context, textIds ...uuid.UUID) error {
	const op errors.Op = "redis_repo.SetTextId"
	textIdsStr := make([]any, 0, len(textIds))

	for _, textId := range textIds {
		textIdsStr = append(textIdsStr, textId.String())
	}

	if err := repo.redisClient.SAdd(ctx, textIdsKey, textIdsStr...).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) TextIdsKeyExists(ctx context.Context) (bool, error) {
	const op errors.Op = "redis_repo.TextIdsKeyExists"

	r, err := repo.redisClient.Exists(ctx, textIdsKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r != 0, nil
}

func (repo *RedisRepository) TextIdExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.TextIdExists"

	r, err := repo.redisClient.SMIsMember(ctx, textIdsKey, textId.String()).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r[0], nil
}

func (repo *RedisRepository) DeleteTextIdsKey(ctx context.Context) error {
	const op errors.Op = "redis_repo.DeleteTextIdsKey"

	if err := repo.redisClient.Del(ctx, textIdsKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

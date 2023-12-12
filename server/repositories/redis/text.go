package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (repo *RedisRepository) SetTextId(ctx context.Context, tx common.Transaction, textIds ...uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetTextId"
	var cmd = repo.cmdable(tx)

	textIdsStr := make([]any, 0, len(textIds))
	for _, textId := range textIds {
		textIdsStr = append(textIdsStr, textId.String())
	}

	if err := cmd.SAdd(ctx, textIdsKey, textIdsStr...).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) TextIdsKeyExists(ctx context.Context) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.TextIdsKeyExists"
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.Exists(ctx, textIdsKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r != 0, nil
}

func (repo *RedisRepository) TextIdExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.TextIdExists"
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.SMIsMember(ctx, textIdsKey, textId.String()).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r[0], nil
}

func (repo *RedisRepository) DeleteTextIdsKey(ctx context.Context, tx common.Transaction) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteTextIdsKey"
	var cmd = repo.cmdable(tx)

	if err := cmd.Del(ctx, textIdsKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

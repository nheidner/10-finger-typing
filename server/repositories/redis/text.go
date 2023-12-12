package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"context"

	"github.com/google/uuid"
)

func (repo *RedisRepository) SetTextId(ctx context.Context, tx common.Transaction, textIds ...uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetTextId"
	textIdsStr := make([]any, 0, len(textIds))
	var cmd = repo.cmdable(tx)

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

	r, err := repo.redisClient.Exists(ctx, textIdsKey).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return r != 0, nil
}

func (repo *RedisRepository) TextIdExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.TextIdExists"

	r, err := repo.redisClient.SMIsMember(ctx, textIdsKey, textId.String()).Result()
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

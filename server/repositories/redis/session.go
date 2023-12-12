package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"time"

	"github.com/google/uuid"
)

func (repo *RedisRepository) SetSession(ctx context.Context, tx common.Transaction, tokenHash string, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetSession"
	sessionKey := getSessionKey(tokenHash)
	var cmd = repo.cmdable(tx)

	if err := cmd.Set(ctx, sessionKey, userId.String(), models.SessionDurationSec*time.Second).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteSession(ctx context.Context, tx common.Transaction, tokenHash string) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteSession"
	sessionKey := getSessionKey(tokenHash)
	var cmd = repo.cmdable(tx)

	if err := cmd.Del(ctx, sessionKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllSessions(ctx context.Context) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteAllSessions"

	if err := deleteKeysByPattern(ctx, repo, "sessions:*"); err != nil {
		return errors.E(op, err)
	}

	return nil
}

package redis_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"context"
	"time"

	"github.com/google/uuid"
)

// users:[tokenhash] user_id
func getSessionKey(tokenHash string) string {
	return "sessions:" + tokenHash
}

func (repo *RedisRepository) SetSession(ctx context.Context, tokenHash string, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.SetSession"
	sessionKey := getSessionKey(tokenHash)

	if err := repo.redisClient.Set(ctx, sessionKey, userId.String(), models.SessionDurationSec*time.Second).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	const op errors.Op = "redis_repo.DeleteSession"
	sessionKey := getSessionKey(tokenHash)

	if err := repo.redisClient.Del(ctx, sessionKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllSessions(ctx context.Context) error {
	const op errors.Op = "redis_repo.DeleteAllSessions"

	if err := deleteKeysByPattern(ctx, repo, "sessions:*"); err != nil {
		return errors.E(op, err)
	}

	return nil
}

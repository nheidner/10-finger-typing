package redis_repo

import (
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
	sessionKey := getSessionKey(tokenHash)

	return repo.redisClient.Set(ctx, sessionKey, userId.String(), models.SessionDurationSec*time.Second).Err()
}

func (repo *RedisRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	sessionKey := getSessionKey(tokenHash)

	return repo.redisClient.Del(ctx, sessionKey).Err()
}

func (repo *RedisRepository) DeleteAllSessions(ctx context.Context) error {
	return deleteKeysByPattern(ctx, repo, "sessions:*")
}

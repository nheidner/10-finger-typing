package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// text_ids set: text ids
	textIdsKey = "text_ids"
)

type TextRedisRepository struct {
	redisClient *redis.Client
}

func NewTextRedisRepository(redisClient *redis.Client) *TextRedisRepository {
	return &TextRedisRepository{redisClient}
}

func (tr *TextRedisRepository) CreateInRedis(ctx context.Context, textIds ...uuid.UUID) error {
	textIdsStr := make([]any, 0, len(textIds))
	for _, textId := range textIds {
		textIdsStr = append(textIdsStr, textId.String())
	}

	return tr.redisClient.SAdd(ctx, textIdsKey, textIdsStr...).Err()
}

func (tr *TextRedisRepository) AllTextsAreInRedis(ctx context.Context) (bool, error) {
	r, err := tr.redisClient.Exists(ctx, textIdsKey).Result()

	return r != 0, err
}

func (tr *TextRedisRepository) TextExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	r, err := tr.redisClient.SMIsMember(ctx, textIdsKey, textId.String()).Result()
	if err != nil {
		return false, err
	}

	return r[0], nil
}

func (tr *TextRedisRepository) DeleteAllFromRedis() error {
	return tr.redisClient.Del(context.Background(), textIdsKey).Err()
}

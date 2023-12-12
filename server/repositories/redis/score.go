package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (repo *RedisRepository) GetCurrentGameScores(ctx context.Context, roomId uuid.UUID) ([]models.Score, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameScores"
	currentGameScoresUserIdsKey := getCurrentGameScoresUserIdsKey(roomId)

	userIdStrs, err := repo.redisClient.ZRevRange(ctx, currentGameScoresUserIdsKey, 0, -1).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	scores := make([]models.Score, 0, len(userIdStrs))

	for _, userIdStr := range userIdStrs {
		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}

		currentGameScoreKey := getCurrentGameScoreKey(roomId, userId)

		scoreStr, err := repo.redisClient.Get(ctx, currentGameScoreKey).Result()
		if err != nil {
			return nil, errors.E(op, err)
		}

		var score models.Score
		if err := json.Unmarshal([]byte(scoreStr), &score); err != nil {
			return nil, errors.E(op, err)
		}

		scores = append(scores, score)
	}

	return scores, nil
}

func (repo *RedisRepository) SetCurrentGameScore(ctx context.Context, tx common.Transaction, roomId uuid.UUID, score models.Score) error {
	const op errors.Op = "redis_repo.RedisRepository.SetCurrentGameScore"

	scoreJson, err := json.Marshal(&score)
	if err != nil {
		return errors.E(op, err)
	}

	// PIPELINE start if no outer pipeline exists
	cmd, innerTx := repo.beginPipelineIfNoOuterTransactionExists(tx)

	currentGameScoresKey := getCurrentGameScoreKey(roomId, score.UserId)
	cmd.Set(ctx, currentGameScoresKey, scoreJson, 0)

	currentGameScoresUserIdsKey := getCurrentGameScoresUserIdsKey(roomId)
	scoreUserId := redis.Z{Score: score.WordsPerMinute, Member: score.UserId.String()}
	cmd.ZAdd(ctx, currentGameScoresUserIdsKey, scoreUserId)

	// PIPELINE commit
	if innerTx != nil {
		if err := innerTx.Commit(ctx); err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (repo *RedisRepository) DeleteCurrentGameScores(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteCurrentGameScores"

	pattern := getCurrentGameKey(roomId) + ":scores:*"
	if err := deleteKeysByPattern(ctx, repo, pattern); err != nil {
		return errors.E(op, err)
	}

	return nil
}

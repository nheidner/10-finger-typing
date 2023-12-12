package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// GET METHODS
func (repo *RedisRepository) GetCurrentGameUserIds(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameUserIds"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameUserIdsKey = getCurrentGameUserIdsKey(roomId)

	r, err := cmd.SMembers(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	gameUserIds, err := stringsToUuids(r)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return gameUserIds, nil
}

func (repo *RedisRepository) GetCurrentGameUsersNumber(ctx context.Context, roomId uuid.UUID) (int, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameUsersNumber"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameUserIdsKey = getCurrentGameUserIdsKey(roomId)

	r, err := cmd.SCard(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return 0, errors.E(op, err)
	}

	return int(r), nil
}

func (repo *RedisRepository) GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (models.GameStatus, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameStatus"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameKey = getCurrentGameKey(roomId)

	gameStatusInt, err := cmd.HGet(ctx, currentGameKey, currentGameStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.UnstartedGameStatus, errors.E(op, common.ErrNotFound)
	case err != nil:
		return models.UnstartedGameStatus, errors.E(op, err)
	}

	return models.GameStatus(gameStatusInt), nil
}

func (repo *RedisRepository) GetCurrentGameId(ctx context.Context, roomId uuid.UUID) (uuid.UUID, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameId"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameKey = getCurrentGameKey(roomId)

	gameIdStr, err := cmd.HGet(ctx, currentGameKey, currentGameIdField).Result()
	switch {
	case err == redis.Nil:
		return uuid.Nil, errors.E(op, common.ErrNotFound)
	case err != nil:
		return uuid.Nil, errors.E(op, err)
	}

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	return gameId, nil
}

func (repo *RedisRepository) GetCurrentGame(ctx context.Context, roomId uuid.UUID) (*models.Game, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGame"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameKey = getCurrentGameKey(roomId)

	r, err := cmd.HGetAll(ctx, currentGameKey).Result()
	switch {
	case err != nil:
		return nil, err
	case len(r) == 0:
		return nil, errors.E(op, common.ErrNotFound)
	}

	status := models.UnstartedGameStatus
	statusStr, ok := r[currentGameStatusField]
	if ok {
		statusInt, err := strconv.Atoi(statusStr)
		if err != nil {
			return nil, errors.E(op, err)
		}

		status = models.GameStatus(statusInt)
	}

	textId := uuid.UUID{}
	textIdStr, ok := r[currentGameTextIdField]
	if ok {
		textId, err = uuid.Parse(textIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	gameId := uuid.UUID{}
	gameIdStr, ok := r[currentGameIdField]
	if ok {
		gameId, err = uuid.Parse(gameIdStr)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	return &models.Game{
		ID:     gameId,
		TextId: textId,
		RoomId: roomId,
		Status: status,
	}, nil
}

// SET METHODS
func (repo *RedisRepository) SetNewCurrentGame(ctx context.Context, tx common.Transaction, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetNewCurrentGame"
	var cmd = repo.cmdable(tx)
	var currentGameKey = getCurrentGameKey(roomId)

	if len(userIds) == 0 {
		err := fmt.Errorf("at least one user id must be specified")
		return errors.E(op, err)
	}

	statusStr := strconv.Itoa(int(models.UnstartedGameStatus))
	gameIdStr := newGameId.String()
	textIdStr := textId.String()
	currentGameValue := map[string]string{
		currentGameIdField:     gameIdStr,
		currentGameTextIdField: textIdStr,
		currentGameStatusField: statusStr,
	}
	if err := cmd.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) SetCurrentGameUser(ctx context.Context, tx common.Transaction, roomId, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetCurrentGameUser"
	var cmd = repo.cmdable(tx)
	var currentGameUserIdsKey = getCurrentGameUserIdsKey(roomId)

	if err := cmd.SAdd(ctx, currentGameUserIdsKey, userId.String()).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) SetCurrentGameStatus(ctx context.Context, tx common.Transaction, roomId uuid.UUID, gameStatus models.GameStatus) error {
	const op errors.Op = "redis_repo.RedisRepository.SetCurrentGameStatus"
	currentGameKey := getCurrentGameKey(roomId)
	var cmd = repo.cmdable(tx)

	if err := cmd.HSet(ctx, currentGameKey, currentGameStatusField, strconv.Itoa(int(gameStatus))).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllCurrentGameUsers(ctx context.Context, tx common.Transaction, roomId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteAllCurrentGameUsers"
	var currentGameUserIdsKey = getCurrentGameUserIdsKey(roomId)
	var cmd = repo.cmdable(tx)

	if err := cmd.Del(ctx, currentGameUserIdsKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// IS.. METHODS
func (repo *RedisRepository) IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.IsCurrentGame"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameKey = getCurrentGameKey(roomId)

	r, err := cmd.HGet(ctx, currentGameKey, currentGameIdField).Result()
	switch {
	case err == redis.Nil:
		return false, nil
	case err != nil:
		return false, errors.E(op, err)
	}

	return r == gameId.String(), nil
}

func (repo *RedisRepository) IsCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) (isCurrentGameUser bool, err error) {
	const op errors.Op = "redis_repo.RedisRepository.IsCurrentGameUser"
	var cmd redis.Cmdable = repo.redisClient
	var currentGameUserIdsKey = getCurrentGameUserIdsKey(roomId)

	isCurrentGameUser, err = cmd.SIsMember(ctx, currentGameUserIdsKey, userId.String()).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return isCurrentGameUser, nil
}

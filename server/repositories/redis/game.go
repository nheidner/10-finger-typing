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

const (
	currentGameStatusField = "status"
	currentGameIdField     = "game_id"
	currentGameTextIdField = "text_id"
)

// rooms:[room_id]:current_game hash: id, text_id, status
func getCurrentGameKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":current_game"
}

// rooms:[room_id]:current_game:user_ids set: game user ids
func getCurrentGameUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":user_ids"
}

// GET METHODS
func (repo *RedisRepository) GetCurrentGameUserIds(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameUserIds"
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := repo.redisClient.SMembers(ctx, currentGameUserIdsKey).Result()
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
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := repo.redisClient.SCard(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return 0, errors.E(op, err)
	}

	return int(r), nil
}

func (repo *RedisRepository) GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (models.GameStatus, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetCurrentGameStatus"
	currentGameKey := getCurrentGameKey(roomId)

	gameStatusInt, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameStatusField).Int()
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
	currentGameKey := getCurrentGameKey(roomId)

	gameIdStr, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
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
	currentGameKey := getCurrentGameKey(roomId)

	r, err := repo.redisClient.HGetAll(ctx, currentGameKey).Result()
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
func (repo *RedisRepository) SetNewCurrentGame(ctx context.Context, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetNewCurrentGame"
	if len(userIds) == 0 {
		err := fmt.Errorf("at least one user id must be specified")
		return errors.E(op, err)
	}

	currentGameKey := getCurrentGameKey(roomId)
	statusStr := strconv.Itoa(int(models.UnstartedGameStatus))
	gameIdStr := newGameId.String()
	textIdStr := textId.String()
	currentGameValue := map[string]string{
		currentGameIdField:     gameIdStr,
		currentGameTextIdField: textIdStr,
		currentGameStatusField: statusStr,
	}
	if err := repo.redisClient.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return errors.E(op, err)
	}

	// TODO: isnt that wrong
	// currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)
	// userIdStrs := make([]interface{}, 0, len(userIds))
	// for _, userId := range userIds {
	// 	userIdStrs = append(userIdStrs, userId.String())
	// }
	// if err := repo.redisClient.SAdd(ctx, currentGameUserIdsKey, userIdStrs...).Err(); err != nil {
	// 	return errors.E(op, err)
	// }

	return nil
}

func (repo *RedisRepository) SetCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.SetCurrentGameUser"
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	if err := repo.redisClient.SAdd(ctx, currentGameUserIdsKey, userId.String()).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) SetCurrentGameStatus(ctx context.Context, roomId uuid.UUID, gameStatus models.GameStatus) error {
	const op errors.Op = "redis_repo.RedisRepository.SetCurrentGameStatus"
	currentGameKey := getCurrentGameKey(roomId)

	if err := repo.redisClient.HSet(ctx, currentGameKey, currentGameStatusField, strconv.Itoa(int(gameStatus))).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllCurrentGameUsers(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteAllCurrentGameUsers"
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	if err := repo.redisClient.Del(ctx, currentGameUserIdsKey).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// IS.. METHODS
func (repo *RedisRepository) IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.IsCurrentGame"
	currentGameKey := getCurrentGameKey(roomId)

	r, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
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
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	isCurrentGameUser, err = repo.redisClient.SIsMember(ctx, currentGameUserIdsKey, userId.String()).Result()
	if err != nil {
		return false, errors.E(op, err)
	}

	return isCurrentGameUser, nil
}

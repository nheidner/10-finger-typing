package redis_repo

import (
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
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := repo.redisClient.SMembers(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return nil, err
	}

	gameUserIds := make([]uuid.UUID, 0, len(r))
	for _, gameUserIdStr := range r {
		gameUserId, err := uuid.Parse(gameUserIdStr)
		if err != nil {
			return nil, err
		}

		gameUserIds = append(gameUserIds, gameUserId)
	}

	return gameUserIds, nil
}

func (repo *RedisRepository) GetCurrentGameUsersNumber(ctx context.Context, roomId uuid.UUID) (int, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := repo.redisClient.SCard(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return 0, err
	}

	return int(r), nil
}

func (repo *RedisRepository) GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (models.GameStatus, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameStatusInt, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.UnstartedGameStatus, nil
	case err != nil:
		return models.UnstartedGameStatus, err
	}

	return models.GameStatus(gameStatusInt), nil
}

func (repo *RedisRepository) GetCurrentGameId(ctx context.Context, roomId uuid.UUID) (uuid.UUID, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameIdStr, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return uuid.Nil, err
	}

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		return uuid.Nil, err
	}

	return gameId, nil
}

func (repo *RedisRepository) GetCurrentGame(ctx context.Context, roomId uuid.UUID) (*models.Game, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := repo.redisClient.HGetAll(ctx, currentGameKey).Result()
	if err != nil {
		return nil, err
	}

	status := models.UnstartedGameStatus
	statusStr, ok := r[currentGameStatusField]
	if ok {
		statusInt, err := strconv.Atoi(statusStr)
		if err != nil {
			return nil, err
		}

		status = models.GameStatus(statusInt)
	}

	textId := uuid.UUID{}
	textIdStr, ok := r[currentGameTextIdField]
	if ok {
		textId, err = uuid.Parse(textIdStr)
		if err != nil {
			return nil, err
		}
	}

	gameId := uuid.UUID{}
	gameIdStr, ok := r[currentGameIdField]
	if ok {
		gameId, err = uuid.Parse(gameIdStr)
		if err != nil {
			return nil, err
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
	if len(userIds) == 0 {
		return fmt.Errorf("at least one user id must be specified")
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
		return err
	}

	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)
	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}
	if err := repo.redisClient.SAdd(ctx, currentGameUserIdsKey, userIdStrs...).Err(); err != nil {
		return err
	}

	return nil
}

func (repo *RedisRepository) SetGameUser(ctx context.Context, roomId, userId uuid.UUID) error {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return repo.redisClient.SAdd(ctx, currentGameUserIdsKey, userId.String()).Err()
}

func (repo *RedisRepository) SetCurrentGameStatus(ctx context.Context, roomId uuid.UUID, gameStatus models.GameStatus) error {
	currentGameKey := getCurrentGameKey(roomId)

	return repo.redisClient.HSet(ctx, currentGameKey, currentGameStatusField, strconv.Itoa(int(gameStatus))).Err()
}

// IS.. METHODS
func (repo *RedisRepository) IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := repo.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return false, err
	}

	return r == gameId.String(), nil
}

func (repo *RedisRepository) IsCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) (bool, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return repo.redisClient.SIsMember(ctx, currentGameUserIdsKey, userId.String()).Result()
}

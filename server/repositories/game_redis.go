package repositories

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

type GameRedisRepository struct {
	redisClient *redis.Client
}

func NewGameRedisRepository(redisClient *redis.Client) *GameRedisRepository {
	return &GameRedisRepository{redisClient}
}

// GET METHODS
func (gr *GameRedisRepository) GetCurrentGameUsersIds(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := gr.redisClient.SMembers(ctx, currentGameUserIdsKey).Result()
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

func (gr *GameRedisRepository) GetCurrentGameUsersNumber(ctx context.Context, roomId uuid.UUID) (int, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := gr.redisClient.SCard(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return 0, err
	}

	return int(r), nil
}

func (gr *GameRedisRepository) GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (models.GameStatus, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameStatusInt, err := gr.redisClient.HGet(ctx, currentGameKey, currentGameStatusField).Int()
	switch {
	case err == redis.Nil:
		return models.NilGameStatus, nil
	case err != nil:
		return models.NilGameStatus, err
	}

	return models.GameStatus(gameStatusInt), nil
}

func (gr *GameRedisRepository) GetCurrentGameId(ctx context.Context, roomId uuid.UUID) (uuid.UUID, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameIdStr, err := gr.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return uuid.Nil, err
	}

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		return uuid.Nil, err
	}

	return gameId, nil
}

func (gr *GameRedisRepository) GetCurrentGameFromRedis(ctx context.Context, roomId uuid.UUID) (*models.Game, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := gr.redisClient.HGetAll(ctx, currentGameKey).Result()
	if err != nil {
		return nil, err
	}

	status := models.NilGameStatus
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
func (gr *GameRedisRepository) SetNewCurrentGameInRedis(ctx context.Context, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
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
	if err := gr.redisClient.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return err
	}

	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)
	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}
	if err := gr.redisClient.SAdd(ctx, currentGameUserIdsKey, userIdStrs...).Err(); err != nil {
		return err
	}

	return nil
}

func (gr *GameRedisRepository) AddGameUserInRedis(ctx context.Context, roomId, userId uuid.UUID) error {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return gr.redisClient.SAdd(ctx, currentGameUserIdsKey, userId).Err()
}

func (gr *GameRedisRepository) SetCurrentGameStatusInRedis(ctx context.Context, roomId uuid.UUID, gameStatus models.GameStatus) error {
	currentGameKey := getCurrentGameKey(roomId)

	return gr.redisClient.HSet(ctx, currentGameKey, currentGameStatusField, gameStatus).Err()
}

// IS.. METHODS
func (gr *GameRedisRepository) IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := gr.redisClient.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return false, err
	}

	return r == gameId.String(), nil
}

func (gr *GameRedisRepository) IsCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) (bool, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return gr.redisClient.SIsMember(ctx, currentGameUserIdsKey, userId.String()).Result()
}

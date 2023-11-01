package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Game struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	TextId    uuid.UUID       `json:"textId" gorm:"not null"`
	RoomId    uuid.UUID       `json:"roomId" gorm:"not null"`
	Scores    []Score         `json:"-"`
	Status    GameStatus      `json:"status" gorm:"-"`
}

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

type GameStatus int

const (
	NilGameStatus GameStatus = iota
	UnstartedGameStatus
	CountdownGameStatus
	StartedGameStatus
	FinishedGameStatus
)

func (s *GameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *GameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type GameService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

type CreateGameInput struct {
	TextId uuid.UUID `json:"textId"`
}

func (gs *GameService) SetNewCurrentGame(ctx context.Context, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
	if len(userIds) == 0 {
		return fmt.Errorf("at least one user id must be specified")
	}

	currentGameKey := getCurrentGameKey(roomId)
	statusStr := strconv.Itoa(int(UnstartedGameStatus))
	gameIdStr := newGameId.String()
	currentGameValue := map[string]string{
		currentGameIdField:     gameIdStr,
		currentGameTextIdField: textId.String(),
		currentGameStatusField: statusStr,
	}
	if err := gs.RDB.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return err
	}

	currentGameUseridsKey := getCurrentGameUserIdsKey(roomId)
	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}
	if err := gs.RDB.SAdd(ctx, currentGameUseridsKey, userIdStrs...).Err(); err != nil {
		return err
	}

	return nil
}

func (gs *GameService) IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := gs.RDB.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return false, err
	}

	return r == gameId.String(), nil
}

func (gs *GameService) AddGameUser(ctx context.Context, roomId, userId uuid.UUID) error {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return gs.RDB.SAdd(ctx, currentGameUserIdsKey, userId).Err()
}

func (gs *GameService) GetCurrentGameUsers(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := gs.RDB.SMembers(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return nil, err
	}

	gameUsers := make([]uuid.UUID, 0, len(r))
	for _, gu := range r {
		gameUser, err := uuid.Parse(gu)
		if err != nil {
			return nil, err
		}

		gameUsers = append(gameUsers, gameUser)
	}

	return gameUsers, nil
}

func (gs *GameService) IsCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) (bool, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	return gs.RDB.SIsMember(ctx, currentGameUserIdsKey, userId.String()).Result()
}

func (gs *GameService) GetCurrentGameUsersNumber(ctx context.Context, roomId uuid.UUID) (int, error) {
	currentGameUserIdsKey := getCurrentGameUserIdsKey(roomId)

	r, err := gs.RDB.SCard(ctx, currentGameUserIdsKey).Result()
	if err != nil {
		return 0, err
	}

	return int(r), nil
}

func (gs *GameService) GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (GameStatus, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameStatusInt, err := gs.RDB.HGet(ctx, currentGameKey, currentGameStatusField).Int()
	switch {
	case err == redis.Nil:
		return NilGameStatus, nil
	case err != nil:
		return NilGameStatus, err
	}

	return GameStatus(gameStatusInt), nil
}

func (gs *GameService) GetCurrentGameId(ctx context.Context, roomId uuid.UUID) (uuid.UUID, error) {
	currentGameKey := getCurrentGameKey(roomId)

	gameIdStr, err := gs.RDB.HGet(ctx, currentGameKey, currentGameIdField).Result()
	if err != nil {
		return uuid.Nil, err
	}

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		return uuid.Nil, err
	}

	return gameId, nil
}

func (gs *GameService) SetCurrentGameStatus(ctx context.Context, roomId uuid.UUID, gameStatus GameStatus) error {
	currentGameKey := getCurrentGameKey(roomId)

	return gs.RDB.HSet(ctx, currentGameKey, currentGameStatusField, gameStatus).Err()
}

func (gs *GameService) GetCurrentGame(ctx context.Context, roomId uuid.UUID) (*Game, error) {
	currentGameKey := getCurrentGameKey(roomId)

	r, err := gs.RDB.HGetAll(ctx, currentGameKey).Result()
	if err != nil {
		return nil, err
	}

	status := NilGameStatus
	statusStr, ok := r[currentGameStatusField]
	if ok {
		statusInt, err := strconv.Atoi(statusStr)
		if err != nil {
			return nil, err
		}

		status = GameStatus(statusInt)
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

	return &Game{
		ID:     gameId,
		TextId: textId,
		RoomId: roomId,
		Status: status,
	}, nil
}

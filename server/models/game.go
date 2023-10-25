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

type GameStatus int

const (
	NilGameStatus GameStatus = iota
	UnstartedGameStatus
	StartedGameStatus
	FinishedGameStatus
)

func (s *GameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *GameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Subscriber struct {
	UserId         uuid.UUID        `json:"userId"`
	StartTimeStamp *time.Time       `json:"startTime"`
	Status         SubscriberStatus `json:"status"`
}

type Game struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	TextId      uuid.UUID       `json:"textId" gorm:"not null"`
	RoomId      uuid.UUID       `json:"roomId" gorm:"not null"`
	Scores      []Score         `json:"-"`
	Status      GameStatus      `json:"status" gorm:"-"`
	Subscribers []Subscriber    `json:"subscribers" gorm:"-"`
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
		"id":      gameIdStr,
		"text_id": textId.String(),
		"status":  statusStr,
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

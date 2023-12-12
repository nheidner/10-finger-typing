package models

import (
	"10-typing/errors"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Game struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	DeletedAt       *gorm.DeletedAt `json:"-" gorm:"index"`
	TextId          uuid.UUID       `json:"textId" gorm:"not null"`
	RoomId          uuid.UUID       `json:"roomId" gorm:"not null"`
	GameSubscribers []uuid.UUID     `json:"gameSubscribers" gorm:"-"`
	// Scores    []Score         `json:"-"` // TODO: cannot have foreign key fk_games_scores for case when adding game score before adding game
	Status GameStatus `json:"status" gorm:"-"`
}

type GameStatus int

const (
	UnstartedGameStatus GameStatus = iota
	CountdownGameStatus
	StartedGameStatus
	FinishedGameStatus
)

func (s *GameStatus) String() (string, error) {
	const op errors.Op = "models.GameStatus.String"
	fields := []string{"unstarted", "countdown", "started", "finished"}

	if int(*s) >= len(fields) {
		err := fmt.Errorf("invalid GameStatus")
		return "", errors.E(op, err)
	}

	return fields[*s], nil
}

func (s *GameStatus) MarshalJSON() ([]byte, error) {
	const op errors.Op = "models.GameStatus.MarshalJSON"
	gameStatusStr, err := s.String()
	if err != nil {
		return nil, errors.E(op, err)
	}

	gameStatusJson, err := json.Marshal(gameStatusStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return gameStatusJson, nil
}

type CreateGameInput struct {
	TextId uuid.UUID `json:"textId"`
}

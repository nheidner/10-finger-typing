package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

type CreateGameInput struct {
	TextId uuid.UUID `json:"textId"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameUser struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	DeletedAt  *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	UserId     uint            `json:"userId" gorm:"not null"`
	GameId     uuid.UUID       `json:"gameId" gorm:"not null"`
	ScoreId    uint            `json:"scoreId" gorm:"not null"`
	IsActive   bool            `json:"isActive" gorm:"not null;default:false"`
	IsFinished bool            `json:"isFinished" gorm:"not null;default:false"`
	IsAdmin    bool            `json:"isAdmin" gorm:"not null;default:false"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time       `json:"createdAt"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Room      Room            `json:"-"`
	RoomID    uuid.UUID       `json:"-"`
	IsUsed    bool            `json:"-"`
}

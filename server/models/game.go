package models

import (
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
	TextId    uint            `json:"textId" gorm:"not null"`
	RoomId    uuid.UUID       `json:"roomId" gorm:"not null"`
	Room      Room            `json:"-"`
	GameUsers []GameUser      `json:"-"`
}

type GameService struct {
	DB    *gorm.DB
	Redis *redis.Client
}

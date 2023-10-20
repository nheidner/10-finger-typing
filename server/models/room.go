package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// rooms:[roomId] hash {id, ... }
// rooms:[roomId]:subscribers set of userIds
// rooms:[roomId]:active_game hash {}
// rooms:[roomId]:active_game:users set of userIds
// conns:[userId] set of connection ids

type Room struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Subscribers []User          `json:"subscribers" gorm:"many2many:user_rooms"`
	Tokens      []Token         `json:"-"`
	Games       []Game          `json:"-"`
}

type CreateRoomInput struct {
	UserIds []uuid.UUID `json:"userIds"`
	Emails  []string    `json:"emails" binding:"dive,email"`
}

type RoomService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

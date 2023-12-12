package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Room struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	DeletedAt       *gorm.DeletedAt `json:"-" gorm:"index"`
	Users           []User          `json:"-" gorm:"many2many:user_rooms"` // saved in DB
	AdminId         uuid.UUID       `json:"adminId" gorm:"not null"`
	Admin           User            `json:"-" gorm:"foreignKey:AdminId"`
	Tokens          []Token         `json:"-"`
	Games           []Game          `json:"-"`
	GameDurationSec int             `json:"gameDurationSec" gorm:"default:5;not null"`
}

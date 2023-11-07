package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	SessionDuration = 60 * 60 * 24 * 7 // 1 week in seconds
	CookieSession   = "SID"
)

type Session struct {
	*gorm.Model
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserId    uuid.UUID `json:"userId"`
	Token     string    `json:"token" gorm:"-"` // token that is not saved in the database
	TokenHash string    `json:"tokenHash" gorm:"not null;type:varchar(510)"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Text struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	DeletedAt         *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Language          string          `json:"language" gorm:"not null;type:varchar(255)"`
	Text              string          `json:"text" gorm:"not null;type:text"`
	Punctuation       bool            `json:"punctuation" gorm:"not null;default:false"`
	SpecialCharacters int             `json:"specialCharacters" gorm:"not null;default:0"`
	Numbers           int             `json:"numbers" gorm:"not null;default:0"`
	Scores            []Score         `json:"-"`
	Games             []Game          `json:"-"`
}

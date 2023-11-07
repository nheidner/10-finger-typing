package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Text struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" faker:"-"`
	CreatedAt         time.Time       `json:"createdAt" faker:"-"`
	UpdatedAt         time.Time       `json:"updatedAt" faker:"-"`
	DeletedAt         *gorm.DeletedAt `json:"deletedAt" gorm:"index" faker:"-"`
	Language          string          `json:"language" gorm:"not null;type:varchar(255)" faker:"oneof: de, en, fr"`
	Text              string          `json:"text" gorm:"not null;type:text" faker:"-"`
	Punctuation       bool            `json:"punctuation" gorm:"not null;default:false"`
	SpecialCharacters int             `json:"specialCharacters" gorm:"not null;default:0" faker:"boundary_start=1, boundary_end=20"`
	Numbers           int             `json:"numbers" gorm:"not null;default:0" faker:"boundary_start=1, boundary_end=20"`
	Scores            []Score         `json:"-" faker:"-"`
	Games             []Game          `json:"-" faker:"-"`
}

package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Score struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" faker:"-"`
	CreatedAt      time.Time       `json:"createdAt" faker:"-"`
	UpdatedAt      time.Time       `json:"updatedAt" faker:"-"`
	DeletedAt      *gorm.DeletedAt `json:"deletedAt" gorm:"index" faker:"-"`
	WordsPerMinute float64         `json:"wordsPerMinute" gorm:"type:DECIMAL GENERATED ALWAYS AS (words_typed::DECIMAL * 60.0 / time_elapsed) STORED" faker:"-"`
	WordsTyped     int             `json:"wordsTyped" faker:"boundary_start=50, boundary_end=1000"`
	TimeElapsed    float64         `json:"timeElapsed" faker:"oneof: 60.0, 120.0, 180.0"`
	Accuracy       float64         `json:"accuracy" gorm:"type:DECIMAL GENERATED ALWAYS AS (100.0 - (number_errors::DECIMAL * 100.0 / words_typed::DECIMAL)) STORED" faker:"-"`
	NumberErrors   int             `json:"numberErrors" faker:"-"`
	Errors         ErrorsJSON      `json:"errors" gorm:"type:jsonb" faker:"-"`
	UserId         uuid.UUID       `json:"userId" gorm:"not null" faker:"-"`
	TextId         uuid.UUID       `json:"textId" gorm:"not null" faker:"-"`
	GameId         uuid.UUID       `json:"gameId" faker:"-"`
}

type ErrorsJSON map[string]int

func (j ErrorsJSON) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ErrorsJSON) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &j)
}

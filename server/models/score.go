package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Score struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	WordsPerMinute float64         `json:"wordsPerMinute" gorm:"type:DECIMAL GENERATED ALWAYS AS (words_typed::DECIMAL * 60.0 / time_elapsed) STORED"`
	WordsTyped     int             `json:"wordsTyped"`
	TimeElapsed    float64         `json:"timeElapsed"`
	Accuracy       float64         `json:"accuracy" gorm:"type:DECIMAL GENERATED ALWAYS AS (100.0 - (number_errors::DECIMAL * 100.0 / words_typed::DECIMAL)) STORED"`
	NumberErrors   int             `json:"numberErrors"`
	Errors         ErrorsJSON      `json:"errors" gorm:"type:jsonb"`
	UserId         uuid.UUID       `json:"userId" gorm:"not null"`
	TextId         uuid.UUID       `json:"textId" gorm:"not null"`
	GameId         uuid.UUID       `json:"gameId"`
}

type ErrorsJSON map[string]int

func (j ErrorsJSON) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ErrorsJSON) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &j)
}

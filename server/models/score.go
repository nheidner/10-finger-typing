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

type CreateScoreInput struct {
	WordsTyped  int        `json:"wordsTyped" binding:"required" faker:"boundary_start=50, boundary_end=1000"`
	TimeElapsed float64    `json:"timeElapsed" binding:"required" faker:"oneof: 60.0, 120.0, 180.0"`
	Errors      ErrorsJSON `json:"errors" binding:"required,typingerrors"`
	TextId      uuid.UUID  `json:"textId" binding:"required"`
	UserId      uuid.UUID  `json:"-"`
	GameId      uuid.UUID  `json:"-"`
}

type FindScoresQuery struct {
	SortOptions []SortOption
	UserId      uuid.UUID
	GameId      uuid.UUID
	Username    string `form:"username"`
}

type FindScoresSortOption struct {
	Column string `validate:"required,oneof=accuracy errors created_at"`
	Order  string `validate:"required,oneof=desc asc"`
}

type ScoreService struct {
	DB *gorm.DB
}

type ErrorsJSON map[string]int

func (j ErrorsJSON) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ErrorsJSON) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &j)
}

// func (ss *ScoreService) Create(input CreateScoreInput) (*Score, error) {
// 	numberErrors := 0
// 	for _, value := range input.Errors {
// 		numberErrors += value
// 	}

// 	score := Score{
// 		WordsTyped:   input.WordsTyped,
// 		TimeElapsed:  input.TimeElapsed,
// 		Errors:       input.Errors,
// 		UserId:       input.UserId,
// 		GameId:       input.GameId,
// 		NumberErrors: numberErrors,
// 		TextId:       input.TextId,
// 	}

// 	omitFields := []string{"WordsPerMinute", "Accuracy"}

// 	if score.GameId == uuid.Nil {
// 		omitFields = append(omitFields, "GameId")
// 	}

// 	createResult := ss.DB.Omit(omitFields...).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "words_per_minute"}, {Name: "words_typed"}, {Name: "time_elapsed"}, {Name: "accuracy"}, {Name: "number_errors"}, {Name: "errors"}}}).
// 		Create(&score)

// 	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
// 		return nil, createResult.Error
// 	}

// 	return &score, nil
// }

// func (ss *ScoreService) DeleteAll() error {
// 	return ss.DB.Exec("TRUNCATE scores RESTART IDENTITY CASCADE").Error
// }

package models

import (
	custom_errors "10-typing/errors"
	"net/http"

	"gorm.io/gorm"
)

type Score struct {
	*gorm.Model
	ID             uint           `json:"id" gorm:"primary_key"`
	WordsPerMinute float64        `json:"wordsPerMinute" gorm:"-"`
	WordsTyped     int            `json:"wordsTyped"`
	TimeElapsed    float64        `json:"timeElapsed"`
	Accuracy       float64        `json:"accuracy" gorm:"-"`
	NumberErrors   int            `json:"numberErrors"`
	Errors         map[string]int `json:"errors" gorm:"type:jsonb"`
	UserId         uint           `json:"userId"`
}

type CreateScoreInput struct {
	WordsTyped  int            `json:"wordsTyped" binding:"required"`
	TimeElapsed float64        `json:"timeElapsed" binding:"required"`
	Errors      map[string]int `json:"errors" binding:"required,typingerrors"`
	UserId      uint           `json:"userId" binding:"required"`
}

type ScoreService struct {
	DB *gorm.DB
}

func (ss *ScoreService) Create(input CreateScoreInput) (*Score, error) {
	score := Score{
		WordsTyped:  input.WordsTyped,
		TimeElapsed: input.TimeElapsed,
		Errors:      input.Errors,
		UserId:      input.UserId,
	}

	createResult := ss.DB.Create(&score)
	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}

	return &score, nil
}

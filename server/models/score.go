package models

import (
	custom_errors "10-typing/errors"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ErrorsJSON map[string]int

type Score struct {
	ID             uint            `json:"id" gorm:"primary_key"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	WordsPerMinute float64         `json:"wordsPerMinute" gorm:"type:DECIMAL GENERATED ALWAYS AS (words_typed::DECIMAL * 60.0 / time_elapsed) STORED"`
	WordsTyped     int             `json:"wordsTyped"`
	TimeElapsed    float64         `json:"timeElapsed"`
	Accuracy       float64         `json:"accuracy" gorm:"type:DECIMAL GENERATED ALWAYS AS (100.0 - (number_errors::DECIMAL * 100.0 / words_typed::DECIMAL)) STORED"`
	NumberErrors   int             `json:"numberErrors"`
	Errors         ErrorsJSON      `json:"errors" gorm:"type:jsonb"`
	UserId         uint            `json:"userId"`
}

type CreateScoreInput struct {
	WordsTyped  int        `json:"wordsTyped" binding:"required"`
	TimeElapsed float64    `json:"timeElapsed" binding:"required"`
	Errors      ErrorsJSON `json:"errors" binding:"required,typingerrors"`
}

type FindScoresQuery struct {
	SortOptions []SortOption
	UserId      uint
	Username    string `form:"username"`
}

type FindScoresSortOption struct {
	Column string `validate:"required,oneof=accuracy errors created_at"`
	Order  string `validate:"required,oneof=desc asc"`
}

type ScoreService struct {
	DB *gorm.DB
}

func (j ErrorsJSON) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *ErrorsJSON) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

func (ss *ScoreService) FindScores(query FindScoresQuery) (*[]Score, error) {
	var scores []Score

	findScoresDbQuery := ss.DB
	if query.UserId != 0 {
		findScoresDbQuery = findScoresDbQuery.Where("user_id = ?", query.UserId)
	}
	if query.Username != "" {
		findScoresDbQuery = findScoresDbQuery.Joins("INNER JOIN users ON scores.user_id = users.id").Where("users.username = ?", query.Username)
	}

	for _, sortOption := range query.SortOptions {
		findScoresDbQuery = findScoresDbQuery.Order(clause.OrderByColumn{Column: clause.Column{Name: sortOption.Column}, Desc: sortOption.Order == "desc"})
	}
	if len(query.SortOptions) == 0 {
		findScoresDbQuery = findScoresDbQuery.Order("created_at desc")
	}

	findScoresDbQuery.Find(&scores)

	if findScoresDbQuery.Error != nil {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}

	return &scores, nil
}

func (ss *ScoreService) Create(userId uint, input CreateScoreInput) (*Score, error) {
	numberErrors := 0
	for _, value := range input.Errors {
		numberErrors += value
	}

	score := Score{
		WordsTyped:   input.WordsTyped,
		TimeElapsed:  input.TimeElapsed,
		Errors:       input.Errors,
		UserId:       userId,
		NumberErrors: numberErrors,
	}

	createResult := ss.DB.Omit("WordsPerMinute", "Accuracy").
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "words_per_minute"}, {Name: "words_typed"}, {Name: "time_elapsed"}, {Name: "accuracy"}, {Name: "number_errors"}, {Name: "errors"}}}).
		Create(&score)

	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}

	return &score, nil
}

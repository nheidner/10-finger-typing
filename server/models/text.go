package models

import (
	custom_errors "10-typing/errors"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Text struct {
	ID                uint            `json:"id" gorm:"primary_key"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	DeletedAt         *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Language          string          `json:"language" gorm:"not null;type:varchar(255)"`
	Text              string          `json:"text" gorm:"not null;type:text"`
	Punctuation       bool            `json:"punctuation" gorm:"not null;default:false"`
	SpecialCharacters int             `json:"specialCharacters" gorm:"not null;default:0"`
	Numbers           int             `json:"numbers" gorm:"not null;default:0"`
	Scores            []Score         `json:"-"`
}

type FindTextQuery struct {
	Language             string `form:"language" binding:"required"`
	Punctuation          bool   `form:"punctuation"`
	SpecialCharactersGte int    `form:"specialCharacters[gte]"`
	SpecialCharactersLte int    `form:"specialCharacters[lte]"`
	NumbersGte           int    `form:"numbers[gte]"`
	NumbersLte           int    `form:"numbers[lte]"`
}

type TextService struct {
	DB *gorm.DB
}

func (ts TextService) FindNewOneByUserId(userId uint, query FindTextQuery) (*Text, error) {
	var text Text

	result := ts.DB.
		Joins("LEFT JOIN scores s1 ON texts.id = s1.text_id").
		Joins("LEFT JOIN scores s2 ON s1.text_id = s2.text_id AND s2.user_id = ?", userId).
		Where("s2.text_id IS NULL").
		Where("language = ?", query.Language).
		Where("punctuation = ?", query.Punctuation)

	if query.SpecialCharactersGte != 0 {
		result = result.Where("special_characters >= ?", query.SpecialCharactersGte)
	}
	if query.SpecialCharactersLte != 0 {
		result = result.Where("special_characters <= ?", query.SpecialCharactersLte)
	}
	if query.NumbersGte != 0 {
		result = result.Where("numbers >= ?", query.NumbersGte)
	}
	if query.NumbersLte != 0 {
		result = result.Where("numbers <= ?", query.NumbersLte)
	}

	result.First(&text)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		internalServerError := custom_errors.HTTPError{Message: "error querying text", Status: http.StatusInternalServerError, Details: result.Error.Error()}
		return nil, internalServerError
	}

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &text, nil
}

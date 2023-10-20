package models

import (
	custom_errors "10-typing/errors"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

type FindTextQuery struct {
	Language             string `form:"language" binding:"required"`
	Punctuation          bool   `form:"punctuation"`
	SpecialCharactersGte int    `form:"specialCharacters[gte]"`
	SpecialCharactersLte int    `form:"specialCharacters[lte]"`
	NumbersGte           int    `form:"numbers[gte]"`
	NumbersLte           int    `form:"numbers[lte]"`
}

type CreateTextInput struct {
	Language          string `json:"language" binding:"required" faker:"oneof: de en fr"`
	Punctuation       bool   `json:"punctuation"`
	SpecialCharacters int    `json:"specialCharacters" faker:"boundary_start=1, boundary_end=20"`
	Numbers           int    `json:"numbers" faker:"boundary_start=1, boundary_end=20"`
}

type TextService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func (ti *CreateTextInput) String() string {
	return fmt.Sprintf("language: %s, punctuation: %t, number of special characters: %d, number of numbers: %d, length: 100 words", ti.Language, ti.Punctuation, ti.SpecialCharacters, ti.Numbers)
}

func (ts TextService) FindNewOneByUserId(userId uuid.UUID, query FindTextQuery) (*Text, error) {
	var text Text

	result := ts.DB.
		Joins("LEFT JOIN scores s1 ON texts.id = s1.text_id").
		Joins("LEFT JOIN scores s2 ON s1.text_id = s2.text_id AND s2.user_id = ?", userId).
		Where("s2.text_id IS NULL").
		Where("language = ?", query.Language).
		Where("punctuation = ?", query.Punctuation).
		Order("created_at DESC")

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

func (ts TextService) Create(ctx context.Context, input CreateTextInput, gptText string) (*Text, error) {
	text := Text{
		Language:          input.Language,
		Text:              gptText,
		Punctuation:       input.Punctuation,
		SpecialCharacters: input.SpecialCharacters,
		Numbers:           input.Numbers,
	}

	createResult := ts.DB.Create(&text)
	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		return nil, createResult.Error
	}

	if err := ts.CreateInRedis(ctx, text.ID); err != nil {
		return nil, err
	}

	return &text, nil
}

func (ts *TextService) CreateInRedis(ctx context.Context, textId uuid.UUID) error {
	textIdsKey := getTextIdsKey()

	r, err := ts.RDB.Exists(ctx, textIdsKey).Result()
	if err != nil {
		return err
	}
	if r != 0 {
		return ts.RDB.SAdd(ctx, textIdsKey, textId.String()).Err()
	}

	// if no text id key exists, new one with all text ids from DB is created
	allTexts, err := ts.GetAllTexts()
	if err != nil {
		return err
	}

	allTextIds := make([]interface{}, 0, len(allTexts)+1)
	for _, text := range allTexts {
		allTextIds = append(allTextIds, text.ID.String())
	}

	allTextIds = append(allTextIds, textId.String())

	return ts.RDB.SAdd(ctx, textIdsKey, allTextIds...).Err()
}

func (tx *TextService) GetAllTexts() ([]Text, error) {
	var texts []Text

	if err := tx.DB.Select("id").Find(&texts).Error; err != nil {
		return nil, err
	}

	return texts, nil
}

func (ts *TextService) TextExists(ctx context.Context, textId uuid.UUID) (bool, error) {
	textIdsKey := getTextIdsKey()

	r, err := ts.RDB.SMIsMember(ctx, textIdsKey, textId.String()).Result()
	if err != nil {
		return false, err
	}

	return r[0], nil
}

func (ts *TextService) DeleteAll() error {
	textIdsKey := getTextIdsKey()
	if err := ts.RDB.Del(context.Background(), textIdsKey).Err(); err != nil {
		return err
	}

	return ts.DB.Exec("TRUNCATE texts RESTART IDENTITY CASCADE").Error
}

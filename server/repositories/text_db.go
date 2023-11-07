package repositories

import (
	"10-typing/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TextDbRepository struct {
	db *gorm.DB
}

func NewTextDbRepository(db *gorm.DB) *TextDbRepository {
	return &TextDbRepository{db}
}

// returns nil for *models.Text and nil for error when no record could be found
func (tr *TextDbRepository) FindNewTextByUserId(
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	result := tr.db.
		Joins("LEFT JOIN scores s1 ON texts.id = s1.text_id").
		Joins("LEFT JOIN scores s2 ON s1.text_id = s2.text_id AND s2.user_id = ?", userId).
		Where("s2.text_id IS NULL").
		Where("language = ?", language).
		Where("punctuation = ?", punctuation).
		Order("created_at DESC")

	if specialCharactersGte != 0 {
		result = result.Where("special_characters >= ?", specialCharactersGte)
	}
	if specialCharactersLte != 0 {
		result = result.Where("special_characters <= ?", specialCharactersLte)
	}
	if numbersGte != 0 {
		result = result.Where("numbers >= ?", numbersGte)
	}
	if numbersLte != 0 {
		result = result.Where("numbers <= ?", numbersLte)
	}

	var text models.Text

	result.First(&text)

	if result.Error != nil {
		return nil, errors.New("error querying text: " + result.Error.Error())
	}

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &text, nil
}

func (tr *TextDbRepository) Create(text models.Text) (*models.Text, error) {
	createResult := tr.db.Create(&text)
	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		return nil, createResult.Error
	}

	return &text, nil
}

func (tr *TextDbRepository) GetAllTextIds() ([]uuid.UUID, error) {
	var textIds []uuid.UUID

	if err := tr.db.Model(&models.Text{}).Pluck("id", &textIds).Error; err != nil {
		return nil, err
	}

	return textIds, nil
}

func (tr *TextDbRepository) DeleteAll() error {
	return tr.db.Exec("TRUNCATE texts RESTART IDENTITY CASCADE").Error
}

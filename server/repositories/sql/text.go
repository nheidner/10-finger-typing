package sql_repo

import (
	"10-typing/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// returns nil for *models.Text and nil for error when no record could be found
func (repo *SQLRepository) FindNewTextByUserId(
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	result := repo.db.
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

func (repo *SQLRepository) CreateText(text models.Text) (*models.Text, error) {
	createResult := repo.db.Create(&text)
	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		return nil, createResult.Error
	}

	return &text, nil
}

func (repo *SQLRepository) FindAllTextIds() ([]uuid.UUID, error) {
	var textIds []uuid.UUID

	if err := repo.db.Model(&models.Text{}).Pluck("id", &textIds).Error; err != nil {
		return nil, err
	}

	return textIds, nil
}

func (repo *SQLRepository) DeleteAllTexts() error {
	return repo.db.Exec("TRUNCATE texts RESTART IDENTITY CASCADE").Error
}

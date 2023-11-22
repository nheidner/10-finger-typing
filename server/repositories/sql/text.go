package sql_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// returns nil for *models.Text and nil for error when no record could be found
func (repo *SQLRepository) FindNewTextForUser(
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindNewTextForUser"

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

	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, errors.E(op, repositories.ErrNotFound)
	case result.Error != nil:
		return nil, errors.E(op, result.Error)
	}

	return &text, nil
}

func (repo *SQLRepository) FindAllTextIds() ([]uuid.UUID, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindAllTextIds"
	var textIds []uuid.UUID

	result := repo.db.Model(&models.Text{}).Pluck("id", &textIds)

	if result.Error != nil {
		return nil, errors.E(op, result.Error)
	}

	return textIds, nil
}

func (repo *SQLRepository) FindTextById(textId uuid.UUID) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindTextById"
	var text = models.Text{
		ID: textId,
	}

	if err := repo.db.First(&text).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, repositories.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &text, nil
}

func (repo *SQLRepository) CreateText(text models.Text) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateText"

	if err := repo.db.Create(&text).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &text, nil
}

func (repo *SQLRepository) DeleteAllTexts() error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllTexts"

	if err := repo.db.Exec("TRUNCATE texts RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

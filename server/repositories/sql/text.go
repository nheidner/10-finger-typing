package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"

	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// returns nil for *models.Text and nil for error when no record could be found
func (repo *SQLRepository) FindNewTextForUser(ctx context.Context,
	tx common.Transaction,
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindNewTextForUser"
	db := repo.dbConn(tx)

	result := db.WithContext(ctx).
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
		return nil, errors.E(op, common.ErrNotFound)
	case result.Error != nil:
		return nil, errors.E(op, result.Error)
	}

	return &text, nil
}

func (repo *SQLRepository) FindAllTextIds(ctx context.Context, tx common.Transaction) ([]uuid.UUID, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindAllTextIds"
	db := repo.dbConn(tx)
	var textIds []uuid.UUID

	result := db.WithContext(ctx).Model(&models.Text{}).Pluck("id", &textIds)

	if result.Error != nil {
		return nil, errors.E(op, result.Error)
	}

	return textIds, nil
}

func (repo *SQLRepository) FindTextById(ctx context.Context, tx common.Transaction, textId uuid.UUID) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindTextById"
	db := repo.dbConn(tx)
	var text = models.Text{
		ID: textId,
	}

	if err := db.WithContext(ctx).First(&text).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &text, nil
}

func (repo *SQLRepository) CreateTextAndCache(ctx context.Context, tx common.Transaction, cacheRepo common.CacheRepository, text models.Text) (*models.Text, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateText"
	db := repo.dbConn(tx)

	if err := db.WithContext(ctx).Create(&text).Error; err != nil {
		return nil, errors.E(op, err)
	}

	// create text in redis and additionally: if no text ids key exists, query all text ids from DB and write them to text ids key
	allTextsAreInRedis, err := cacheRepo.TextIdsKeyExists(ctx)
	switch {
	case err != nil:
		return nil, errors.E(op, err)
	case !allTextsAreInRedis:
		allTextIds, err := repo.FindAllTextIds(ctx, tx)
		if err != nil {
			return nil, errors.E(op, err)
		}

		allTextIds = append(allTextIds, text.ID)

		if err = cacheRepo.SetTextId(ctx, nil, allTextIds...); err != nil {
			return nil, errors.E(op, err)
		}
	default:
		if err = cacheRepo.SetTextId(ctx, nil, text.ID); err != nil {
			return nil, errors.E(op, err)
		}
	}

	return &text, nil
}

func (repo *SQLRepository) DeleteAllTexts(ctx context.Context, tx common.Transaction) error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllTexts"
	db := repo.dbConn(tx)

	if err := db.WithContext(ctx).Exec("TRUNCATE texts RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

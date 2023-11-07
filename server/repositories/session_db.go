package repositories

import (
	"10-typing/models"
	"errors"

	"gorm.io/gorm"
)

type SessionDbRepository struct {
	db *gorm.DB
}

func NewSessionDbRepository(db *gorm.DB) *SessionDbRepository {
	return &SessionDbRepository{db}
}

func (sr *SessionDbRepository) Create(newSession models.Session) (*models.Session, error) {
	result := sr.db.Create(&newSession)
	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, errors.New("new sessions found")
	}

	return &newSession, nil
}

func (sr *SessionDbRepository) DeleteByTokenHash(tokenHash string) error {
	result := sr.db.Delete(&models.Session{}, "token_hash = ?", tokenHash)
	switch {
	case result.Error != nil:
		return result.Error
	case result.RowsAffected == 0:
		return errors.New("not found")
	}

	return nil
}

func (sr *SessionDbRepository) DeleteAll() error {
	return sr.db.Exec("TRUNCATE sessions RESTART IDENTITY CASCADE").Error
}

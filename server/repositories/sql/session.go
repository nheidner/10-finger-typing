package sql_repo

import (
	"10-typing/models"
	"errors"
)

func (repo *SQLRepository) CreateSession(newSession models.Session) (*models.Session, error) {
	result := repo.db.Create(&newSession)
	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, errors.New("new sessions found")
	}

	return &newSession, nil
}

func (repo *SQLRepository) DeleteSessionByTokenHash(tokenHash string) error {
	result := repo.db.Delete(&models.Session{}, "token_hash = ?", tokenHash)
	switch {
	case result.Error != nil:
		return result.Error
	case result.RowsAffected == 0:
		return errors.New("not found")
	}

	return nil
}

func (repo *SQLRepository) DeleteAllSessions() error {
	return repo.db.Exec("TRUNCATE sessions RESTART IDENTITY CASCADE").Error
}

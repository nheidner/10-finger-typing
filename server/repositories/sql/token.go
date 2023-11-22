package sql_repo

import (
	"10-typing/errors"
	"10-typing/models"

	"github.com/google/uuid"
)

func (repo *SQLRepository) CreateToken(roomId uuid.UUID) (*models.Token, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateToken"

	token := models.Token{
		RoomID: roomId,
	}

	if err := repo.db.Create(&token).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &token, nil
}

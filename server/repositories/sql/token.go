package sql_repo

import (
	"10-typing/models"

	"github.com/google/uuid"
)

func (repo *SQLRepository) CreateToken(roomId uuid.UUID) (*models.Token, error) {
	token := models.Token{
		RoomID: roomId,
	}

	if err := repo.db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

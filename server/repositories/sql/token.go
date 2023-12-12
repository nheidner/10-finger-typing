package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
)

func (repo *SQLRepository) CreateToken(ctx context.Context, tx common.Transaction, roomId uuid.UUID) (*models.Token, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateToken"
	db := repo.dbConn(tx)

	token := models.Token{
		RoomID: roomId,
	}

	if err := db.WithContext(ctx).Create(&token).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &token, nil
}

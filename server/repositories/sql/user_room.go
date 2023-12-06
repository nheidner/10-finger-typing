package sql_repo

import (
	"10-typing/errors"
	"context"

	"github.com/google/uuid"
)

func (repo *SQLRepository) CreateUserRoom(ctx context.Context, userId, roomId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.CreateUserRoom"
	join := map[string]any{"room_id": roomId, "user_id": userId}

	if err := repo.db.WithContext(ctx).Table("user_rooms").Create(&join).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

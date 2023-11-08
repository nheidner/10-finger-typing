package sql_repo

import (
	"github.com/google/uuid"
)

func (repo *SQLRepository) CreateUserRoom(userId, roomId uuid.UUID) error {
	join := map[string]any{"room_id": roomId, "user_id": userId}

	return repo.db.Table("user_rooms").Create(&join).Error
}

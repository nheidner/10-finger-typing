package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoomDbRepository struct {
	db *gorm.DB
}

func NewUserRoomDbRepository(db *gorm.DB) *UserRoomDbRepository {
	return &UserRoomDbRepository{db}
}

func (ur *UserRoomDbRepository) Create(userId, roomId uuid.UUID) error {
	join := map[string]any{"room_id": roomId, "user_id": userId}

	return ur.db.Table("user_rooms").Create(&join).Error
}

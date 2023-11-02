package repositories

import (
	"10-typing/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomDbRepository struct {
	db *gorm.DB
}

func NewRoomDbRepository(db *gorm.DB) *RoomDbRepository {
	return &RoomDbRepository{db}
}

func (rr *RoomDbRepository) FindInDb(roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	var room models.Room
	result := rr.db.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.id = ?", roomId).
		Where("ur.user_id = ?", userId).
		Find(&room)
	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, fmt.Errorf("no room found")
	}

	if err := rr.db.Model(room).Association("Users").Find(&(room.Users)); err != nil {
		return nil, err
	}

	return &room, nil
}

func (rr *RoomDbRepository) FindRoomWithUsers(roomId uuid.UUID) (*models.Room, error) {
	var room = models.Room{
		ID: roomId,
	}

	if err := rr.db.Preload("Users").Find(&room).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func (rr *RoomDbRepository) Create(newRoom *models.Room) error {
	return rr.db.Create(newRoom).Error
}

func (rr *RoomDbRepository) SoftDeleteRoomFromDB(roomId uuid.UUID) error {
	if err := rr.db.Delete(&models.Room{}, roomId).Error; err != nil {
		return err
	}
	if err := rr.db.Delete(&models.Game{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}
	if err := rr.db.Delete(&models.Token{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}

	return rr.db.Table("user_rooms").Where("room_id = ?", roomId).Delete(nil).Error
}

func (rr *RoomDbRepository) DeleteAll() error {
	return rr.db.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

package sql_repo

import (
	"10-typing/models"
	"fmt"

	"github.com/google/uuid"
)

func (repo *SQLRepository) FindRoomByUser(roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	var room models.Room
	result := repo.db.
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

	if err := repo.db.Model(&room).Association("Users").Find(&(room.Users)); err != nil {
		return nil, err
	}

	return &room, nil
}

func (repo *SQLRepository) FindRoom(roomId uuid.UUID) (*models.Room, error) {
	var room = models.Room{
		ID: roomId,
	}

	if err := repo.db.Preload("Users").Find(&room).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func (repo *SQLRepository) CreateRoom(newRoom models.Room) (*models.Room, error) {
	if err := repo.db.Create(&newRoom).Error; err != nil {
		return nil, err
	}

	return &newRoom, nil
}

func (repo *SQLRepository) SoftDeleteRoom(roomId uuid.UUID) error {
	if err := repo.db.Delete(&models.Room{}, roomId).Error; err != nil {
		return err
	}
	if err := repo.db.Delete(&models.Game{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}
	if err := repo.db.Delete(&models.Token{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}

	return repo.db.Table("user_rooms").Where("room_id = ?", roomId).Delete(nil).Error
}

func (repo *SQLRepository) DeleteAllRooms() error {
	return repo.db.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

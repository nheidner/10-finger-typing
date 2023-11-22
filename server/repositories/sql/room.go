package sql_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *SQLRepository) FindRoomWithUsers(roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindRoomByUser"

	var room = models.Room{
		ID: roomId,
	}
	if err := repo.db.First(&room).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, repositories.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	if err := repo.db.Model(&room).Association("Users").Find(&(room.Users)); err != nil {
		return nil, errors.E(op, err)
	}

	return &room, nil
}

func (repo *SQLRepository) FindRoom(roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindRoom"

	var room = models.Room{
		ID: roomId,
	}

	if err := repo.db.Preload("Users").First(&room).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, repositories.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &room, nil
}

func (repo *SQLRepository) CreateRoom(newRoom models.Room) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateRoom"

	if err := repo.db.Create(&newRoom).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &newRoom, nil
}

func (repo *SQLRepository) SoftDeleteRoom(roomId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.SoftDeleteRoom"

	if err := repo.db.Delete(&models.Room{}, roomId).Error; err != nil {
		return errors.E(op, err)
	}
	if err := repo.db.Delete(&models.Game{}, "room_id = ?", roomId).Error; err != nil {
		return errors.E(op, err)
	}
	if err := repo.db.Delete(&models.Token{}, "room_id = ?", roomId).Error; err != nil {
		return errors.E(op, err)
	}

	if err := repo.db.Table("user_rooms").Where("room_id = ?", roomId).Delete(nil).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) DeleteAllRooms() error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllRooms"

	if err := repo.db.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

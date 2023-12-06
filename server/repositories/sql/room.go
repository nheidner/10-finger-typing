package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *SQLRepository) FindRoomWithUsers(ctx context.Context, roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindRoomByUser"

	var room = models.Room{
		ID: roomId,
	}
	if err := repo.db.WithContext(ctx).First(&room).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &room, nil
}

func (repo *SQLRepository) FindRoom(ctx context.Context, roomId uuid.UUID) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindRoom"

	var room = models.Room{
		ID: roomId,
	}

	if err := repo.db.WithContext(ctx).Preload("Users").First(&room).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &room, nil
}

func (repo *SQLRepository) CreateRoom(ctx context.Context, newRoom models.Room) (*models.Room, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateRoom"

	if err := repo.db.WithContext(ctx).Create(&newRoom).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &newRoom, nil
}

func (repo *SQLRepository) SoftDeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.SoftDeleteRoom"

	if err := repo.db.WithContext(ctx).Delete(&models.Room{}, roomId).Error; err != nil {
		return errors.E(op, err)
	}
	if err := repo.db.WithContext(ctx).Delete(&models.Game{}, "room_id = ?", roomId).Error; err != nil {
		return errors.E(op, err)
	}
	if err := repo.db.WithContext(ctx).Delete(&models.Token{}, "room_id = ?", roomId).Error; err != nil {
		return errors.E(op, err)
	}

	if err := repo.db.WithContext(ctx).Table("user_rooms").Where("room_id = ?", roomId).Delete(nil).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) DeleteAllRooms(ctx context.Context) error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllRooms"

	if err := repo.db.WithContext(ctx).Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomService struct {
	roomDbRepo    *repositories.RoomDbRepository
	roomRedisRepo *repositories.RoomRedisRepository
	userRoomDb    *repositories.UserRoomDbRepository
}

func NewRoomService(
	roomDbRepo *repositories.RoomDbRepository,
	roomRedisRepo *repositories.RoomRedisRepository,
	userRoomDb *repositories.UserRoomDbRepository,
) *RoomService {
	return &RoomService{roomDbRepo, roomRedisRepo, userRoomDb}
}

func (rs *RoomService) Find(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	room, err := rs.roomRedisRepo.FindInRedis(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	if room == nil {
		room, err = rs.roomDbRepo.FindInDb(roomId, userId)
		if err != nil {
			return nil, err
		}

		if err = rs.roomRedisRepo.CreateRoomInRedis(ctx, *room); err != nil {
			// no error should be returned
			log.Println(err)
		}
	}

	return room, nil
}

func (rs *RoomService) Create(tx *gorm.DB, userIds []uuid.UUID, emails []string, adminId uuid.UUID) (*models.Room, error) {
	var newRoom = &models.Room{
		AdminId: adminId,
	}

	if err := rs.roomDbRepo.Create(newRoom); err != nil {
		return nil, err
	}

	// room subscribers
	for _, userId := range userIds {
		if err := rs.userRoomDb.Create(userId, newRoom.ID); err != nil {
			return nil, err
		}
	}

	newRoom, err := rs.roomDbRepo.FindRoomWithUsers(newRoom.ID)
	if err != nil {
		return nil, err
	}

	if err := rs.roomRedisRepo.CreateRoomInRedis(context.Background(), *newRoom); err != nil {
		return nil, err
	}

	return newRoom, nil
}

func returnAndRollBackIfNeeded(tx *gorm.DB, err error) (*models.Room, error) {
	if tx == nil {
		tx.Rollback()
	}

	return nil, err
}

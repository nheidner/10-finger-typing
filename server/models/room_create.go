package models

import (
	"context"
	"log"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateRoomInput struct {
	UserIds []uuid.UUID `json:"userIds"`
	Emails  []string    `json:"emails" binding:"dive,email"`
}

func (rs *RoomService) Create(tx *gorm.DB, input CreateRoomInput, adminId uuid.UUID) (*Room, error) {
	db := tx
	if db == nil {
		db = rs.DB.Begin()
	}

	var room = Room{
		AdminId: adminId,
	}
	if err := db.Create(&room).Error; err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	// subscribers
	for _, userId := range input.UserIds {
		join := map[string]any{"room_id": room.ID, "user_id": userId}

		if err := db.Table("user_rooms").Create(&join).Error; err != nil {
			return returnAndRollBackIfNeeded(tx, err)
		}
	}

	if tx == nil {
		db.Commit()
	}

	if err := db.Preload("Users").Find(&room).Error; err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	if err := rs.createRoomInRedis(context.Background(), &room); err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	return &room, nil
}

func (rs *RoomService) createRoomInRedis(ctx context.Context, room *Room) error {
	// add room
	roomKey := getRoomKey(room.ID)
	roomValue := map[string]any{
		roomAdminIdField:   room.AdminId.String(),
		roomCreatedAtField: room.CreatedAt.UnixMilli(),
		roomUpdatedAtField: room.UpdatedAt.UnixMilli(),
	}
	if err := rs.RDB.HSet(ctx, roomKey, roomValue).Err(); err != nil {
		return err
	}

	// add room subscriber ids
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(room.ID)
	roomSubscriberIdsValue := make([]string, 0, len(room.Users))
	for _, subscriber := range room.Users {
		roomSubscriberIdsValue = append(roomSubscriberIdsValue, subscriber.ID.String())
	}

	if err := rs.RDB.SAdd(ctx, roomSubscriberIdsKey, roomSubscriberIdsValue).Err(); err != nil {
		return err
	}

	// add room subscribers
	for _, subscriber := range room.Users {
		roomSubscriberKey := getRoomSubscriberKey(room.ID, subscriber.ID)
		roomSubscriberValue := map[string]any{
			roomSubscriberUsernameField:   subscriber.Username,
			roomSubscriberStatusField:     strconv.Itoa(int(NilSubscriberStatus)),
			roomSubscriberGameStatusField: strconv.Itoa(int(NilSubscriberGameStatus)),
		}

		if err := rs.RDB.HSet(ctx, roomSubscriberKey, roomSubscriberValue).Err(); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func returnAndRollBackIfNeeded(tx *gorm.DB, err error) (*Room, error) {
	if tx == nil {
		tx.Rollback()
	}

	return nil, err
}

func (rs *RoomService) DeleteAll() error {
	return rs.DB.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

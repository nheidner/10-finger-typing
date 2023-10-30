package models

import (
	"context"

	"github.com/google/uuid"
)

func (rs *RoomService) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	if err := rs.softDeleteRoomFromDB(roomId); err != nil {
		return err
	}

	return rs.deleteRoomFromRedis(ctx, roomId)
}

func (rs *RoomService) softDeleteRoomFromDB(roomId uuid.UUID) error {
	if err := rs.DB.Delete(&Room{}, roomId).Error; err != nil {
		return err
	}
	if err := rs.DB.Delete(&Game{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}
	if err := rs.DB.Delete(&Token{}, "room_id = ?", roomId).Error; err != nil {
		return err
	}
	if err := rs.DB.Table("user_rooms").Where("room_id = ?", roomId).Delete(nil).Error; err != nil {
		return err
	}

	return nil
}

func (rs *RoomService) deleteRoomFromRedis(ctx context.Context, roomId uuid.UUID) error {
	roomKey := getRoomKey(roomId)

	// first need to send terminate action message so that all websocket that remained connected, disconnect
	if err := rs.PublishAction(ctx, roomId, TerminateAction); err != nil {
		return err
	}

	pattern := roomKey + "*"
	iter := rs.RDB.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		rs.RDB.Del(ctx, key)
	}

	return iter.Err()
}

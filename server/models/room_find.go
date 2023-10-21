package models

import (
	"10-typing/utils"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (rs *RoomService) Find(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*Room, error) {
	room, err := rs.findRoomFromCache(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	if room == nil {
		room, err = rs.findRoomFromDB(roomId, userId)
		if err != nil {
			return nil, err
		}

		if err = rs.storeRoomToCache(ctx, room); err != nil {
			// no error should be returned
			log.Println(err)
		}
	}

	return room, nil
}

func (rs *RoomService) findRoomFromDB(roomId uuid.UUID, userId uuid.UUID) (*Room, error) {
	var room Room
	result := rs.DB.
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

	if err := rs.DB.Model(room).Association("Subscribers").Find(&(room.Subscribers)); err != nil {
		return nil, err
	}

	return &room, nil
}

func (rs *RoomService) findRoomFromCache(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*Room, error) {
	return rs.findInRedis(ctx, roomId, userId)
}

func (rs *RoomService) storeRoomToCache(ctx context.Context, room *Room) error {
	return rs.createInRedis(ctx, room)
}

func (rs *RoomService) findInRedis(ctx context.Context, roomId uuid.UUID, userId uuid.UUID) (*Room, error) {
	roomKey := getRoomKey(roomId)

	roomData, err := rs.RDB.HGetAll(ctx, roomKey).Result()
	switch {
	case err != nil:
		return nil, err
	case len(roomData) == 0:
		return nil, nil
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	roomSubscriberIds, err := rs.RDB.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	userIdStr := userId.String()
	if !utils.SliceContains[string](roomSubscriberIds, userIdStr) {
		return nil, fmt.Errorf("user is not subscribed to room")
	}

	roomSubscribers := make([]User, 0, len(roomSubscriberIds))
	for _, roomSubscriberIdStr := range roomSubscriberIds {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, err
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		roomSubscriber, err := rs.RDB.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		subscriber := User{
			ID:       roomSubscriberId,
			Username: roomSubscriber["username"],
		}

		roomSubscribers = append(roomSubscribers, subscriber)
	}

	createdAt, err := stringToTime(roomData["createdAt"])
	if err != nil {
		return nil, err
	}
	updatedAt, err := stringToTime(roomData["updatedAt"])
	if err != nil {
		return nil, err
	}

	return &Room{
		ID:          roomId,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Subscribers: roomSubscribers,
	}, nil
}

func stringToTime(data string) (time.Time, error) {
	intVal, err := strconv.Atoi(data)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(intVal), 0), nil
}

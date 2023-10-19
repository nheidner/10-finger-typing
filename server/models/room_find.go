package models

import (
	"10-typing/utils"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (rs *RoomService) Find(roomId uuid.UUID, userId uint) (*Room, error) {
	cachedRoom, err := rs.findInRedis(context.Background(), roomId, userId)
	if cachedRoom != nil {
		return cachedRoom, nil
	}
	if err != nil {
		return nil, err
	}

	var room Room
	result := rs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.id = ?", roomId).
		Where("ur.user_id = ?", userId).
		Find(&room)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no room found")
	}

	if err := rs.DB.Model(room).Association("Subscribers").Find(&(room.Subscribers)); err != nil {
		return nil, err
	}

	// no error needs to be returned
	rs.createInRedis(context.Background(), &room)

	return &room, nil
}

func (rs *RoomService) findInRedis(ctx context.Context, roomId uuid.UUID, userId uint) (*Room, error) {
	roomKey := getRoomKey(roomId)

	roomData, err := rs.RDB.HGetAll(ctx, roomKey).Result()
	if err != nil {
		return nil, err
	}
	if len(roomData) == 0 {
		return nil, nil
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	roomSubscriberIds, err := rs.RDB.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}
	if len(roomSubscriberIds) == 0 {
		return nil, nil
	}

	userIdStr := strconv.Itoa(int(userId))
	if !utils.SliceContains[string](roomSubscriberIds, userIdStr) {
		return nil, fmt.Errorf("user is not subscribed to room")
	}

	roomSubscribers := make([]User, 0, len(roomSubscriberIds))
	for _, roomSubscriberId := range roomSubscriberIds {
		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		roomSubscriber, err := rs.RDB.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}
		if len(roomSubscriber) == 0 {
			return nil, nil
		}

		roomSubscriberIdUint, err := strconv.Atoi(roomSubscriberId)
		if err != nil {
			return nil, err
		}

		subscriber := User{
			ID:       uint(roomSubscriberIdUint),
			Username: roomSubscriber["username"],
		}

		roomSubscribers = append(roomSubscribers, subscriber)
	}

	createdAtInt, err := strconv.Atoi(roomData["createdAt"])
	if err != nil {
		return nil, err
	}
	updatedAtInt, err := strconv.Atoi(roomData["createdAt"])
	if err != nil {
		return nil, err
	}

	return &Room{
		ID:          roomId,
		CreatedAt:   time.Unix(int64(createdAtInt), int64((createdAtInt%1000)*1e6)),
		UpdatedAt:   time.Unix(int64(updatedAtInt), int64((updatedAtInt%1000)*1e6)),
		Subscribers: roomSubscribers,
	}, nil
}

package models

import (
	custom_errors "10-typing/errors"
	"10-typing/utils"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (rs *RoomService) Find(roomId uuid.UUID, userId uint) (*Room, error) {
	room, _ := rs.findInRedis(context.Background(), roomId, userId)
	if room != nil {
		return room, nil
	}

	if result := rs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.id = ?", roomId).
		Where("ur.user_id = ?", userId).
		Find(room); (result.Error != nil) || (result.RowsAffected == 0) {
		badRequestError := custom_errors.HTTPError{Message: "no room found", Status: http.StatusBadRequest, Details: result.Error.Error()}
		return nil, badRequestError
	}

	// update the cache

	return room, nil
}

func (rs *RoomService) findInRedis(ctx context.Context, roomId uuid.UUID, userId uint) (*Room, error) {
	roomKey := getRoomKey(roomId)

	f, err := rs.RDB.HGetAll(ctx, roomKey).Result()
	if err != nil {
		return nil, err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	s, err := rs.RDB.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	if !utils.SliceContains[string](s, strconv.Itoa(int(userId))) {
		return nil, fmt.Errorf("user is not subscribed to room")
	}

	roomSubscribers := make([]*User, 0, len(s))
	for _, roomSubscriberId := range s {

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		rs, err := rs.RDB.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		roomSubscriberIdUint, err := strconv.Atoi(roomSubscriberId)
		if err != nil {
			return nil, err
		}

		subscriber := User{
			ID:       uint(roomSubscriberIdUint),
			Username: rs["username"],
		}

		roomSubscribers = append(roomSubscribers, &subscriber)
	}

	createdAtInt, err := strconv.Atoi(f["createdAt"])
	if err != nil {
		return nil, err
	}
	updatedAtInt, err := strconv.Atoi(f["createdAt"])
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

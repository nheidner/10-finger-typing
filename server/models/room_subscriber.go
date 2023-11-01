package models

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RoomSubscriber struct {
	UserId     uuid.UUID            `json:"userId"`
	Status     SubscriberStatus     `json:"status"`
	Username   string               `json:"username"`
	GameStatus SubscriberGameStatus `json:"gameStatus"`
}

const (
	roomSubscriberStatusField     = "status"
	roomSubscriberGameStatusField = "game_status"
	roomSubscriberUsernameField   = "username"
)

// rooms:[room_id]:subscribers_ids set: user ids
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

// rooms:[room_id]:subscribers:[user_id] hash: status, username, game_status
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

// rooms:[room_id]:subscribers:[user_id]:conns set: connection ids
func getRoomSubscriberConnectionsKey(roomId, userid uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userid) + ":conns"
}

type SubscriberStatus int

const (
	NilSubscriberStatus SubscriberStatus = iota
	InactiveSubscriberStatus
	ActiveSubscriberStatus
)

func (s *SubscriberStatus) String() string {
	return []string{"undefined", "inactive", "active"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type SubscriberGameStatus int

const (
	NilSubscriberGameStatus SubscriberGameStatus = iota
	UnstartedSubscriberGameStatus
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type RoomSubscriberService struct {
	RDB *redis.Client
}

func (rss *RoomSubscriberService) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rss.RDB.SAdd(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (rss *RoomSubscriberService) RemoveRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rss.RDB.SRem(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (rss *RoomSubscriberService) SetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberStatusField: strconv.Itoa(int(status))}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, roomSubscriberStatusField).Int()
	if err != nil {
		return NilSubscriberStatus, err
	}

	return SubscriberStatus(status), nil
}

func (rss *RoomSubscriberService) SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{roomSubscriberGameStatusField: strconv.Itoa(int(status))}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (SubscriberGameStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, roomSubscriberGameStatusField).Int()
	if err != nil {
		return NilSubscriberGameStatus, err
	}

	return SubscriberGameStatus(status), nil
}

func (rss *RoomSubscriberService) GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]RoomSubscriber, error) {
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)

	r, err := rss.RDB.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	roomSubscribers := make([]RoomSubscriber, 0, len(r))
	for _, roomSubscriberIdStr := range r {
		roomSubscriberId, err := uuid.Parse(roomSubscriberIdStr)
		if err != nil {
			return nil, err
		}

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		r, err := rss.RDB.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		status := NilSubscriberStatus
		statusStr, ok := r[roomSubscriberStatusField]
		if ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err != nil {
				return nil, err
			}
			status = SubscriberStatus(statusInt)
		}

		subscriberGameStatus := NilSubscriberGameStatus
		subscriberGameStatusStr, ok := r[roomSubscriberGameStatusField]
		if ok {
			subscriberGameStatusInt, err := strconv.Atoi(subscriberGameStatusStr)
			if err != nil {
				return nil, err
			}

			subscriberGameStatus = SubscriberGameStatus(subscriberGameStatusInt)
		}

		username := r[roomSubscriberUsernameField]

		roomSubscribers = append(roomSubscribers, RoomSubscriber{
			UserId:     roomSubscriberId,
			Status:     status,
			GameStatus: subscriberGameStatus,
			Username:   username,
		})
	}

	return roomSubscribers, nil
}

package models

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	subscriberStatusField     = "status"
	subscriberGameStatusField = "game_status"
)

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

type WSMessage struct {
	Type    string                 `json:"type"`    // user_joined (userId), new_game (textId, gameId), results (...), cursor (position), countdown_start, user_left (userId), initial_state(initial state)
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func (rss *RoomSubscriberService) SetRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rss.RDB.SAdd(ctx, roomSubscriberConnectionsKey, connectionId.String()).Err()
}

func (rss *RoomSubscriberService) RemoveRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)

	return rss.RDB.SRem(ctx, roomSubscriberConnectionsKey, connectionId).Err()
}

func (rss *RoomSubscriberService) SetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID, status SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{subscriberStatusField: status.String()}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, subscriberStatusField).Int()
	if err != nil {
		return NilSubscriberStatus, err
	}

	return SubscriberStatus(status), nil
}

func (rss *RoomSubscriberService) SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{subscriberGameStatusField: status.String()}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (SubscriberGameStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, subscriberGameStatusField).Int()
	if err != nil {
		return NilSubscriberGameStatus, err
	}

	return SubscriberGameStatus(status), nil
}

func (rss *RoomSubscriberService) GetMessages(ctx context.Context, roomId uuid.UUID, startTimestamp time.Time) <-chan []byte {
	out := make(chan []byte)

	go func() {
		roomStreamKey := getRoomStreamKey(roomId)
		id := "$"
		if (startTimestamp != time.Time{}) {
			id = strconv.FormatInt(startTimestamp.UnixMilli(), 10)
		}

		for {
			r, err := rss.RDB.XRead(ctx, &redis.XReadArgs{
				Streams: []string{roomStreamKey, id},
				Count:   1,
				Block:   0,
			}).Result()
			if err != nil {
				log.Println("error reading stream: ", err)
				continue
			}

			id = r[0].Messages[0].ID
			values := r[0].Messages[0].Values

			out <- []byte(values["data"].(string))
		}
	}()

	return out
}

// func (rss *RoomSubscriberService) Publish(ctx context.Context, rs *RoomSubscriber, msg WSMessage) error {
// 	roomStreamKey := getRoomStreamKey(rs.RoomId)
// 	data, err := json.Marshal(msg)
// 	if err != nil {
// 		return err
// 	}

// 	return rss.RDB.XAdd(ctx, &redis.XAddArgs{
// 		Stream: roomStreamKey,
// 		Values: map[string]interface{}{"data": data},
// 	}).Err()
// }

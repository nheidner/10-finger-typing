package models

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

type RoomSubscriber struct {
	ConnectionId uuid.UUID
	RoomId       uuid.UUID
	UserId       uuid.UUID
	Conn         *websocket.Conn
}

func (rss *RoomSubscriberService) NewRoomSubscriber(ctx context.Context, conn *websocket.Conn, roomId, userId uuid.UUID) *RoomSubscriber {
	roomSubscriberConnectionId := uuid.New()

	return &RoomSubscriber{
		ConnectionId: roomSubscriberConnectionId,
		RoomId:       roomId,
		UserId:       userId,
		Conn:         conn,
	}
}

func (rss *RoomSubscriberService) InitRoomSubscriber(ctx context.Context, rs *RoomSubscriber) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(rs.RoomId, rs.UserId)
	err := rss.RDB.SAdd(ctx, roomSubscriberConnectionsKey, rs.ConnectionId.String()).Err()
	if err != nil {
		return err
	}

	return rss.SetRoomSubscriberStatus(ctx, rs, ActiveSubscriberStatus)
}

func (rss *RoomSubscriberService) SetRoomSubscriberStatus(ctx context.Context, rs *RoomSubscriber, status SubscriberStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(rs.RoomId, rs.UserId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{subscriberStatusField: status.String()}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberStatus(ctx context.Context, rs *RoomSubscriber) (SubscriberStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(rs.RoomId, rs.UserId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, subscriberStatusField).Int()
	if err != nil {
		return NilSubscriberStatus, err
	}

	return SubscriberStatus(status), nil
}

func (rss *RoomSubscriberService) SetRoomSubscriberGameStatus(ctx context.Context, rs *RoomSubscriber, status SubscriberGameStatus) error {
	roomSubscriberKey := getRoomSubscriberKey(rs.RoomId, rs.UserId)

	return rss.RDB.HSet(ctx, roomSubscriberKey, map[string]interface{}{subscriberGameStatusField: status.String()}).Err()
}

func (rss *RoomSubscriberService) GetRoomSubscriberGameStatus(ctx context.Context, rs *RoomSubscriber) (SubscriberGameStatus, error) {
	roomSubscriberKey := getRoomSubscriberKey(rs.RoomId, rs.UserId)

	status, err := rss.RDB.HGet(ctx, roomSubscriberKey, subscriberGameStatusField).Int()
	if err != nil {
		return NilSubscriberGameStatus, err
	}

	return SubscriberGameStatus(status), nil
}

func (rss *RoomSubscriberService) Subscribe(ctx context.Context, rs *RoomSubscriber, startTimestamp time.Time) {
	roomStreamKey := getRoomStreamKey(rs.RoomId)
	id := "$"
	if (startTimestamp != time.Time{}) {
		id = strconv.Itoa(int(startTimestamp.Unix()))
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

		wsjson.Write(ctx, rs.Conn, values["data"])
	}
}

func (rss *RoomSubscriberService) Publish(ctx context.Context, rs *RoomSubscriber, msg WSMessage) error {
	roomStreamKey := getRoomStreamKey(rs.RoomId)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return rss.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{"data": data},
	}).Err()
}

func (rss *RoomSubscriberService) Close(ctx context.Context, rs *RoomSubscriber) error {
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(rs.RoomId, rs.UserId)
	err := rss.RDB.SRem(ctx, roomSubscriberConnectionsKey, rs.ConnectionId).Err()
	if err != nil {
		return err
	}

	err = rss.SetRoomSubscriberStatus(ctx, rs, InactiveSubscriberStatus)
	if err != nil {
		return err
	}

	return rs.Conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

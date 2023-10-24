package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type RoomSubscriberService struct {
	RDB *redis.Client
}

type WSMessage struct {
	Type    string                 `json:"type"`    // user_joined (userId), new_game (textId, gameId), results (...), cursor (position), countdown_start
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

type RoomSubscriber struct {
	ConnectionId uuid.UUID
	RoomId       uuid.UUID
	Userid       uuid.UUID
	Conn         *websocket.Conn
}

func (rss *RoomSubscriberService) InitRoomSubscriber(ctx context.Context, conn *websocket.Conn, roomId, userId uuid.UUID) (*RoomSubscriber, error) {
	roomSubscriberConnectionId := uuid.New()

	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(roomId, userId)
	err := rss.RDB.SAdd(ctx, roomSubscriberConnectionsKey, roomSubscriberConnectionId.String()).Err()
	if err != nil {
		return nil, err
	}

	return &RoomSubscriber{
		ConnectionId: roomSubscriberConnectionId,
		Conn:         conn,
		RoomId:       roomId,
		Userid:       userId,
	}, nil
}

func (rss *RoomSubscriberService) Subscribe(ctx context.Context, rs *RoomSubscriber, startTimestamp time.Time) {
	roomStreamKey := getRoomStreamKey(rs.RoomId)
	id := "$"
	if (startTimestamp != time.Time{}) {
		id = startTimestamp.String()
	}

	for {
		r, err := rss.RDB.XRead(ctx, &redis.XReadArgs{
			Streams: []string{roomStreamKey, id},
			Count:   1,
			Block:   0,
		}).Result()
		if err != nil {
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
	roomSubscriberConnectionsKey := getRoomSubscriberConnectionsKey(rs.RoomId, rs.Userid)
	err := rss.RDB.SRem(ctx, roomSubscriberConnectionsKey, rs.ConnectionId).Err()
	if err != nil {
		return err
	}

	return rs.Conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

package models

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

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

func (rss *RoomSubscriberService) GetMessages(ctx context.Context, roomId uuid.UUID, startTimestamp time.Time) (<-chan []byte, <-chan error) {
	out := make(chan []byte)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

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
				errCh <- err
				break
			}

			id = r[0].Messages[0].ID
			values := r[0].Messages[0].Values

			if action, ok := values["action"]; ok {
				switch action {
				case "terminate":
					return
				default:
				}
			}

			out <- []byte(values["data"].(string))
		}
	}()

	return out, errCh
}

func (rss *RoomSubscriberService) Publish(ctx context.Context, roomId uuid.UUID, msg WSMessage) error {
	roomStreamKey := getRoomStreamKey(roomId)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return rss.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{"data": data},
	}).Err()
}

func (rss *RoomSubscriberService) RemoveRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	err := rss.RDB.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		return err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	return rss.RDB.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err()
}

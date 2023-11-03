package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type roomSubscription struct {
	connectionId            uuid.UUID
	roomId                  uuid.UUID
	userId                  uuid.UUID
	conn                    *websocket.Conn
	roomSubscriberRedisRepo *repositories.RoomSubscriberRedisRepository
	roomStreamRedisRepo     *repositories.RoomStreamRedisRepository
}

func newRoomSubscription(
	conn *websocket.Conn, roomId, userId uuid.UUID,
	roomSubscriberRedisRepo *repositories.RoomSubscriberRedisRepository,
	roomStreamRedisRepo *repositories.RoomStreamRedisRepository,
) *roomSubscription {
	roomSubscriptionConnectionId := uuid.New()

	return &roomSubscription{
		connectionId:            roomSubscriptionConnectionId,
		roomId:                  roomId,
		userId:                  userId,
		conn:                    conn,
		roomSubscriberRedisRepo: roomSubscriberRedisRepo,
		roomStreamRedisRepo:     roomStreamRedisRepo,
	}
}

func (rs *roomSubscription) initRoomSubscriber(ctx context.Context) error {
	err := rs.roomSubscriberRedisRepo.SetRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)
	if err != nil {
		return err
	}

	return rs.roomSubscriberRedisRepo.SetRoomSubscriberStatus(ctx, rs.roomId, rs.userId, models.ActiveSubscriberStatus)
}

func (rs *roomSubscription) close(ctx context.Context) error {
	rs.roomSubscriberRedisRepo.RemoveRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)

	err := rs.roomSubscriberRedisRepo.SetRoomSubscriberStatus(ctx, rs.roomId, rs.userId, models.InactiveSubscriberStatus)
	if err != nil {
		return err
	}

	return rs.conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

func (rs *roomSubscription) subscribe(ctx context.Context, startTimestamp time.Time) error {
	messagesCh, errCh := rs.roomStreamRedisRepo.GetPushMessages(ctx, rs.roomId, startTimestamp)

	for {
		select {
		case message, ok := <-messagesCh:
			if !ok {
				return nil
			}

			rs.conn.Write(ctx, websocket.MessageText, message)
		case err, ok := <-errCh:
			if !ok {
				return nil
			}

			return err
		}
	}
}

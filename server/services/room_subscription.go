package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type roomSubscription struct {
	connectionId uuid.UUID
	roomId       uuid.UUID
	userId       uuid.UUID
	conn         *websocket.Conn
	cacheRepo    repositories.CacheRepository
	cancel       context.CancelFunc
}

func newRoomSubscription(
	conn *websocket.Conn, roomId, userId uuid.UUID,
	cacheRepo repositories.CacheRepository,
	cancel context.CancelFunc,
) *roomSubscription {
	roomSubscriptionConnectionId := uuid.New()

	return &roomSubscription{
		connectionId: roomSubscriptionConnectionId,
		roomId:       roomId,
		userId:       userId,
		conn:         conn,
		cacheRepo:    cacheRepo,
		cancel:       cancel,
	}
}

func (rs *roomSubscription) initRoomSubscriber(ctx context.Context) error {
	go func() {
		for {
			var v interface{}
			err := wsjson.Read(context.Background(), rs.conn, &v)
			if err != nil {
				log.Println("error reading from WS connection :", err)

				if websocket.CloseStatus(err) == websocket.StatusGoingAway {
					rs.cancel()
				}

				return
			}
		}
	}()

	if err := rs.cacheRepo.SetRoomSubscriberStatus(ctx, rs.roomId, rs.userId, models.ActiveSubscriberStatus); err != nil {
		return err
	}

	return rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, models.PushMessage{
		Type:    models.UserJoined,
		Payload: rs.userId,
	})
}

func (rs *roomSubscription) close(ctx context.Context) error {
	err := rs.cacheRepo.SetRoomSubscriberStatus(ctx, rs.roomId, rs.userId, models.InactiveSubscriberStatus)
	if err != nil {
		return err
	}

	userLeavePushMessage := models.PushMessage{
		Type:    models.UserLeft,
		Payload: rs.userId,
	}
	if err = rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, userLeavePushMessage); err != nil {
		return err
	}

	return rs.conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

func (rs *roomSubscription) subscribe(ctx context.Context, startTimestamp time.Time) error {
	messagesCh, errCh := rs.cacheRepo.GetPushMessages(ctx, rs.roomId, startTimestamp)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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

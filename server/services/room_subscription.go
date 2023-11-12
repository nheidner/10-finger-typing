package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const (
	observeRoomSubscriberStatusMaxDurationSeconds = 60 * 60 * 24
	observeRoomSubscriberStatusIntervalSeconds    = 4
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
			messageType, message, err := rs.conn.Read(context.Background())
			if err != nil {
				log.Println("error reading from WS connection :", err)

				if websocket.CloseStatus(err) == websocket.StatusGoingAway {
					rs.cancel()
				}

				return
			}

			if messageType == websocket.MessageText {
				var msg map[string]any
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Println("error unmarshalling message: >>", err)
					continue
				}
				switch msg["type"] {
				case "cursor":
					// TODO: handle cursor
				case "ping":
					response := map[string]any{"type": "pong"}
					responseBytes, err := json.Marshal(response)
					if err != nil {
						log.Println("error marshalling ping message: >>", err)
						continue
					}

					if err := rs.conn.Write(context.Background(), websocket.MessageText, responseBytes); err != nil {
						log.Println("error writing message: >>", err)
					}
				}
			}
		}
	}()

	roomSubscriberStatusHasBeenUpdated, err := rs.cacheRepo.SetRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)
	if err != nil {
		return err
	}

	if roomSubscriberStatusHasBeenUpdated {
		go observeRoomSubscriberStatus(context.Background(), rs.cacheRepo, rs.roomId, rs.userId)

		return rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, models.PushMessage{
			Type:    models.UserJoined,
			Payload: rs.userId,
		})
	}

	return nil
}

func (rs *roomSubscription) close(ctx context.Context) error {
	roomSubscriberStatusHasBeenUpdated, err := rs.cacheRepo.DeleteRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)
	if err != nil {
		return err
	}

	if roomSubscriberStatusHasBeenUpdated {
		userLeavePushMessage := models.PushMessage{
			Type:    models.UserLeft,
			Payload: rs.userId,
		}
		if err = rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, userLeavePushMessage); err != nil {
			return err
		}
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

func observeRoomSubscriberStatus(ctx context.Context, cacheRepo repositories.CacheRepository, roomId, userId uuid.UUID) {
	t := time.NewTicker(observeRoomSubscriberStatusIntervalSeconds * time.Second)
	maxT := time.NewTimer(observeRoomSubscriberStatusMaxDurationSeconds * time.Second)
	defer maxT.Stop()
	defer t.Stop()

	messagesCh, errCh := cacheRepo.GetPushMessages(ctx, roomId, time.Time{})

	for {
		select {
		// when a user left and shortly after (less than ticker duration) a new connection with a new ticker is established, two tickers would be running
		case messageData, ok := <-messagesCh:
			if !ok {
				log.Println("error getting push messages")
				return
			}

			var message models.PushMessage
			if err := json.Unmarshal(messageData, &message); err != nil {
				log.Println("error unmarshalling push message", err)
				return
			}
			messagePayloadStr, ok := message.Payload.(string)
			if !ok {
				log.Println("payload is not a string")
				return
			}
			if message.Type == models.UserLeft && messagePayloadStr == userId.String() {
				log.Println("stop roomSubscriber checker after receiving user_left message")
				return
			}
		case err := <-errCh:
			if err != nil {
				log.Println("error getting push messages:", err)
			}

			return
		case <-ctx.Done():
			log.Println("context done:", ctx.Err())
			return
		case <-t.C:
			numberRoomSubscriberConns, roomSubscriberStatusHasBeenUpdated, err := cacheRepo.GetRoomSubscriberStatus(ctx, roomId, userId)
			if err != nil {
				log.Println("error getting room subscriber status", err)
				return
			}

			if roomSubscriberStatusHasBeenUpdated {
				userLeavePushMessage := models.PushMessage{
					Type:    models.UserLeft,
					Payload: userId,
				}
				if err = cacheRepo.PublishPushMessage(ctx, roomId, userLeavePushMessage); err != nil {
					log.Println("error publishing push message:", err)
				}

				log.Println("stop roomSubscriber checker")
				return
			}

			if numberRoomSubscriberConns == 0 {
				log.Println("stop roomSubscriber checker")
				return
			}
		case <-maxT.C:
			log.Println("max time to update roomSubscriberStatus reached")
			return
		}
	}
}

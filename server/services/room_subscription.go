package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	observeRoomSubscriberStatusMaxDurationMinutes = 60 * 24
	observeRoomSubscriberStatusIntervalSeconds    = 4
)

type roomSubscription struct {
	connectionId uuid.UUID
	roomId       uuid.UUID
	userId       uuid.UUID
	conn         *websocket.Conn
	cacheRepo    repositories.CacheRepository
}

func newRoomSubscription(
	conn *websocket.Conn, roomId, userId uuid.UUID,
	cacheRepo repositories.CacheRepository,
) *roomSubscription {
	roomSubscriptionConnectionId := uuid.New()

	return &roomSubscription{
		connectionId: roomSubscriptionConnectionId,
		roomId:       roomId,
		userId:       userId,
		conn:         conn,
		cacheRepo:    cacheRepo,
	}
}

func (rs *roomSubscription) sendInitialState(ctx context.Context, room models.Room) error {
	existingRoomSubscribers, err := rs.cacheRepo.GetRoomSubscribers(ctx, room.ID)
	if err != nil {
		// log.Println("failed to get room subscribers:", err)
		return err
	}

	currentGame, err := rs.cacheRepo.GetCurrentGame(ctx, room.ID)
	if err != nil {
		// log.Println("failed to get current room:", err)
		return err
	}

	room.Subscribers = existingRoomSubscribers
	room.CurrentGame = currentGame

	initialMessage := &models.PushMessage{
		Type:    models.InitialState,
		Payload: room,
	}

	return wsjson.Write(ctx, rs.conn, initialMessage)
}

// reads from WS connection and handles incoming ping and cursor messages.
func (rs *roomSubscription) handleMessages(ctx context.Context) error {
	for {
		messageType, message, err := rs.conn.Read(ctx)
		if err != nil {
			// log.Println("error reading from WS connection :", err)

			return err
		}

		if messageType == websocket.MessageText {
			var msg map[string]any
			if err := json.Unmarshal(message, &msg); err != nil {
				// log.Println("error unmarshalling message: >>", err)

				continue
			}
			switch msg["type"] {
			case "cursor":
				// TODO: handle cursor
			case "ping":
				response := map[string]any{"type": "pong"}
				responseBytes, err := json.Marshal(response)
				if err != nil {
					// log.Println("error marshalling ping message: >>", err)

					continue
				}

				if err := rs.writeTimeout(ctx, 5*time.Second, responseBytes); err != nil {
					// log.Println("error writing message: >>", err)

					return err
				}
			}
		}
	}
}

func (rs *roomSubscription) handleRoomSubscriberStatus(ctx context.Context) error {
	// TODO: it must be clearer what the following code is doing
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
	pushMessageResultCh := rs.cacheRepo.GetPushMessages(ctx, rs.roomId, startTimestamp)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pushMessageResult, ok := <-pushMessageResultCh:
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !ok {
				return errors.New("pushMessageResultCh closed")
			}
			if pushMessageResult.Error != nil {
				return pushMessageResult.Error
			}
			if err := rs.conn.Write(ctx, websocket.MessageText, pushMessageResult.Value); err != nil {
				return err
			}
		}
	}
}

func (rs *roomSubscription) writeTimeout(ctx context.Context, timeout time.Duration, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return rs.conn.Write(ctx, websocket.MessageText, msg)
}

func observeRoomSubscriberStatus(ctx context.Context, cacheRepo repositories.CacheRepository, roomId, userId uuid.UUID) {
	ctx, cancel := context.WithCancel(ctx)
	t := time.NewTicker(observeRoomSubscriberStatusIntervalSeconds * time.Second)
	maxT := time.NewTimer(observeRoomSubscriberStatusMaxDurationMinutes * time.Second)
	defer maxT.Stop()
	defer t.Stop()
	defer cancel()

	pushMessageResultCh := cacheRepo.GetPushMessages(ctx, roomId, time.Time{})

	for {
		select {
		// when a user left and shortly after (less than ticker duration) a new connection with a new ticker is established, two tickers would be running at the same time
		case pushMessageResult, ok := <-pushMessageResultCh:
			if !ok {
				log.Println("error getting push messages")
				return
			}
			if !ok {
				log.Println("pushMessageResultCh closed")
				return
			}
			if pushMessageResult.Error != nil {
				log.Println("error getting push messages:", pushMessageResult.Error)
				return
			}

			var message models.PushMessage
			if err := json.Unmarshal(pushMessageResult.Value, &message); err != nil {
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

				log.Println("stop roomSubscriber checker roomSubscriberStatusHasBeenUpdated")
				return
			}

			if numberRoomSubscriberConns == 0 {
				log.Println("stop roomSubscriber checker numberRoomSubscriberConns == 0")
				return
			}
		case <-maxT.C:
			log.Println("max time to update roomSubscriberStatus reached")
			return
		}
	}
}

package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"encoding/json"
	"fmt"

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
	cacheRepo    common.CacheRepository
	logger       common.Logger
}

func newRoomSubscription(
	conn *websocket.Conn, roomId, userId uuid.UUID,
	cacheRepo common.CacheRepository,
	logger common.Logger,
) *roomSubscription {
	roomSubscriptionConnectionId := uuid.New()

	return &roomSubscription{
		connectionId: roomSubscriptionConnectionId,
		roomId:       roomId,
		userId:       userId,
		conn:         conn,
		cacheRepo:    cacheRepo,
		logger:       logger,
	}
}

func (rs *roomSubscription) sendInitialState(ctx context.Context, room models.Room) error {
	const op errors.Op = "services.roomSubscription.sendInitialState"

	existingRoomSubscribers, err := rs.cacheRepo.GetRoomSubscribers(ctx, room.ID)
	if err != nil {
		return errors.E(op, err)
	}

	currentGame, err := rs.cacheRepo.GetCurrentGame(ctx, room.ID)
	if err != nil {
		return errors.E(op, err)
	}

	currentGameUserIds, err := rs.cacheRepo.GetCurrentGameUserIds(ctx, room.ID)
	if err != nil {
		return errors.E(op, err)
	}

	room.Subscribers = existingRoomSubscribers
	room.CurrentGame = currentGame
	room.CurrentGame.GameSubscribers = currentGameUserIds

	initialMessage := &models.PushMessage{
		Type:    models.InitialState,
		Payload: room,
	}

	if err := wsjson.Write(ctx, rs.conn, initialMessage); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// reads from WS connection and handles incoming ping and cursor messages.
func (rs *roomSubscription) handleMessages(ctx context.Context) error {
	const op errors.Op = "services.roomSubscription.handleMessages"

	for {
		messageType, message, err := rs.conn.Read(ctx)
		if err != nil {
			return errors.E(op, err)
		}

		if messageType == websocket.MessageText {
			var msg map[string]any
			if err := json.Unmarshal(message, &msg); err != nil {
				rs.logger.Error(errors.E(op, err))

				continue
			}
			switch msg["type"] {
			case "cursor":
				// TODO: handle cursor
			case "ping":
				response := map[string]any{"type": "pong"}
				responseBytes, err := json.Marshal(response)
				if err != nil {
					rs.logger.Error(errors.E(op, err))

					continue
				}

				if err := rs.writeTimeout(ctx, 5*time.Second, responseBytes); err != nil {
					return errors.E(op, err)
				}
			}
		}
	}
}

func (rs *roomSubscription) handleRoomSubscriberStatus(ctx context.Context) error {
	const op errors.Op = "services.roomSubscription.handleRoomSubscriberStatus"

	// TODO: it must be clearer what the following code is doing
	roomSubscriberStatusHasBeenUpdated, err := rs.cacheRepo.SetRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)
	if err != nil {
		return errors.E(op, err)
	}

	if roomSubscriberStatusHasBeenUpdated {
		go observeRoomSubscriberStatus(context.Background(), rs.cacheRepo, rs.roomId, rs.userId, rs.logger)

		if err := rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, models.PushMessage{
			Type:    models.UserJoined,
			Payload: rs.userId,
		}); err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (rs *roomSubscription) close(ctx context.Context) error {
	const op errors.Op = "services.roomSubscription.close"

	roomSubscriberStatusHasBeenUpdated, err := rs.cacheRepo.DeleteRoomSubscriberConnection(ctx, rs.roomId, rs.userId, rs.connectionId)
	if err != nil {
		return errors.E(op, err)
	}

	if roomSubscriberStatusHasBeenUpdated {
		userLeavePushMessage := models.PushMessage{
			Type:    models.UserLeft,
			Payload: rs.userId,
		}
		if err = rs.cacheRepo.PublishPushMessage(ctx, rs.roomId, userLeavePushMessage); err != nil {
			return errors.E(op, err)
		}
	}

	if err := rs.conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages"); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (rs *roomSubscription) subscribe(ctx context.Context, startTimestamp time.Time) error {
	const op errors.Op = "services.roomSubscription.subscribe"

	pushMessageResultCh := rs.cacheRepo.GetPushMessages(ctx, rs.roomId, startTimestamp)

	for {
		select {
		case <-ctx.Done():
			return errors.E(op, ctx.Err())
		case pushMessageResult, ok := <-pushMessageResultCh:
			if ctx.Err() != nil {
				return errors.E(op, ctx.Err())
			}
			if !ok {
				err := fmt.Errorf("pushMessageResultCh closed")
				return errors.E(op, err)
			}
			if pushMessageResult.Error != nil {
				return errors.E(op, pushMessageResult.Error)
			}
			if err := rs.conn.Write(ctx, websocket.MessageText, pushMessageResult.Value); err != nil {
				return errors.E(op, err)
			}
		}
	}
}

func (rs *roomSubscription) writeTimeout(ctx context.Context, timeout time.Duration, msg []byte) error {
	const op errors.Op = "services.roomSubscription.writeTimeout"

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := rs.conn.Write(ctx, websocket.MessageText, msg); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func observeRoomSubscriberStatus(ctx context.Context, cacheRepo common.CacheRepository, roomId, userId uuid.UUID, logger common.Logger) {
	const op errors.Op = "services.observeRoomSubscriberStatus"

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
				err := fmt.Errorf("error getting push messages")
				logger.Error(errors.E(op, err))
				return
			}
			if !ok {
				err := fmt.Errorf("pushMessageResultCh closed")
				logger.Error(errors.E(op, err))
				return
			}
			if pushMessageResult.Error != nil {
				logger.Error(errors.E(op, pushMessageResult.Error))
				return
			}

			var message models.PushMessage
			if err := json.Unmarshal(pushMessageResult.Value, &message); err != nil {
				logger.Error(errors.E(op, err))
				return
			}
			messagePayloadStr, ok := message.Payload.(string)
			if !ok {
				err := fmt.Errorf("payload is not a string")
				logger.Error(errors.E(op, err))
				return
			}
			if message.Type == models.UserLeft && messagePayloadStr == userId.String() {
				logger.Info("stop roomSubscriber checker after receiving user_left message")
				return
			}
		case <-ctx.Done():
			logger.Error(errors.E(op, ctx.Err()))
			return
		case <-t.C:
			numberRoomSubscriberConns, roomSubscriberStatusHasBeenUpdated, err := cacheRepo.GetRoomSubscriberStatus(ctx, roomId, userId)
			if err != nil {
				logger.Error(errors.E(op, err))
				return
			}

			if roomSubscriberStatusHasBeenUpdated {
				userLeavePushMessage := models.PushMessage{
					Type:    models.UserLeft,
					Payload: userId,
				}
				if err = cacheRepo.PublishPushMessage(ctx, roomId, userLeavePushMessage); err != nil {
					logger.Error(errors.E(op, err))
				}

				logger.Info("stop roomSubscriber checker roomSubscriberStatusHasBeenUpdated")
				return
			}

			if numberRoomSubscriberConns == 0 {
				logger.Info("stop roomSubscriber checker numberRoomSubscriberConns == 0")
				return
			}
		case <-maxT.C:
			logger.Info("max time to update roomSubscriberStatus reached")
			return
		}
	}
}

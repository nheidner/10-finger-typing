package models

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RoomStreamService struct {
	RDB *redis.Client
}

const (
	streamEntryTypeField    = "type"
	streamEntryMessageField = "message"
	streamEntryActionField  = "action"
)

// rooms:[room_id]:stream stream: action: "terminate/..", data: message stringified json
func getRoomStreamKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":stream"
}

type StreamEntryType int

const (
	ActionStreamEntryType StreamEntryType = iota
	PushMessageStreamEntryType
)

type StreamActionType int

const (
	TerminateAction StreamActionType = iota
	GameUserScoreAction
)

type PushMessageType int

const (
	UserJoined PushMessageType = iota
	NewGame
	Cursor
	CountdownStart
	UserLeft
	InitialState
	GameScores
)

func (p *PushMessageType) String() string {
	return []string{"user_joined", "new_game", "cursor", "countdown_start", "user_left", "initial_state", "game_result"}[*p]
}

func (p *PushMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

type PushMessage struct {
	Type    PushMessageType `json:"type"`
	Payload any             `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func (rss *RoomStreamService) PublishPushMessage(ctx context.Context, roomId uuid.UUID, pushMessage PushMessage) error {
	roomStreamKey := getRoomStreamKey(roomId)
	pushMessageData, err := json.Marshal(pushMessage)
	if err != nil {
		return err
	}

	return rss.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]interface{}{
			streamEntryTypeField:    strconv.Itoa(int(PushMessageStreamEntryType)),
			streamEntryMessageField: pushMessageData,
		},
	}).Err()
}

func (rss *RoomStreamService) PublishAction(ctx context.Context, roomId uuid.UUID, action StreamActionType) error {
	roomStreamKey := getRoomStreamKey(roomId)

	return rss.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: roomStreamKey,
		Values: map[string]string{
			streamEntryTypeField:   strconv.Itoa(int(PushMessageStreamEntryType)),
			streamEntryActionField: strconv.Itoa(int(action)),
		},
	}).Err()
}

func (rss *RoomStreamService) GetPushMessages(ctx context.Context, roomId uuid.UUID, startTime time.Time) (<-chan []byte, <-chan error) {
	out := make(chan []byte)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		roomStreamKey := getRoomStreamKey(roomId)
		id := "$"
		if (startTime != time.Time{}) {
			id = strconv.FormatInt(startTime.UnixMilli(), 10)
		}

		for {
			r, err := rss.RDB.XRead(ctx, &redis.XReadArgs{
				Streams: []string{roomStreamKey, id},
				Count:   1,
				Block:   0,
			}).Result()
			if err != nil {
				errCh <- err
				return
			}

			id = r[0].Messages[0].ID
			values := r[0].Messages[0].Values

			streamEntryType, ok := values[streamEntryTypeField]
			if !ok {
				errCh <- errors.New("no " + streamEntryTypeField + " field in stream entry")
				return
			}

			switch streamEntryType {
			case strconv.Itoa(int(ActionStreamEntryType)):
				if values[streamEntryActionField] == strconv.Itoa(int(TerminateAction)) {
					log.Println("stream consumer is terminated")
					return
				}
			case strconv.Itoa(int(PushMessageStreamEntryType)):
				messsage, ok := values[streamEntryMessageField]
				if !ok {
					errCh <- errors.New("no " + streamEntryMessageField + " field in stream entry")
					return
				}

				messsageStr, ok := messsage.(string)
				if !ok {
					errCh <- errors.New("underlying type of " + streamEntryMessageField + " stream entry field is not string")
					return
				}

				out <- []byte(messsageStr)
			default:
				errCh <- errors.New(streamEntryTypeField + " has not a correct value in stream entry")
				return
			}
		}
	}()

	return out, errCh
}

func (rss *RoomStreamService) GetAction(ctx context.Context, roomId uuid.UUID, startTime time.Time) (<-chan StreamActionType, <-chan error) {
	out := make(chan StreamActionType)
	errCh := make(chan error)

	go func() {
		defer close(out)
		defer close(errCh)

		roomStreamKey := getRoomStreamKey(roomId)
		id := "$"
		if (startTime != time.Time{}) {
			id = strconv.FormatInt(startTime.UnixMilli(), 10)
		}

		for {
			select {
			// this does not really work because rss.RDR.XRead is blocking and not a channel operation
			case <-ctx.Done():
				return
			default:
				r, err := rss.RDB.XRead(ctx, &redis.XReadArgs{
					Streams: []string{roomStreamKey, id},
					Count:   1,
					Block:   0,
				}).Result()
				if err != nil {
					errCh <- err
					return
				}

				id = r[0].Messages[0].ID
				values := r[0].Messages[0].Values

				streamEntryType, ok := values[streamEntryTypeField]
				if !ok {
					errCh <- errors.New("no " + streamEntryTypeField + " field in stream entry")
					return
				}

				switch streamEntryType {
				case strconv.Itoa(int(ActionStreamEntryType)):
					action, ok := values[streamEntryActionField]
					if !ok {
						errCh <- errors.New("no " + streamEntryActionField + " field in stream entry")
						return
					}

					if action == strconv.Itoa(int(TerminateAction)) {
						log.Println("stream consumer is terminated")
						return
					}

					actionStr, ok := action.(string)
					if !ok {
						errCh <- errors.New("underlying type of " + streamEntryActionField + " stream entry field is not string")
						return
					}

					actionInt, err := strconv.Atoi(actionStr)
					if err != nil {
						errCh <- errors.New("action cannot be converted to an integer")
						return
					}

					out <- StreamActionType(actionInt)
				case strconv.Itoa(int(PushMessageStreamEntryType)):
				default:
					errCh <- errors.New(streamEntryTypeField + " has not a correct value in stream entry")
					return
				}
			}
		}

	}()

	return out, errCh
}

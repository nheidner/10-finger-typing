package models

import (
	"encoding/json"
	"errors"
)

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

func (p PushMessageType) String() string {
	return []string{"user_joined", "new_game", "cursor", "countdown_start", "user_left", "initial_state", "game_result"}[p]
}

func (p PushMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PushMessageType) ParseFromString(data string) error {
	stringToPushMessageTypeMap := map[string]PushMessageType{
		"user_joined":     UserJoined,
		"new_game":        NewGame,
		"cursor":          Cursor,
		"countdown_start": CountdownStart,
		"user_left":       UserLeft,
		"initial_state":   InitialState,
		"game_result":     GameScores,
	}

	pushMessageType, ok := stringToPushMessageTypeMap[data]
	if !ok {
		return errors.New("invalid PushMessageType")
	}

	*p = pushMessageType

	return nil
}

func (p *PushMessageType) UnmarshalJSON(data []byte) error {

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return p.ParseFromString(s)
}

type PushMessage struct {
	Type PushMessageType `json:"type"`
	// cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
	Payload any `json:"payload"`
}

type StreamSubscriptionResult[T []byte | StreamActionType | *UserNotification] struct {
	Error error
	Value T
}

package models

import (
	"10-typing/errors"
	"encoding/json"

	"fmt"
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
	Countdown
	UserLeft
	InitialState
	GameScores
	GameStarted
)

func (p PushMessageType) String() (string, error) {
	const op errors.Op = "models.PushMessageType.String"
	f := []string{"user_joined", "new_game", "cursor", "countdown", "user_left", "initial_state", "game_result", "game_started"}

	if int(p) >= len(f) {
		err := fmt.Errorf("invalid PushMessageType")
		return "", errors.E(op, err)
	}
	return f[p], nil
}

func (p PushMessageType) MarshalJSON() ([]byte, error) {
	const op errors.Op = "models.PushMessageType.MarshalJSON"
	pushMessageStr, err := p.String()
	if err != nil {
		return nil, errors.E(op, err)
	}

	pushMessageJson, err := json.Marshal(pushMessageStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return pushMessageJson, nil
}

func (p *PushMessageType) ParseFromString(data string) error {
	const op errors.Op = "models.ParseFromString"

	stringToPushMessageTypeMap := map[string]PushMessageType{
		"user_joined":   UserJoined,
		"new_game":      NewGame,
		"cursor":        Cursor,
		"countdown":     Countdown,
		"user_left":     UserLeft,
		"initial_state": InitialState,
		"game_result":   GameScores,
		"game_started":  GameStarted,
	}

	pushMessageType, ok := stringToPushMessageTypeMap[data]
	if !ok {
		err := fmt.Errorf("invalid PushMessageType")
		return errors.E(op, err)
	}

	*p = pushMessageType

	return nil
}

func (p *PushMessageType) UnmarshalJSON(data []byte) error {
	const op errors.Op = "models.UnmarshalJSON"

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.E(op, err)
	}

	if err := p.ParseFromString(s); err != nil {
		return errors.E(op, err)
	}

	return nil
}

type PushMessage struct {
	Type PushMessageType `json:"type"`
	// cursor: cursor position, user; start: time_stamp; finish: time_stamp; user_added: user; countdown: time_stamp; game_started: nil
	Payload any `json:"payload"`
}

type StreamSubscriptionResult[T []byte | StreamActionType | *UserNotification] struct {
	Error error
	Value T
}

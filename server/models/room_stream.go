package models

import (
	"encoding/json"
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

func (p *PushMessageType) String() string {
	return []string{"user_joined", "new_game", "cursor", "countdown_start", "user_left", "initial_state", "game_result"}[*p]
}

func (p *PushMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

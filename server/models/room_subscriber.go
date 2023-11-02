package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type RoomSubscriber struct {
	UserId     uuid.UUID            `json:"userId"`
	Status     SubscriberStatus     `json:"status"`
	Username   string               `json:"username"`
	GameStatus SubscriberGameStatus `json:"gameStatus"`
}

type SubscriberStatus int

const (
	NilSubscriberStatus SubscriberStatus = iota
	InactiveSubscriberStatus
	ActiveSubscriberStatus
)

func (s *SubscriberStatus) String() string {
	return []string{"undefined", "inactive", "active"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type SubscriberGameStatus int

const (
	NilSubscriberGameStatus SubscriberGameStatus = iota
	UnstartedSubscriberGameStatus
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

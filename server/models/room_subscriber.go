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
	InactiveSubscriberStatus SubscriberStatus = iota
	ActiveSubscriberStatus
)

func (s *SubscriberStatus) String() string {
	return []string{"inactive", "active"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type SubscriberGameStatus int

const (
	UnstartedSubscriberGameStatus SubscriberGameStatus = iota
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() string {
	return []string{"unstarted", "started", "finished"}[*s]
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

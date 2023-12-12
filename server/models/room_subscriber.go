package models

import (
	"10-typing/errors"
	"encoding/json"
	"fmt"

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

func (s *SubscriberStatus) String() (string, error) {
	const op errors.Op = "models.SubscriberStatus.String"
	f := []string{"inactive", "active"}

	if int(*s) >= len(f) {
		err := fmt.Errorf("invalid SubscriberStatus")
		return "", errors.E(op, err)
	}

	return f[*s], nil
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	const op errors.Op = "models.SubscriberStatus.MarshalJSON"

	subscriberStatusStr, err := s.String()
	if err != nil {
		return nil, errors.E(op, err)
	}
	subscriberStatusJson, err := json.Marshal(subscriberStatusStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return subscriberStatusJson, nil
}

type SubscriberGameStatus int

const (
	UnstartedSubscriberGameStatus SubscriberGameStatus = iota
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() (string, error) {
	const op errors.Op = "models.SubscriberGameStatus.String"
	f := []string{"unstarted", "started", "finished"}

	if int(*s) >= len(f) {
		err := fmt.Errorf("invalid SubscriberGameStatus")
		return "", errors.E(op, err)
	}

	return f[*s], nil
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	const op errors.Op = "models.SubscriberGameStatus.MarshalJSON"

	subscriberGameStatusStr, err := s.String()
	if err != nil {
		return nil, errors.E(op, err)
	}

	subscriberGameStatusJson, err := json.Marshal(subscriberGameStatusStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return subscriberGameStatusJson, nil
}

package models

import (
	"encoding/json"
	"errors"
)

type UserNotificationType int

const (
	RoomInvitation UserNotificationType = iota
)

func (n UserNotificationType) String() string {
	return []string{"room_invitation"}[n]
}

func (n UserNotificationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *UserNotificationType) ParseFromString(data string) error {
	stringToUserNotificationTypeMap := map[string]UserNotificationType{
		"room_invitation": RoomInvitation,
	}

	userNotificationType, ok := stringToUserNotificationTypeMap[data]
	if !ok {
		return errors.New("invalid UserNotificationType")
	}

	*n = userNotificationType

	return nil
}

func (n *UserNotificationType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return n.ParseFromString(s)
}

type UserNotification struct {
	Id      string               `json:"id"`
	Type    UserNotificationType `json:"type"`
	Payload any                  `json:"payload"`
}

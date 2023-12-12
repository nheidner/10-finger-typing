package models

import (
	"10-typing/errors"
	"encoding/json"
	"fmt"
)

type UserNotificationType int

const (
	RoomInvitation UserNotificationType = iota
)

func (n UserNotificationType) String() (string, error) {
	const op errors.Op = "models.UserNotificationType.String"
	f := []string{"room_invitation"}

	if int(n) >= len(f) {
		err := fmt.Errorf("invalid UserNotificationType")
		return "", errors.E(op, err)
	}

	return f[n], nil
}

func (n UserNotificationType) MarshalJSON() ([]byte, error) {
	const op errors.Op = "models.UserNotificationType.MarshalJSON"

	userNotificationStr, err := n.String()
	if err != nil {
		return nil, errors.E(op, err)
	}

	userNotificationJson, err := json.Marshal(userNotificationStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return userNotificationJson, nil
}

func (n *UserNotificationType) ParseFromString(data string) error {
	const op errors.Op = "models.UserNotificationType.ParseFromString"

	stringToUserNotificationTypeMap := map[string]UserNotificationType{
		"room_invitation": RoomInvitation,
	}

	userNotificationType, ok := stringToUserNotificationTypeMap[data]
	if !ok {
		err := fmt.Errorf("invalid UserNotificationType")
		return errors.E(op, err)
	}

	*n = userNotificationType

	return nil
}

func (n *UserNotificationType) UnmarshalJSON(data []byte) error {
	const op errors.Op = "models.UserNotificationType.UnmarshalJSON"

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.E(op, err)
	}

	if err := n.ParseFromString(s); err != nil {
		return errors.E(op, err)
	}

	return nil
}

type UserNotification struct {
	Id      string               `json:"id"`
	Type    UserNotificationType `json:"type"`
	Payload any                  `json:"payload"`
}

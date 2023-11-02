package models

import (
	"github.com/google/uuid"
)

type CreateRoomInput struct {
	UserIds []uuid.UUID `json:"userIds"`
	Emails  []string    `json:"emails" binding:"dive,email"`
}

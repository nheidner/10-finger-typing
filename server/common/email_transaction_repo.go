package common

import "github.com/google/uuid"

type EmailTransactionRepository interface {
	InviteNewUserToRoom(email string, token uuid.UUID) error
	InviteUserToRoom(email, username string) error
}

package models

import (
	"fmt"

	"github.com/google/uuid"
)

type EmailTransactionService struct {
	ApiKey string
}

func (es *EmailTransactionService) InviteNewUserToRoom(email string, token uuid.UUID) error {
	fmt.Println("new user: email :", email)
	return nil
}

func (es *EmailTransactionService) InviteUserToRoom(email, username string) error {
	fmt.Println("existing user: email :", email, "username :", username)
	return nil
}

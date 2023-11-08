package email_transaction_repo

import (
	"fmt"

	"github.com/google/uuid"
)

type EmailTransactionRepository struct {
	apiKey string
}

func NewEmailTransactionRepository(apiKey string) *EmailTransactionRepository {
	return &EmailTransactionRepository{apiKey}
}

func (er *EmailTransactionRepository) InviteNewUserToRoom(email string, token uuid.UUID) error {
	fmt.Println("new user: email :", email)
	return nil
}

func (er *EmailTransactionRepository) InviteUserToRoom(email, username string) error {
	fmt.Println("existing user: email :", email, "username :", username)
	return nil
}

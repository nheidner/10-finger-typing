package controllers

import "10-typing/models"

type Rooms struct {
	RoomService             *models.RoomService
	TokenService            *models.TokenService
	UserService             *models.UserService
	EmailTransactionService *models.EmailTransactionService
}

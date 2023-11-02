package controllers

import (
	"10-typing/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateRoomInput struct {
	UserIds []uuid.UUID `json:"userIds"`
	Emails  []string    `json:"emails" binding:"dive,email"`
}

func (r *Rooms) CreateRoom(c *gin.Context) {
	var input CreateRoomInput

	authenticatedUser, err := processCreateRoomHTTPParams(c, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := r.RoomService.DB.Begin()

	var emails []string

	for _, email := range input.Emails {
		user, err := r.UserService.FindByEmail(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		if user == nil {
			emails = append(emails, email)
			continue
		}

		input.UserIds = append(input.UserIds, user.ID)
	}

	// create room
	room, err := r.RoomService.Create(tx, input, authenticatedUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		tx.Rollback()
		return
	}

	// create tokens and send invites to non registered users
	for _, email := range emails {
		token, err := r.TokenService.Create(tx, room.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			tx.Rollback()
		}

		err = r.EmailTransactionService.InviteNewUserToRoom(email, token.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			tx.Rollback()
		}
	}

	// send invites to registered users
	for _, roomSubscriber := range room.Users {
		if roomSubscriber.ID == authenticatedUser.ID {
			continue
		}

		err = r.EmailTransactionService.InviteUserToRoom(roomSubscriber.Email, roomSubscriber.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			tx.Rollback()
		}
	}

	tx.Commit()

	models.StripSensitiveUserInformation(room.Users, authenticatedUser)

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func processCreateRoomHTTPParams(c *gin.Context, input *models.CreateRoomInput) (*models.User, error) {
	userContext, _ := c.Get("user")
	user, _ := userContext.(*models.User)

	if err := c.ShouldBindJSON(input); err != nil {
		return nil, fmt.Errorf("error processing HTTP body: %w", err)
	}

	if (len(input.UserIds) == 0) && (len(input.Emails) == 0) {
		return nil, fmt.Errorf("you cannot create a room just for yourself")
	}

	for _, userId := range input.UserIds {
		if userId == user.ID {
			return nil, fmt.Errorf("you cannot create a room for yourself with yourself")
		}
	}

	for _, email := range input.Emails {
		if email == user.Email {
			return nil, fmt.Errorf("you cannot create a room for yourself with yourself")
		}
	}

	input.UserIds = append(input.UserIds, user.ID)

	return user, nil
}

package controllers

import (
	"10-typing/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rooms) CreateRoom(c *gin.Context) {
	var input models.CreateRoomInput

	user, err := processCreateRoomHTTPParams(c, &input)
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
	room, err := r.RoomService.Create(tx, input)
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
	for _, roomSubscriber := range room.Subscribers {
		if roomSubscriber.ID == user.ID {
			continue
		}

		err = r.EmailTransactionService.InviteUserToRoom(roomSubscriber.Email, roomSubscriber.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			tx.Rollback()
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func processCreateRoomHTTPParams(c *gin.Context, input *models.CreateRoomInput) (*models.User, error) {
	userContext, _ := c.Get("user")
	user, _ := userContext.(*models.User)

	if err := c.ShouldBindJSON(input); err != nil {
		return nil, fmt.Errorf("error processing HTTP body: %w", err)
	}

	for _, userId := range input.UserIds {
		if userId == user.ID {
			return nil, fmt.Errorf("you cannot create a room for yourself with yourself")
		}
	}

	input.UserIds = append(input.UserIds, user.ID)

	return user, nil
}

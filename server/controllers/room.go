package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Rooms struct {
	RoomService *models.RoomService
}

func (r Rooms) CreateRoom(c *gin.Context) {
	var input models.CreateRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userContext, _ := c.Get("user")

	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		return
	}

	for _, userId := range input.UserIds {
		if userId == user.ID {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "you cannot create a room just for yourself"})
			return
		}
	}

	input.UserIds = append(input.UserIds, user.ID)

	room, err := r.RoomService.Create(input)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

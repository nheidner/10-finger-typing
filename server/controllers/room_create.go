package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (r Rooms) CreateRoom(c *gin.Context) {
	var input models.CreateRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Error processing HTTP body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userContext, _ := c.Get("user")
	user, ok := userContext.(*models.User)
	if !ok {
		log.Println("Could not read user from route context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		return
	}

	for _, userId := range input.UserIds {
		if userId == user.ID {
			log.Println("You cannot create a room for yourself with yourself")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "you cannot create a room for yourself with yourself"})
			return
		}
	}

	textIdUrlParam := c.Param("textid")
	textId, err := strconv.ParseUint(textIdUrlParam, 10, 0)
	if err != nil {
		log.Println("Error parsing the text id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the text id"})
		return
	}

	input.TextId = uint(textId)
	input.UserIds = append(input.UserIds, user.ID)

	room, err := r.RoomService.Create(input)
	if err != nil {
		log.Println("Error creating room:", err)
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

package controllers

import (
	"10-typing/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Games struct {
	GameService *models.GameService
}

func (g *Games) CreateGame(c *gin.Context) {
	var input models.CreateGameInput

	userContext, _ := c.Get("user")
	user, _ := userContext.(*models.User)

	roomIdUrlParam := c.Param("roomid")
	roomId, err := uuid.Parse(roomIdUrlParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// validate if roomId exists
	// validate that no other unstarted game exists

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := g.GameService.Create(nil, input, roomId, user.ID)
	if err != nil {
		log.Println("Error creating room:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": game})
}

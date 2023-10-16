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
	RoomService *models.RoomService
}

func (g *Games) CreateGame(c *gin.Context) {
	user, createUserInput, roomId, err := processCreateGameHTTPParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// validate that no other unstarted game for this room exists
	hasUnstartedGames, err := g.RoomService.HasUnstartedGames(roomId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if hasUnstartedGames {
		c.JSON(http.StatusBadRequest, gin.H{"error": "there are already unstarted games for this room"})
		return
	}

	game, err := g.GameService.Create(nil, *createUserInput, roomId, user.ID)
	if err != nil {
		log.Println("Error creating room:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": game})
}

func processCreateGameHTTPParams(c *gin.Context) (*models.User, *models.CreateGameInput, uuid.UUID, error) {
	var input models.CreateGameInput

	userContext, _ := c.Get("user")
	user, _ := userContext.(*models.User)

	roomIdUrlParam := c.Param("roomid")
	roomId, err := uuid.Parse(roomIdUrlParam)
	if err != nil {
		return nil, nil, uuid.Nil, err
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, nil, uuid.Nil, err
	}

	return user, &input, roomId, nil
}

package controllers

import (
	"10-typing/models"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Games struct {
	GameService *models.GameService
	RoomService *models.RoomService
	TextService *models.TextService
}

func (g *Games) CreateGame(c *gin.Context) {
	user, textId, roomId, err := processCreateGameHTTPParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var game = models.Game{
		ID: uuid.New(),
	}
	var ctx = context.Background()

	// validate
	textExists, err := g.TextService.TextExists(ctx, textId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !textExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text does not exist"})
		return
	}

	roomHasActiveGame, err := g.RoomService.RoomHasActiveGame(ctx, roomId)
	if err != nil {
		log.Println("RoomHasActiveGame error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if roomHasActiveGame {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room has active game at the moment"})
		return
	}

	if err := g.GameService.SetNewCurrentGame(ctx, game.ID, textId, roomId, user.ID); err != nil {
		log.Println("SetNewCurrentGame", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{
		"id": game.ID.String(),
	}})
}

// }

func processCreateGameHTTPParams(c *gin.Context) (user *models.User, textId, roomId uuid.UUID, err error) {
	var input models.CreateGameInput

	userContext, _ := c.Get("user")
	user, _ = userContext.(*models.User)

	roomIdUrlParam := c.Param("roomid")
	roomId, err = uuid.Parse(roomIdUrlParam)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	return user, input.TextId, roomId, nil
}

package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	gameService *services.GameService
}

func NewGameController(gameService *services.GameService) *GameController {
	return &GameController{gameService}
}

func (gc *GameController) CreateGame(c *gin.Context) {
	var input models.CreateGameInput

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gameId, err := gc.gameService.SetNewCurrentGame(user.ID, roomId, input.TextId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{
		"id": gameId.String(),
	}})
}

func (gc *GameController) StartGame(c *gin.Context) {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		log.Println("error processing user from context: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		log.Println("error processing roomid url path segment : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	if err = gc.gameService.AddUserToGame(roomId, user.ID); err != nil {
		log.Println("error adding user to game", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	if err = gc.gameService.InitiateGameIfReady(roomId); err != nil {
		log.Println("error initiating game", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "game started"})
}

func (gc *GameController) FinishGame(c *gin.Context) {
	var input CreateScoreInput

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	if err = gc.gameService.UserFinishesGame(roomId, user.ID, input.TextId, input.WordsTyped, input.TimeElapsed, input.Errors); err != nil {
		log.Println("error user finishing game: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "score added"})
}

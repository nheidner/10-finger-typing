package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameController struct {
	gameService *services.GameService
}

func NewGameController(gameService *services.GameService) *GameController {
	return &GameController{gameService}
}

func (gc *GameController) CreateGame(c *gin.Context) {
	var input models.CreateGameInput

	user, err := getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomId, err := getRoomIdFromPath(c)
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
	user, err := getUserFromContext(c)
	if err != nil {
		log.Println("error processing user from context: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	roomId, err := getRoomIdFromPath(c)
	if err != nil {
		log.Println("error processing roomid url path segment : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	if err = gc.gameService.AddUserToGame(roomId, user.ID); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	if err = gc.gameService.InitiateGameIfReady(roomId); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "game started"})
}

type CreateScoreInput struct {
	WordsTyped  int               `json:"wordsTyped" binding:"required" faker:"boundary_start=50, boundary_end=1000"`
	TimeElapsed float64           `json:"timeElapsed" binding:"required" faker:"oneof: 60.0, 120.0, 180.0"`
	Errors      models.ErrorsJSON `json:"errors" binding:"required,typingerrors"`
	TextId      uuid.UUID         `json:"textId" binding:"required"`
	UserId      uuid.UUID         `json:"-"`
	GameId      uuid.UUID         `json:"-"`
}

func (gc *GameController) FinishGame(c *gin.Context) {
	var input CreateScoreInput

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	user, err := getUserFromContext(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	roomId, err := getRoomIdFromPath(c)
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

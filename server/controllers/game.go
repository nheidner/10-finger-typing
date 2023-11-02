package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Games struct {
	GameService           *models.GameService
	RoomService           *models.RoomService
	TextService           *models.TextService
	RoomSubscriberService *models.RoomSubscriberService
	ScoreService          *models.ScoreService
	RoomStreamService     *models.RoomStreamService
}

type GameController struct {
	gameService *services.GameService
}

func NewGameController(gameService *services.GameService) *GameController {
	return &GameController{gameService}
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

	currentGameStatus, err := g.GameService.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		log.Println("current game status error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	roomHasActiveGame := currentGameStatus == models.StartedGameStatus
	if roomHasActiveGame {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room has active game at the moment"})
		return
	}

	if err := g.GameService.SetNewCurrentGameInRedis(ctx, game.ID, textId, roomId, user.ID); err != nil {
		log.Println("SetNewCurrentGameInRedis", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{
		"id": game.ID.String(),
	}})
}

func processCreateGameHTTPParams(c *gin.Context) (user *models.User, textId, roomId uuid.UUID, err error) {
	var input models.CreateGameInput

	user, err = getUserFromContext(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	roomId, err = getRoomIdFromPath(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	return user, input.TextId, roomId, nil
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

func (g *Games) FinishGame(c *gin.Context) {
	user, roomId, scoreInput, err := processFinishGameHTTPParams(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	// check if game status is not already finished
	currentGameStatus, err := g.GameService.GetCurrentGameStatus(c.Request.Context(), roomId)
	switch {
	case err != nil:
		log.Println("error setting current game status: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	case currentGameStatus != models.StartedGameStatus:
		log.Println("game has not the correct status: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	// get game id
	gameId, err := g.GameService.GetCurrentGameId(c.Request.Context(), roomId)
	if err != nil {
		log.Println("error when getting current game id from redis: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	scoreInput.GameId = gameId
	scoreInput.UserId = user.ID

	_, err = g.ScoreService.Create(*scoreInput)
	if err != nil {
		log.Println("error when creating a new score: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	// post action on stream
	err = g.RoomStreamService.PublishAction(c.Request.Context(), roomId, models.GameUserScoreAction)
	if err != nil {
		log.Println("error when publishing the game user score action: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "score added"})
}

func processFinishGameHTTPParams(c *gin.Context) (*models.User, uuid.UUID, *models.CreateScoreInput, error) {
	var input models.CreateScoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, uuid.Nil, nil, err
	}

	user, err := getUserFromContext(c)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	roomId, err := getRoomIdFromPath(c)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	return user, roomId, &input, err
}

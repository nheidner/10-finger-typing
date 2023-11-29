package controllers

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	gameService *services.GameService
	logger      common.Logger
}

func NewGameController(gameService *services.GameService, logger common.Logger) *GameController {
	return &GameController{gameService, logger}
}

func (gc *GameController) CreateGame(c *gin.Context) {
	const op errors.Op = "controllers.GameController.CreateGame"
	var input models.CreateGameInput

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	gameId, err := gc.gameService.CreateNewCurrentGame(c.Request.Context(), user.ID, roomId, input.TextId)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), gc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{
		"id": gameId.String(),
	}})
}

func (gc *GameController) StartGame(c *gin.Context) {
	const op errors.Op = "controllers.GameController.StartGame"
	ctx := c.Request.Context()

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	if err = gc.gameService.AddUserToGame(ctx, roomId, user.ID); err != nil {
		utils.WriteError(c, errors.E(op, err), gc.logger)
		return
	}

	if err = gc.gameService.InitiateGameIfReady(ctx, roomId); err != nil {
		utils.WriteError(c, errors.E(op, err), gc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "game started"})
}

func (gc *GameController) FinishGame(c *gin.Context) {
	const op errors.Op = "controllers.GameController.FinishGame"
	var input CreateScoreInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), gc.logger)
		return
	}

	if err = gc.gameService.UserFinishesGame(c.Request.Context(), roomId, user.ID, input.TextId, input.WordsTyped, input.TimeElapsed, input.Errors); err != nil {
		utils.WriteError(c, errors.E(op, err), gc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "score added"})
}

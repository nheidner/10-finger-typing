package controllers

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateScoreInput struct {
	WordsTyped  int               `json:"wordsTyped" binding:"required"`
	TimeElapsed float64           `json:"timeElapsed" binding:"required"`
	Errors      models.ErrorsJSON `json:"errors" binding:"required,typingerrors"`
	TextId      uuid.UUID         `json:"textId" binding:"required"`
}

type FindScoresSortOption struct {
	Column string `validate:"required,oneof=accuracy errors created_at"`
	Order  string `validate:"required,oneof=desc asc"`
}

type FindScoresQuery struct {
	UserId   uuid.UUID
	GameId   uuid.UUID
	Username string `form:"username"`
}

type ScoreController struct {
	scoreService *services.ScoreService
}

func NewScoreController(scoreService *services.ScoreService) *ScoreController {
	return &ScoreController{scoreService}
}

func (sc *ScoreController) CreateScore(c *gin.Context) {
	const op errors.Op = "controllers.ScoreController.CreateScore"
	var input CreateScoreInput

	if err := c.ShouldBindJSON(&input); err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	score, err := sc.scoreService.Create(c.Request.Context(), uuid.Nil, userId, input.TextId, input.WordsTyped, input.TimeElapsed, input.Errors)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": score})
}

func (sc *ScoreController) FindScoresByUser(c *gin.Context) {
	const op errors.Op = "controllers.ScoreController.FindScoresByUser"
	var query FindScoresQuery

	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	sortOptions, err := models.BindSortByQuery(c, FindScoresSortOption{})
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	scores, err := sc.scoreService.FindScores(c.Request.Context(), userId, uuid.Nil, query.Username, sortOptions)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

func (sc *ScoreController) FindScores(c *gin.Context) {
	const op errors.Op = "controllers.ScoreController.FindScores"
	var query FindScoresQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	sortOptions, err := models.BindSortByQuery(c, FindScoresSortOption{})
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	scores, err := sc.scoreService.FindScores(c.Request.Context(), query.UserId, query.GameId, query.Username, sortOptions)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

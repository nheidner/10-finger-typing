package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"log"
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

type ScoreController struct {
	scoreService *services.ScoreService
}

func NewScoreController(scoreService *services.ScoreService) *ScoreController {
	return &ScoreController{scoreService}
}

func (sc *ScoreController) CreateScore(c *gin.Context) {
	var input CreateScoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := sc.scoreService.Create(uuid.Nil, userId, input.TextId, input.WordsTyped, input.TimeElapsed, input.Errors)
	if err != nil {
		log.Println("error when creating a new score: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": score})
}

type FindScoresQuery struct {
	UserId   uuid.UUID
	GameId   uuid.UUID
	Username string `form:"username"`
}

func (sc *ScoreController) FindScoresByUser(c *gin.Context) {
	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query FindScoresQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sortOptions, err := models.BindSortByQuery(c, FindScoresSortOption{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scores, err := sc.scoreService.FindScores(userId, uuid.Nil, query.Username, sortOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

func (sc *ScoreController) FindScores(c *gin.Context) {
	var query FindScoresQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sortOptions, err := models.BindSortByQuery(c, FindScoresSortOption{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scores, err := sc.scoreService.FindScores(query.UserId, query.GameId, query.Username, sortOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

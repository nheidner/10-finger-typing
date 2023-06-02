package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Scores struct {
	ScoreService *models.ScoreService
}

type ErrorTyping string

func (s Scores) CreateScore(c *gin.Context) {
	var input models.CreateScoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := s.ScoreService.Create(input)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": score})
}

func (s Scores) FindScores(c *gin.Context) {
	sortOptions, err := models.BindSortByQuery(c, models.FindScoresSortOption{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := models.FindScoresQuery{
		SortOptions: sortOptions,
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scores, err := s.ScoreService.FindScores(query)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"
	"strconv"

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

	userIdUrlParam := c.Param("userid")
	userId, err := strconv.ParseUint(userIdUrlParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := s.ScoreService.Create(uint(userId), input)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": score})
}

func (s Scores) FindScoresByUser(c *gin.Context) {
	userIdParam := c.Param("userid")
	userId, err := strconv.ParseUint(userIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := models.FindScoresQuery{
		UserId: uint(userId),
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sortOptions, err := models.BindSortByQuery(c, models.FindScoresSortOption{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query.SortOptions = sortOptions

	scores, err := s.ScoreService.FindScores(query)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

func (s Scores) FindScores(c *gin.Context) {
	var query models.FindScoresQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sortOptions, err := models.BindSortByQuery(c, models.FindScoresSortOption{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query.SortOptions = sortOptions

	scores, err := s.ScoreService.FindScores(query)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": scores})
}

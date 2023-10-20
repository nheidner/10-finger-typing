package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Texts struct {
	TextService   *models.TextService
	OpenAiService *models.OpenAiService
}

func (t Texts) FindText(c *gin.Context) {
	userIdUrlParam := c.Param("userid")
	userId, err := uuid.Parse(userIdUrlParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query models.FindTextQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := t.TextService.FindNewOneByUserId(userId, query)
	if err != nil {
		fmt.Println("err :>>", err.(custom_errors.HTTPError).Details)
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

func (t Texts) CreateText(c *gin.Context) {
	var input models.CreateTextInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gptText, err := t.OpenAiService.GenerateTypingText(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	text, err := t.TextService.Create(context.Background(), input, gptText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

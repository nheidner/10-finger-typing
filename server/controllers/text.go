package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Texts struct {
	TextService   *models.TextService
	OpenAiService *models.OpenAiService
}

func (t Texts) FindText(c *gin.Context) {
	userIdUrlParam := c.Param("userid")
	userId, err := strconv.ParseUint(userIdUrlParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query models.FindTextQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := t.TextService.FindNewOneByUserId(uint(userId), query)
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

	text, err := t.TextService.Create(input, gptText)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

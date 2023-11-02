package controllers

import (
	"10-typing/services"
	"10-typing/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateTextInput struct {
	Language          string `json:"language" binding:"required" faker:"oneof: de en fr"`
	Punctuation       bool   `json:"punctuation"`
	SpecialCharacters int    `json:"specialCharacters" faker:"boundary_start=1, boundary_end=20"`
	Numbers           int    `json:"numbers" faker:"boundary_start=1, boundary_end=20"`
}

type TextController struct {
	textService *services.TextService
}

func NewTextController(textService *services.TextService) *TextController {
	return &TextController{textService}
}

func (tc *TextController) FindText(c *gin.Context) {
	userId, err := utils.GetUserIdFromPath(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query struct {
		Language             string `form:"language" binding:"required"`
		Punctuation          bool   `form:"punctuation"`
		SpecialCharactersGte int    `form:"specialCharacters[gte]"`
		SpecialCharactersLte int    `form:"specialCharacters[lte]"`
		NumbersGte           int    `form:"numbers[gte]"`
		NumbersLte           int    `form:"numbers[lte]"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := tc.textService.FindNewTextForUser(
		userId,
		query.Language,
		query.Punctuation,
		query.SpecialCharactersGte,
		query.SpecialCharactersLte,
		query.NumbersGte,
		query.NumbersLte,
	)

	if err != nil {
		log.Println("err :>>", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

func (tc *TextController) CreateText(c *gin.Context) {
	var input CreateTextInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := tc.textService.Create(input.Language, input.Punctuation, input.SpecialCharacters, input.Numbers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

package controllers

import (
	"10-typing/errors"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TextController struct {
	textService *services.TextService
}

func NewTextController(textService *services.TextService) *TextController {
	return &TextController{textService}
}

func (tc *TextController) FindNewTextForUser(c *gin.Context) {
	const op errors.Op = "controllers.TextController.FindNewTextForUser"

	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
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
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	text, err := tc.textService.FindNewTextForUser(
		c.Request.Context(),
		userId,
		query.Language,
		query.Punctuation,
		query.SpecialCharactersGte,
		query.SpecialCharactersLte,
		query.NumbersGte,
		query.NumbersLte,
	)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

func (tc *TextController) FindTextById(c *gin.Context) {
	const op errors.Op = "controllers.TextController.FindTextById"

	textId, err := utils.GetTextIdFromPath(c)
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	text, err := tc.textService.FindTextById(c.Request.Context(), textId)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

func (tc *TextController) CreateText(c *gin.Context) {
	const op errors.Op = "controllers.TextController.CreateText"
	var input struct {
		Language          string `json:"language" binding:"required"`
		Punctuation       bool   `json:"punctuation"`
		SpecialCharacters int    `json:"specialCharacters"`
		Numbers           int    `json:"numbers"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	text, err := tc.textService.Create(c.Request.Context(), input.Language, "", input.Punctuation, input.SpecialCharacters, input.Numbers)
	if err != nil {
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": text})
}

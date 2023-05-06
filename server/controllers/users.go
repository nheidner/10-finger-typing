package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Users struct {
	UserService *models.UserService
}

func (u Users) FindUser(c *gin.Context) {
	var user models.User
	result := u.UserService.DB.First(&user, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (u Users) CreateUser(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := u.UserService.Create(input)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (u Users) Login(c *gin.Context) {
	var input models.LoginUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	user, err := u.UserService.Authenticate(input.Email, input.Password)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

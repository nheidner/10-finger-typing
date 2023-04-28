package controllers

import (
	"10-typing/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{Email: input.Email, Username: input.Username, Password: input.Password, FirstName: input.FirstName, LastName: input.LastName}
	fmt.Print(user)
	tx := models.DB.Create(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": tx.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

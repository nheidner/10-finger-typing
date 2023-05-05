package controllers

import (
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const userPwPepper = "secret-random-string"

func FindUser(c *gin.Context) {
	var user models.User
	result := models.DB.First(&user, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func CreateUser(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pwBytes := []byte(input.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Email:        input.Email,
		Username:     input.Username,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		IsVerified:   false,
		PasswordHash: string(hashedBytes),
	}

	if result := models.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func Authenticate(c *gin.Context) {
	var input models.LoginUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var user models.User
	result := models.DB.Where("email = ?", input.Email).Find(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if !user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not verified"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password+userPwPepper))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

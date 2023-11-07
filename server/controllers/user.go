package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=255"`
	Password  string `json:"password" binding:"omitempty,min=6,max=255"`
	FirstName string `json:"firstName" binding:"omitempty,min=3,max=255"`
	LastName  string `json:"lastName" binding:"omitempty,min=3,max=255"`
}

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService}
}

func (uc *UserController) FindUsers(c *gin.Context) {
	var query struct {
		Username    string `form:"username"`
		UsernameSub string `form:"username_contains"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authenticatedUser, err := utils.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := uc.userService.FindUsers(query.Username, query.UsernameSub)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stripSensitiveUserInformation(users, authenticatedUser)

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (uc *UserController) FindUser(c *gin.Context) {
	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uc.userService.FindUserById(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var input CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uc.userService.Create(input.Email, input.Username, input.FirstName, input.LastName, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, sessionToken, err := uc.userService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.SetCookie(c.Writer, models.CookieSession, sessionToken)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) Logout(c *gin.Context) {
	token, err := utils.ReadCookie(c.Request, models.CookieSession)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = uc.userService.DeleteSession(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.DeleteCookie(c.Writer, models.CookieSession)
	c.JSON(http.StatusOK, gin.H{"data": "Successfully logged out"})
}

func (uc *UserController) CurrentUser(c *gin.Context) {
	user, _ := utils.GetUserFromContext(c)

	c.JSON(http.StatusOK, gin.H{"data": user})
}

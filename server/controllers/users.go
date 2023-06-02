package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Users struct {
	UserService    *models.UserService
	SessionService *models.SessionService
}

func (u Users) FindUsers(c *gin.Context) {
	// Todo: Add pagination, add filtering, add sorting
}

func (u Users) FindUser(c *gin.Context) {
	user, err := u.UserService.FindOneByUsername(c.Param("username"))
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
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
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	setCookie(c.Writer, CookieSession, session.Token)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (u Users) Logout(c *gin.Context) {
	token, err := readCookie(c.Request, CookieSession)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = u.SessionService.Delete(token)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	deleteCookie(c.Writer, CookieSession)
	c.JSON(http.StatusOK, gin.H{"data": "Successfully logged out"})
}

func (u Users) CurrentUser(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// MIDDLEWARE FUNCTIONS
func (u Users) AuthRequired(c *gin.Context) {
	token, _ := readCookie(c.Request, CookieSession)
	user, err := u.SessionService.User(token)

	if user == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Set("user", user)

	c.Next()
}

func (u Users) IsAuthorizedUser(c *gin.Context) {
	userContext, _ := c.Get("user")
	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		return
	}

	userName := c.Param("username")
	if userName != user.Username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Next()
}

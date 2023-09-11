package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Users struct {
	UserService    *models.UserService
	SessionService *models.SessionService
}

func (u Users) FindUsers(c *gin.Context) {
	var query models.FindUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := u.UserService.FindUsers(query)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (u Users) FindUser(c *gin.Context) {
	userIdParam := c.Param("userid")
	userId, err := strconv.ParseUint(userIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := u.UserService.FindOneById(uint(userId))
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
	token, err := readCookie(c.Request, CookieSession)
	if err != nil {
		log.Println("Session cookie could not be read", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := u.SessionService.User(token)

	if user == nil || err != nil {
		log.Println("User related to session cookie could not be found", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Set("user", user)

	c.Next()
}

// checks if the authenticated user corresponds to the "userid" url parameter
// this middleware function must be used after AuthRequired
func (u Users) UserIdUrlParamMatchesAuthorizedUser(c *gin.Context) {
	userContext, _ := c.Get("user")

	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		c.Abort()
		return
	}

	userIdParam := c.Param("userid")
	userId, err := strconv.ParseUint(userIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if uint(userId) != user.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.Next()
}

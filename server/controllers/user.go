package controllers

import (
	"10-typing/common"
	"10-typing/errors"
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
	logger      common.Logger
}

func NewUserController(userService *services.UserService, logger common.Logger) *UserController {
	return &UserController{userService, logger}
}

func (uc *UserController) FindUsers(c *gin.Context) {
	const op errors.Op = "controllers.UserController.FindUsers"
	var query struct {
		Username    string `form:"username"`
		UsernameSub string `form:"username_contains"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	authenticatedUser, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	users, err := uc.userService.FindUsers(c.Request.Context(), query.Username, query.UsernameSub)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	stripSensitiveUserInformation(users, authenticatedUser)

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (uc *UserController) FindUser(c *gin.Context) {
	const op errors.Op = "controllers.UserController.FindUser"

	userId, err := utils.GetUserIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	user, err := uc.userService.FindUserById(c.Request.Context(), userId)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) CreateUser(c *gin.Context) {
	const op errors.Op = "controllers.UserController.CreateUser"
	var input CreateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	user, err := uc.userService.Create(c.Request.Context(), input.Email, input.Username, input.FirstName, input.LastName, input.Password)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) Login(c *gin.Context) {
	const op errors.Op = "controllers.UserController.Login"
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	user, sessionToken, err := uc.userService.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	utils.SetCookie(c.Writer, models.CookieSession, sessionToken)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) Logout(c *gin.Context) {
	const op errors.Op = "controllers.UserController.Logout"

	token, err := utils.ReadCookie(c.Request, models.CookieSession)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	err = uc.userService.DeleteSession(c.Request.Context(), token)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	utils.DeleteCookie(c.Writer, models.CookieSession)
	c.JSON(http.StatusOK, gin.H{"data": "Successfully logged out"})
}

func (uc *UserController) CurrentUser(c *gin.Context) {
	const op errors.Op = "controllers.UserController.CurrentUser"

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

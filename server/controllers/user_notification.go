package controllers

import (
	"10-typing/errors"
	"10-typing/repositories"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserNotificationController struct {
	userNotificationService *services.UserNotificationService
}

func NewUserNotificationController(userNotificationService *services.UserNotificationService) *UserNotificationController {
	return &UserNotificationController{userNotificationService}
}

func (uc *UserNotificationController) FindRealtimeUserNotification(c *gin.Context) {
	const op errors.Op = "controllers.UserNotificationController.FindRealtimeUserNotification"
	var query struct {
		LastId string `form:"lastId"`
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		errors.WriteError(c, errors.E(op, err, http.StatusBadRequest))
		return
	}

	userNotification, err := uc.userNotificationService.FindRealtimeUserNotification(c.Request.Context(), user.ID, query.LastId)
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		c.JSON(http.StatusOK, gin.H{"data": nil})
	case err != nil:
		errors.WriteError(c, errors.E(op, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userNotification})
}

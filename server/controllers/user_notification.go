package controllers

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserNotificationController struct {
	userNotificationService *services.UserNotificationService
	logger                  common.Logger
}

func NewUserNotificationController(userNotificationService *services.UserNotificationService, logger common.Logger) *UserNotificationController {
	return &UserNotificationController{userNotificationService, logger}
}

func (uc *UserNotificationController) FindRealtimeUserNotification(c *gin.Context) {
	const op errors.Op = "controllers.UserNotificationController.FindRealtimeUserNotification"
	var query struct {
		LastId string `form:"lastId"`
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), uc.logger)
		return
	}

	userNotification, err := uc.userNotificationService.FindRealtimeUserNotification(c.Request.Context(), user.ID, query.LastId)
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusOK, gin.H{"data": nil})
	case err != nil:
		utils.WriteError(c, errors.E(op, err), uc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userNotification})
}

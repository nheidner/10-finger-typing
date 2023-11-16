package controllers

import (
	"10-typing/services"
	"10-typing/utils"
	"log"
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
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query struct {
		LastId string `form:"lastId"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userNotification, err := uc.userNotificationService.FindRealtimeUserNotification(c.Request.Context(), user.ID, query.LastId)
	switch {
	case err != nil:
		log.Println("error finding real time user notification:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	case userNotification == nil:
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userNotification})
}

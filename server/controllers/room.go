package controllers

import (
	"10-typing/models"
	"10-typing/services"
	"10-typing/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Rooms struct {
	RoomService             *models.RoomService
	TokenService            *models.TokenService
	UserService             *models.UserService
	EmailTransactionService *models.EmailTransactionService
	RoomSubscriberService   *models.RoomSubscriberService
	GameService             *models.GameService
	RoomStreamService       *models.RoomStreamService
}

type RoomController struct {
	roomService *services.RoomService
}

func NewRoomController(roomService *services.RoomService) *RoomController {
	return &RoomController{roomService}
}

func (rc *RoomController) LeaveRoom(c *gin.Context) {
	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		log.Println("Failed to process HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		log.Println("Failed to process HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	if err = rc.roomService.LeaveRoom(roomId, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}
	if err != nil {
		log.Println("Failed to process HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	if err = rc.roomService.LeaveRoom(roomId, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "OK"})
}

type CreateRoomInput struct {
	UserIds []uuid.UUID `json:"userIds"`
	Emails  []string    `json:"emails" binding:"dive,email"`
}

func (rc *RoomController) CreateRoom(c *gin.Context) {
	var input CreateRoomInput

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := rc.roomService.CreateRoom(input.UserIds, input.Emails, *user)

	utils.StripSensitiveUserInformation(room.Users, user)

	c.JSON(http.StatusOK, gin.H{"data": room})
}

package controllers

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/services"
	"10-typing/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateRoomInput struct {
	UserIds         []uuid.UUID `json:"userIds"`
	Emails          []string    `json:"emails" binding:"dive,email"`
	GameDurationSec int         `json:"gameDurationSec"`
}

type RoomController struct {
	roomService *services.RoomService
	logger      common.Logger
}

func NewRoomController(roomService *services.RoomService, logger common.Logger) *RoomController {
	return &RoomController{roomService, logger}
}

func (rc *RoomController) LeaveRoom(c *gin.Context) {
	const op errors.Op = "controllers.RoomController.LeaveRoom"

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	if err = rc.roomService.LeaveRoom(c.Request.Context(), roomId, user.ID); err != nil {
		utils.WriteError(c, errors.E(op, err), rc.logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "OK"})
}

func (rc *RoomController) CreateRoom(c *gin.Context) {
	const op errors.Op = "controllers.RoomController.CreateRoom"
	var input CreateRoomInput

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	room, err := rc.roomService.CreateRoom(c.Request.Context(), input.UserIds, input.Emails, input.GameDurationSec, *user)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), rc.logger)
		return
	}

	stripSensitiveUserInformation(room.Users, user)

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func (rc *RoomController) ConnectToRoom(c *gin.Context) {
	const op errors.Op = "controllers.RoomController.ConnectToRoom"

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	roomId, err := utils.GetRoomIdFromPath(c)
	if err != nil {
		utils.WriteError(c, errors.E(op, err, http.StatusBadRequest), rc.logger)
		return
	}

	err = rc.roomService.RoomConnect(c.Request.Context(), c, roomId, user)
	if err != nil {
		utils.WriteError(c, errors.E(op, err), rc.logger)
		return
	}
}

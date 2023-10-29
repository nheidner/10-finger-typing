package controllers

import (
	"10-typing/models"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getUuidFromPath(c *gin.Context, segment string) (uuid.UUID, error) {
	segmentValue := c.Param(segment)
	uuidValue, err := uuid.Parse(segmentValue)
	if err != nil {
		return uuid.Nil, err
	}

	return uuidValue, nil
}

func getRoomIdFromPath(c *gin.Context) (roomId uuid.UUID, err error) {
	return getUuidFromPath(c, "roomid")
}

func getGameIdFromPath(c *gin.Context) (gameId uuid.UUID, err error) {
	return getUuidFromPath(c, "gameid")
}

func getUserFromContext(c *gin.Context) (user *models.User, err error) {
	userContext, userExists := c.Get("user")
	if !userExists {
		return nil, errors.New("no user in context")
	}

	user, ok := userContext.(*models.User)
	if !ok {
		return nil, errors.New("value for user key is not of type User")
	}

	return user, nil
}

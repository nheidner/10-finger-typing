package utils

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

func GetRoomIdFromPath(c *gin.Context) (roomId uuid.UUID, err error) {
	return getUuidFromPath(c, "roomid")
}

func GetGameIdFromPath(c *gin.Context) (gameId uuid.UUID, err error) {
	return getUuidFromPath(c, "gameid")
}

func GetUserIdFromPath(c *gin.Context) (userId uuid.UUID, err error) {
	return getUuidFromPath(c, "userid")
}

func GetTextIdFromPath(c *gin.Context) (userId uuid.UUID, err error) {
	return getUuidFromPath(c, "textid")
}

func GetUserFromContext(c *gin.Context) (user *models.User, err error) {
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

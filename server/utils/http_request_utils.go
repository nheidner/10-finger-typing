package utils

import (
	"10-typing/errors"
	"10-typing/models"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getUuidFromPath(c *gin.Context, segment string) (uuid.UUID, error) {
	const op errors.Op = "utils.getUuidFromPath"

	segmentValue := c.Param(segment)
	uuidValue, err := uuid.Parse(segmentValue)
	if err != nil {
		return uuid.Nil, errors.E(op, err)
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
	const op errors.Op = "utils.GetUserFromContext"

	userContext, userExists := c.Get("user")
	if !userExists {
		err := fmt.Errorf("no user in context")
		return nil, errors.E(op, err)
	}

	user, ok := userContext.(*models.User)
	if !ok {
		err := fmt.Errorf("underlying type of %#v is not %T", user, &models.User{})
		return nil, errors.E(op, err)
	}

	return user, nil
}

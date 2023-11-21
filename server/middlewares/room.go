package middlewares

import (
	"10-typing/errors"
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsRoomAdmin(cacheRepo repositories.CacheRepository) func(c *gin.Context) {
	const op errors.Op = "middlewares.IsRoomAdmin"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		isAdmin, err := cacheRepo.RoomHasAdmin(context.Background(), roomId, user.ID)
		if err != nil {
			err = errors.E(op, err, http.StatusInternalServerError)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		if !isAdmin {
			err := fmt.Errorf("authenticated user %s is not the admint of room with id %s", user.Username, roomId.String())
			err = errors.E(op, err, http.StatusForbidden, user.Username)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		c.Next()
	}
}

func IsRoomMember(cacheRepo repositories.CacheRepository) func(c *gin.Context) {
	const op errors.Op = "middlewares.IsRoomMember"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		isRoomMember, err := cacheRepo.RoomHasSubscribers(context.Background(), roomId, user.ID)
		if err != nil {
			err = errors.E(op, err, http.StatusInternalServerError)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		if !isRoomMember {
			err := fmt.Errorf("authenticated user %s is not a member of room with id %s", user.Username, roomId.String())
			err = errors.E(op, err, http.StatusForbidden)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		c.Next()
	}
}

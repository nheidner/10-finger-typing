package middlewares

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsRoomAdmin(cacheRepo common.CacheRepository, logger common.Logger) gin.HandlerFunc {
	const op errors.Op = "middlewares.IsRoomAdmin"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		isAdmin, err := cacheRepo.RoomHasAdmin(context.Background(), roomId, user.ID)
		if err != nil {
			err = errors.E(op, err, http.StatusInternalServerError, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		if !isAdmin {
			err := fmt.Errorf("authenticated user %s is not the admint of room with id %s", user.Username, roomId.String())
			err = errors.E(op, err, http.StatusForbidden, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		c.Next()
	}
}

func IsRoomMember(cacheRepo common.CacheRepository, logger common.Logger) gin.HandlerFunc {
	const op errors.Op = "middlewares.IsRoomMember"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err = errors.E(op, err, http.StatusBadRequest)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		isRoomMember, err := cacheRepo.RoomHasSubscribers(c.Request.Context(), roomId, user.ID)
		if err != nil {
			err = errors.E(op, err, http.StatusInternalServerError, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		if !isRoomMember {
			err := fmt.Errorf("authenticated user %s is not a member of room with id %s", user.Username, roomId.String())
			err = errors.E(op, err, http.StatusForbidden, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		c.Next()
	}
}

package middlewares

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsCurrentGameUser(cacheRepo common.CacheRepository, logger common.Logger) gin.HandlerFunc {
	const op errors.Op = "middlewares.IsCurrentGameUser"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err := errors.E(op, http.StatusBadRequest, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err := errors.E(op, http.StatusInternalServerError, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		isCurrentGameUser, err := cacheRepo.IsCurrentGameUser(c.Request.Context(), roomId, user.ID)
		switch {
		case err != nil:
			err := errors.E(op, http.StatusInternalServerError, err, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		case !isCurrentGameUser:
			err := fmt.Errorf("user with id %s is not subscriber to current game of room with id %s", user.ID.String(), roomId.String())
			err = errors.E(op, http.StatusBadRequest, err, user.Username)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		c.Next()
	}
}

// checks if gameid parameter identifies the current game that the roomid parameter identifies
func IsCurrentGame(cacheRepo common.CacheRepository, logger common.Logger) gin.HandlerFunc {
	const op errors.Op = "middlewares.IsCurrentGame"

	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			err = errors.E(op, http.StatusBadRequest, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		gameId, err := utils.GetGameIdFromPath(c)
		if err != nil {
			err = errors.E(op, http.StatusBadRequest, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		isCurrentGame, err := cacheRepo.IsCurrentGame(c.Request.Context(), roomId, gameId)
		switch {
		case err != nil:
			err = errors.E(op, http.StatusInternalServerError, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		case !isCurrentGame:
			err := fmt.Errorf("game with id %s is not current game of room with id %s", gameId.String(), roomId.String())
			err = errors.E(op, http.StatusBadRequest, err)
			c.Abort()
			utils.WriteError(c, err, logger)

			return
		}

		c.Next()
	}
}

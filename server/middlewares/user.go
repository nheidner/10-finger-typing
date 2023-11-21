package middlewares

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired(cacheRepo repositories.CacheRepository, dbRepo repositories.DBRepository) func(c *gin.Context) {
	const op errors.Op = "middlewares.AuthRequired"

	return func(c *gin.Context) {
		token, err := utils.ReadCookie(c.Request, models.CookieSession)
		if err != nil {
			err = errors.E(op, err, http.StatusUnauthorized)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		tokenHash := utils.HashSessionToken(token)
		user, err := cacheRepo.GetUserBySessionTokenHashInCacheOrDB(context.Background(), dbRepo, tokenHash)
		if err != nil {
			err := errors.E(op, err, http.StatusUnauthorized)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		c.Set("user", user)

		c.Next()
	}
}

// checks if the authenticated user corresponds to the "userid" url parameter
// this middleware function must be used after AuthRequired
func UserIdUrlParamMatchesAuthorizedUser() func(c *gin.Context) {
	const op errors.Op = "middlewares.UserIdUrlParamMatchesAuthorizedUser"

	return func(c *gin.Context) {
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			err = errors.E(op, http.StatusInternalServerError, err)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		userId, err := utils.GetUserIdFromPath(c)
		if err != nil {
			err = errors.E(op, http.StatusBadRequest, err)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		if userId != user.ID {
			err := fmt.Errorf("user id from extracted from path (%s) is different than user id extracted from authenticated user (%s)", userId.String(), user.ID.String())
			err = errors.E(op, http.StatusUnauthorized, err)
			c.Abort()
			errors.WriteError(c, err)

			return
		}

		c.Next()
	}
}

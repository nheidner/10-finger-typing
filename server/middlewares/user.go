package middlewares

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired(cacheRepo repositories.CacheRepository, dbRepo repositories.DBRepository) func(c *gin.Context) {
	const op errors.Op = "middlewares.AuthRequired"

	return func(c *gin.Context) {
		token, err := utils.ReadCookie(c.Request, models.CookieSession)
		if err != nil {
			err = errors.E(op, err, http.StatusUnauthorized, errors.Messages{"message": "Session cookie could not be read"})
			c.Abort()
			errors.WriteError(c, err)
			return
		}

		tokenHash := utils.HashSessionToken(token)
		user, err := cacheRepo.GetUserBySessionTokenHashInCacheOrDB(context.Background(), dbRepo, tokenHash)
		if err != nil {
			err := errors.E(op, err, http.StatusUnauthorized, errors.Messages{"message": "User related to session cookie could not be found"})
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
	return func(c *gin.Context) {
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
			c.Abort()
			return
		}

		userId, err := utils.GetUserIdFromPath(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if userId != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

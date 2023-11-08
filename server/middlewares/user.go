package middlewares

import (
	"10-typing/models"
	"10-typing/repositories"
	"10-typing/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired(dbRepo repositories.DBRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		token, err := utils.ReadCookie(c.Request, models.CookieSession)
		if err != nil {
			log.Println("Session cookie could not be read", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenHash := utils.HashSessionToken(token)
		user, err := dbRepo.FindUserByValidSessionTokenHash(tokenHash)

		if user == nil || err != nil {
			log.Println("User related to session cookie could not be found", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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

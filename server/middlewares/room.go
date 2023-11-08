package middlewares

import (
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsRoomAdmin(cacheRepo repositories.CacheRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			log.Println("roomid parameter could not be parsed", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			log.Println("authenticated user could not be read", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		isAdmin, err := cacheRepo.RoomHasAdmin(context.Background(), roomId, user.ID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if !isAdmin {
			log.Println("authenticated user is not the room admin", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bad request"})
			return
		}

		c.Next()
	}
}

func IsRoomMember(cacheRepo repositories.CacheRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		roomId, err := utils.GetRoomIdFromPath(c)
		if err != nil {
			log.Println("roomid parameter could not be parsed", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		user, err := utils.GetUserFromContext(c)
		if err != nil {
			log.Println("authenticated user could not be read", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		isRoomMember, err := cacheRepo.RoomHasSubscribers(context.Background(), roomId, user.ID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if !isRoomMember {
			log.Println("authenticated user is not a room member")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bad request"})
			return
		}

		c.Next()
	}
}

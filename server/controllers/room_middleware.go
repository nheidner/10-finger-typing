package controllers

import (
	"10-typing/models"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (r *Rooms) IsRoomAdmin(c *gin.Context) {
	roomIdUrlParam := c.Param("roomid")

	roomId, err := uuid.Parse(roomIdUrlParam)
	if err != nil {
		log.Println("roomid parameter could not be parsed", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	userContext, _ := c.Get("user")
	user, ok := userContext.(*models.User)
	if !ok {
		log.Println("authenticated user could not be read", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	isAdmin, err := r.RoomService.RoomHasAdmin(context.Background(), roomId, user.ID)
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

func (r *Rooms) IsRoomMember(c *gin.Context) {
	roomIdUrlParam := c.Param("roomid")

	roomId, err := uuid.Parse(roomIdUrlParam)
	if err != nil {
		log.Println("roomid parameter could not be parsed", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	userContext, _ := c.Get("user")
	user, ok := userContext.(*models.User)
	if !ok {
		log.Println("authenticated user could not be read", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	isRoomMember, err := r.RoomService.RoomHasSubscribers(context.Background(), roomId, user.ID)
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

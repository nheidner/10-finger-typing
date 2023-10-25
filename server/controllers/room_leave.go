package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Rooms) LeaveRoom(c *gin.Context) {
	roomId, user, err := r.processHTTPParams(c)
	if err != nil {
		log.Println("Failed to process HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	isAdmin, err := r.RoomService.RoomHasAdmin(c.Request.Context(), roomId, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if isAdmin {
		if err := r.RoomService.DeleteRoom(c.Request.Context(), roomId); err != nil {
			log.Println("failed to remove room subscriber:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove room subscriber"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "OK"})
		return
	}

	if err = r.RoomSubscriberService.RemoveRoomSubscriber(c.Request.Context(), roomId, user.ID); err != nil {
		log.Println("failed to remove room subscriber:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove room subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "OK"})
}

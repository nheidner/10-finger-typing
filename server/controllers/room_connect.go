package controllers

import (
	"10-typing/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (r *Rooms) ConnectToRoom(c *gin.Context) {
	roomId, user, err := r.processHTTPParams(c)
	if err != nil {
		log.Println("Failed to process HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process HTTP params"})
		return
	}

	timeStamp := time.Now()

	room, err := r.RoomService.Find(c.Request.Context(), roomId, user.ID)
	if err != nil {
		log.Println("no room found:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "no room found"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Println("Failed to accept websocket connection:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to accept websocket connection"})
		return
	}

	roomSubscriber := r.RoomSubscriberService.NewRoomSubscriber(c.Request.Context(), conn, roomId, user.ID)
	defer r.RoomSubscriberService.Close(c.Request.Context(), roomSubscriber)

	err = r.RoomSubscriberService.InitRoomSubscriber(c.Request.Context(), roomSubscriber)
	if err != nil {
		log.Println("Failed to initialise room subscriber:", err)
		return
	}

	wsjson.Write(c.Request.Context(), roomSubscriber.Conn, room)

	r.RoomSubscriberService.Subscribe(c.Request.Context(), roomSubscriber, timeStamp)
}

func (r *Rooms) processHTTPParams(c *gin.Context) (roomId uuid.UUID, user *models.User, err error) {
	// userId
	userContext, _ := c.Get("user")
	user, ok := userContext.(*models.User)
	if !ok {
		return uuid.Nil, nil, fmt.Errorf("could not read user from route context")
	}

	// roomId
	roomIdUrlParam := c.Param("roomid")

	roomId, err = uuid.Parse(roomIdUrlParam)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("error parsing the room id: %w", err)
	}

	return roomId, user, nil
}

package controllers

import (
	"10-typing/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type RoomSubscriber struct {
	ConnectionId uuid.UUID
	RoomId       uuid.UUID
	UserId       uuid.UUID
	Conn         *websocket.Conn
	Rss          *models.RoomSubscriberService
}

func newRoomSubscriber(conn *websocket.Conn, roomId, userId uuid.UUID, rss *models.RoomSubscriberService) *RoomSubscriber {
	roomSubscriberConnectionId := uuid.New()

	return &RoomSubscriber{
		ConnectionId: roomSubscriberConnectionId,
		RoomId:       roomId,
		UserId:       userId,
		Conn:         conn,
		Rss:          rss,
	}
}

func (rs *RoomSubscriber) initRoomSubscriber(ctx context.Context) error {
	err := rs.Rss.SetRoomSubscriberConnection(ctx, rs.RoomId, rs.UserId, rs.ConnectionId)
	if err != nil {
		return err
	}

	return rs.Rss.SetRoomSubscriberStatus(ctx, rs.RoomId, rs.UserId, models.ActiveSubscriberStatus)
}

func (rs *RoomSubscriber) close(ctx context.Context) error {
	rs.Rss.RemoveRoomSubscriberConnection(ctx, rs.RoomId, rs.UserId, rs.ConnectionId)

	err := rs.Rss.SetRoomSubscriberStatus(ctx, rs.RoomId, rs.UserId, models.InactiveSubscriberStatus)
	if err != nil {
		return err
	}

	return rs.Conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
}

func (rs *RoomSubscriber) subscribe(ctx context.Context, startTimestamp time.Time) {
	for message := range rs.Rss.GetMessages(ctx, rs.RoomId, startTimestamp) {
		rs.Conn.Write(ctx, websocket.MessageText, message)
	}
}

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

	roomSubscriber := newRoomSubscriber(conn, roomId, user.ID, r.RoomSubscriberService)
	defer roomSubscriber.close(c.Request.Context())

	err = roomSubscriber.initRoomSubscriber(c.Request.Context())
	if err != nil {
		log.Println("Failed to initialise room subscriber:", err)
		return
	}

	wsjson.Write(c.Request.Context(), roomSubscriber.Conn, room)

	roomSubscriber.subscribe(c.Request.Context(), timeStamp)
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

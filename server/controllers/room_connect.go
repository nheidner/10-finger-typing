package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/memory"
	"10-typing/models"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// save room on database or in memory? - in memory (redis)
// /train should save text in URL?
// http POST request to /texts/:textId/rooms/
// -> other user(s) are notified
// ws connection to /texts/:textId/rooms/:roomId/ws
// room contains ID, users, text(s?), scores?
// types of messages: cursor, start, finish, user added
// one user presses start -> timer: 3, 2, 1, go

// * add frontend (button to add other user)
// * clean up frontend
// * change logic so that user only needs uuid to take part
// * what happens when user is not logged in? or does not have an account?
// * error handling (how to send errors back from server here), add context (f.e. timeouts)

func (r Rooms) ConnectToRoom(c *gin.Context) {
	roomId, user, err := processHTTPParams(c)

	if err != nil {
		log.Println("Error processing the HTTP params:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error processing the HTTP params"})
		return
	}

	_, err = r.RoomService.Find(roomId, user.ID)
	if err != nil {
		log.Println("Error finding the room:", err)
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
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
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	roomManager := memory.GetRoomManager()
	room := roomManager.GetOrCreateNewRoom(roomId)
	subscriber := roomManager.InitRoomSubscriber(conn, room)
	defer roomManager.RemoveSubscriberFromRoom(room, subscriber)

	go subscribeToRoom(c.Request.Context(), subscriber, conn)
	publishToSubscribers(c.Request.Context(), subscriber, conn, user, room)
}

func processHTTPParams(c *gin.Context) (roomId uuid.UUID, user *models.User, err error) {
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

func publishToSubscribers(ctx context.Context, subscriber *memory.Subscriber, conn *websocket.Conn, user *models.User, room *memory.Room) {
	for {
		var message memory.Message
		err := wsjson.Read(ctx, conn, &message)
		message.User = user

		if err != nil {
			return
		}

		for s := range room.Subscribers {
			if s == subscriber {
				continue
			}

			select {
			case s.Msgs <- message:
			default:
				go s.CloseSlow()
			}
		}
	}
}

func subscribeToRoom(ctx context.Context, subscriber *memory.Subscriber, conn *websocket.Conn) {
	for msg := range subscriber.Msgs {
		wsjson.Write(ctx, conn, msg)
	}
}

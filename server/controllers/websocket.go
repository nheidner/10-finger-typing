package controllers

import (
	"10-typing/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	clients = make(map[*websocket.Conn]interface{})
)

// save room on database or in memory? - in memory (redis)
// /train should save text in URL?
// http POST request to /texts/:textId/rooms/
// -> other user(s) are notified
// ws connection to /texts/:textId/rooms/:roomId/ws
// room contains ID, users, text(s?), scores?
// types of messages: cursor, start, finish, user added
// one user presses start -> timer: 3, 2, 1, go

// TODOs:
// improve websocket code (use chat.go example from nhooyr.io/websocket),

type Message struct {
	Type    string                 `json:"type"`    // cursor, start, finish, user_added, countdown
	User    *models.User           `json:"user"`    // user that sent the message except for user_added
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func Websocket(c *gin.Context) {
	userContext, _ := c.Get("user")

	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		c.Abort()
		return
	}

	fmt.Println("user :>>", user)

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		fmt.Println("failed to accept :>>", err)
		return
	}

	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	clients[conn] = struct{}{}

	defer delete(clients, conn)

	for {
		var message Message

		err := wsjson.Read(c.Request.Context(), conn, &message)

		message.User = user

		fmt.Println("message.Payload :>>", message.Payload)

		if err != nil {
			fmt.Println("failed to read :>>", err)
			return
		}

		fmt.Println("message :>>", message)

		for clConn := range clients {
			if clConn == conn {
				continue
			}

			wsjson.Write(c.Request.Context(), clConn, &message)
		}
	}
}

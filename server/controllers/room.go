package controllers

import (
	custom_errors "10-typing/errors"
	"10-typing/models"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Rooms struct {
	RoomService *models.RoomService
}

func (r Rooms) CreateRoom(c *gin.Context) {
	var input models.CreateRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userContext, _ := c.Get("user")

	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		return
	}

	for _, userId := range input.UserIds {
		if userId == user.ID {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "you cannot create a room just for yourself"})
			return
		}
	}

	textIdUrlParam := c.Param("textid")

	textId, err := strconv.ParseUint(textIdUrlParam, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the text id"})
	}

	input.TextId = uint(textId)
	input.UserIds = append(input.UserIds, user.ID)

	room, err := r.RoomService.Create(input)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

var (
	rooms         = make(map[uuid.UUID]*room)
	messageBuffer = 16
)

// save room on database or in memory? - in memory (redis)
// /train should save text in URL?
// http POST request to /texts/:textId/rooms/
// -> other user(s) are notified
// ws connection to /texts/:textId/rooms/:roomId/ws
// room contains ID, users, text(s?), scores?
// types of messages: cursor, start, finish, user added
// one user presses start -> timer: 3, 2, 1, go

type room struct {
	subscribers   map[*subscriber]struct{}
	subscribersMu sync.Mutex
}

type subscriber struct {
	msgs      chan message
	closeSlow func()
}

type message struct {
	Type    string                 `json:"type"`    // cursor, start, finish, user_added, countdown
	User    *models.User           `json:"user"`    // user that sent the message except for user_added
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func (r Rooms) ConnectToRoom(c *gin.Context) {
	userContext, _ := c.Get("user")
	textIdUrlParam := c.Param("textid")
	roomIdUrlParam := c.Param("roomid")

	user, ok := userContext.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user from context"})
		return
	}

	roomId, err := uuid.Parse(roomIdUrlParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the room id"})
		return
	}

	textId, err := strconv.ParseUint(textIdUrlParam, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing the text id"})
		return
	}

	_, err = r.RoomService.Find(roomId, uint(textId), user.ID)
	if err != nil {
		c.JSON(err.(custom_errors.HTTPError).Status, gin.H{"error": err.Error()})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to accept websocket connection"})
		return
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	// checked if is part of room in DB
	// if is part of room -> check if room struct exists -> if not, one is created
	// subscriber struct is created and added to room
	// subscriber/publisher pattern -> publisher: when message is read it is sent to subscriber channels -> subscriber listens to channel and writes to websocket connection
	// improvements:
	// * add DB table rooms (or different name) and add functionality to create room and update room
	// * refactor: use MVC, and put into smaller functions
	// * error handling, add context (f.e. timeouts)

	userRoom, roomExists := rooms[roomId]
	if !roomExists {
		userRoom = &room{
			subscribers: make(map[*subscriber]struct{}),
		}

		rooms[roomId] = userRoom
	}

	roomSubscriber := subscriber{
		msgs: make(chan message, messageBuffer),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}

	userRoom.subscribers[&roomSubscriber] = struct{}{}
	defer func() {
		delete(userRoom.subscribers, &roomSubscriber)

		if len(userRoom.subscribers) == 0 {
			delete(rooms, roomId)
		}
	}()

	// subscriber
	go func() {
		for msg := range roomSubscriber.msgs {
			wsjson.Write(c.Request.Context(), conn, msg)
		}
	}()

	// publisher
	for {
		var message message
		err := wsjson.Read(c.Request.Context(), conn, &message)
		message.User = user

		if err != nil {
			return
		}

		for s := range userRoom.subscribers {
			if s == &roomSubscriber {
				continue
			}

			select {
			case s.msgs <- message:
			default:
				go s.closeSlow()
			}
		}
	}
}

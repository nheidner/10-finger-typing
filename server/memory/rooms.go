package memory

import (
	"10-typing/models"
	"sync"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const defaultMessageBuffer = 16

var (
	once        sync.Once
	roomManager *RoomManager
)

type RoomManager struct {
	rooms         map[uuid.UUID]*Room
	messageBuffer int
}

type Room struct {
	id            uuid.UUID
	Subscribers   map[*Subscriber]struct{}
	SubscribersMu sync.Mutex
}

type Subscriber struct {
	Msgs      chan Message
	CloseSlow func()
}

type Message struct {
	Type    string                 `json:"type"`    // cursor, start, finish, user_added, countdown
	User    *models.User           `json:"user"`    // user that sent the message except for user_added
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func GetRoomManager() *RoomManager {
	once.Do(func() {
		roomManager = &RoomManager{
			rooms:         make(map[uuid.UUID]*Room),
			messageBuffer: defaultMessageBuffer,
		}
	})

	return roomManager
}

func (rm *RoomManager) GetOrCreateNewRoom(roomId uuid.UUID) *Room {
	usedRoom, roomExists := rm.rooms[roomId]
	if !roomExists {
		usedRoom = &Room{
			id:          roomId,
			Subscribers: make(map[*Subscriber]struct{}),
		}

		rm.rooms[roomId] = usedRoom
	}

	return usedRoom
}

func (rm *RoomManager) InitRoomSubscriber(conn *websocket.Conn, r *Room) *Subscriber {
	roomSubscriber := Subscriber{
		Msgs: make(chan Message, rm.messageBuffer),
		CloseSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}

	r.Subscribers[&roomSubscriber] = struct{}{}

	return &roomSubscriber
}

func (rm *RoomManager) RemoveSubscriberFromRoom(r *Room, rs *Subscriber) {
	delete(r.Subscribers, rs)

	if len(r.Subscribers) == 0 {
		delete(rm.rooms, r.id)
	}
}

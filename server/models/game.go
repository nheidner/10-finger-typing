package models

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type SubscriberStatus int

const (
	UnactiveSubscriberStatus SubscriberStatus = iota
	ActiveSubscriberStatus
	HasStartedSubscriberStatus
	HasFinishedSubscriberStatus
)

type GameStatus int

func (s *SubscriberStatus) String() string {
	return []string{"undefined", "active", "hasStarted", "hasFinished"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

const (
	UnstartedGameStatus GameStatus = iota
	StartedGameStatus
	FinishedGameStatus
)

func (s *GameStatus) String() string {
	return []string{"unstarted", "started", "finished"}[*s]
}

func (s *GameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// TODO
// * don't return in json time.Time{}
// * also how to handle time.Time{} in redis?
// * use transactions in redis

// active gameuser is saved in redis, game status, status of user, starting time stamp
// games:[gameId]:users:[userId] { startTimeStamp, status(started, finished),  }
// games:[gameId]:status

// userStartGame(startTimestamp) => startTimeStamp, status: started,
// startGame() => game:isActive: true
// finishGame() => status: finished (if last one then game:isActive: false)
// joinGame() => games:user = {}	// when joining game channel is already opened, then, when websocket connection is ready, data is flushed to user

type WSMessage struct {
	Type    string                 `json:"type"`    // cursor, start, finish, user_added, countdown
	User    *User                  `json:"user"`    // user that sent the message except for user_added
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

type Subscriber struct {
	UserId         uint             `json:"userId"`
	StartTimeStamp time.Time        `json:"startTime"`
	Status         SubscriberStatus `json:"status"`
	// Msgs           chan Message `json:"-"`
	// CloseSlow      func()       `json:"-"`
}

type Game struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	TextId      uint            `json:"textId" gorm:"not null"`
	RoomId      uuid.UUID       `json:"roomId" gorm:"not null"`
	Room        Room            `json:"-"` // active room
	Scores      []Score         `json:"-"`
	Status      GameStatus      `json:"status" gorm:"-"`      // saved in redis
	Subscribers []*Subscriber   `json:"subscribers" gorm:"-"` // saved in redis
}

type GameService struct {
	DB    *gorm.DB
	Redis *redis.Client
}

type CreateGameInput struct {
	TextId uint `json:"textId"`
}

func (gs *GameService) Create(tx *gorm.DB, input CreateGameInput, roomId uuid.UUID, userId uint) (*Game, error) {
	db := gs.getDbOrTx(tx)

	newGame := Game{
		TextId: input.TextId,
		RoomId: roomId,
		Status: UnstartedGameStatus,
	}

	if err := db.Create(&newGame).Error; err != nil {
		return nil, err
	}

	newSubscriber := Subscriber{
		UserId:         userId,
		StartTimeStamp: time.Time{},
		Status:         UnactiveSubscriberStatus,
	}

	err := gs.addGameSubscriber(newGame.ID, &newSubscriber)
	if err != nil {
		return nil, err
	}

	err = gs.addGameStatus(newGame.ID, newGame.Status)
	if err != nil {
		return nil, err
	}

	newGame.Subscribers = append(newGame.Subscribers, &newSubscriber)

	return &newGame, nil
}

func (gs *GameService) addGameStatus(gameId uuid.UUID, status GameStatus) error {
	value := strconv.Itoa(int(status))
	key := getGameStatusKey(gameId)

	return gs.Redis.Set(context.Background(), key, value, 0).Err()
}

func (gs *GameService) addGameSubscriber(gameId uuid.UUID, subscriber *Subscriber) error {
	fields := map[string]any{
		"status":    strconv.Itoa(int(subscriber.Status)),
		"startTime": subscriber.StartTimeStamp.Unix(),
	}
	key := getGameSubscriberKey(gameId, int(subscriber.UserId))

	return gs.Redis.HSet(context.Background(), key, fields).Err()
}

func (gs *GameService) getDbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return gs.DB
}

func getGameSubscriberKey(gameId uuid.UUID, subscriberId int) string {
	return "games:" + gameId.String() + ":users:" + strconv.Itoa(subscriberId)
}

func getGameStatusKey(gameId uuid.UUID) string {
	return "games:" + gameId.String() + ":status"
}

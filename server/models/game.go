package models

import (
	"context"
	"encoding/json"
	"fmt"
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
// * use transactions in redis

// active gameuser is saved in redis, game status, status of user, starting time stamp
// user_data:[userId] { startTimeStamp, status(started, finished),  }
// games:[gameId]:status
// rooms:[roomId]:unstarted_games
// games:[gameId]:user_ids [userId]

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
	UserId         uuid.UUID        `json:"userId"`
	StartTimeStamp *time.Time       `json:"startTime"`
	Status         SubscriberStatus `json:"status"`
}

type Game struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	TextId      uuid.UUID       `json:"textId" gorm:"not null"`
	RoomId      uuid.UUID       `json:"roomId" gorm:"not null"`
	Scores      []Score         `json:"-"`
	Status      GameStatus      `json:"status" gorm:"-"`      // saved in redis
	Subscribers []Subscriber    `json:"subscribers" gorm:"-"` // saved in redis
}

type GameService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

type CreateGameInput struct {
	TextId uuid.UUID `json:"textId"`
}

// rooms:[roomId]:current_game:user_ids
// rooms:[roomId]:current_game text_id status
func (gs *GameService) SetNewCurrentGame(ctx context.Context, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
	if len(userIds) == 0 {
		return fmt.Errorf("at least one user id must be specified")
	}

	currentGameKey := getCurrentGameKey(roomId)
	statusStr := strconv.Itoa(int(UnstartedGameStatus))
	gameIdStr := newGameId.String()
	currentGameValue := map[string]string{
		"id":      gameIdStr,
		"text_id": textId.String(),
		"status":  statusStr,
	}
	if err := gs.RDB.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return err
	}

	currentGameUseridsKey := getCurrentGameUserIdsKey(roomId)
	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}
	if err := gs.RDB.SAdd(ctx, currentGameUseridsKey, userIdStrs...).Err(); err != nil {
		return err
	}

	return nil
}

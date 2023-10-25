package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	gameStatusField           = "status"
	subscriberStatusField     = "status"
	subscriberGameStatusField = "game_status"
)

// rooms:[roomId] hash {id, ... }
// rooms:[roomId]:subscribers set of userIds
// rooms:[roomId]:active_game hash {}
// rooms:[roomId]:active_game:users set of userIds
// conns:[userId] set of connection ids

type Room struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Subscribers []User          `json:"subscribers" gorm:"many2many:user_rooms"`
	AdminId     uuid.UUID       `json:"adminId" gorm:"not null"`
	Admin       User            `json:"-" gorm:"foreignKey:AdminId"`
	Tokens      []Token         `json:"-"`
	Games       []Game          `json:"-"`
}

type SubscriberStatus int

const (
	NilSubscriberStatus SubscriberStatus = iota
	InactiveSubscriberStatus
	ActiveSubscriberStatus
)

type WSMessage struct {
	Type    string                 `json:"type"`    // user_joined (userId), new_game (textId, gameId), results (...), cursor (position), countdown_start, user_left (userId), initial_state(initial state)
	Payload map[string]interface{} `json:"payload"` // cursor: cursor position, start: time_stamp, finish: time_stamp, user_added: user, countdown: time_stamp
}

func (s *SubscriberStatus) String() string {
	return []string{"undefined", "inactive", "active"}[*s]
}

func (s *SubscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type SubscriberGameStatus int

const (
	NilSubscriberGameStatus SubscriberGameStatus = iota
	UnstartedSubscriberGameStatus
	StartedSubscriberGameStatus
	FinishedSubscriberGameStatus
)

func (s *SubscriberGameStatus) String() string {
	return []string{"undefined", "unstarted", "started", "finished"}[*s]
}

func (s *SubscriberGameStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

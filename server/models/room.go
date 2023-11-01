package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Room struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt  `json:"deletedAt" gorm:"index"`
	Users       []User           `json:"-" gorm:"many2many:user_rooms"` // saved in DB
	Subscribers []RoomSubscriber `json:"roomSubscribers" gorm:"-"`      // saved in Redis
	AdminId     uuid.UUID        `json:"adminId" gorm:"not null"`
	Admin       User             `json:"-" gorm:"foreignKey:AdminId"`
	Tokens      []Token          `json:"-"`
	Games       []Game           `json:"-"`
	CurrentGame *Game            `json:"currentGame" gorm:"-"`
}

const (
	roomAdminIdField   = "admin_id"
	roomCreatedAtField = "created_at"
	roomUpdatedAtField = "updated_at"
)

type RoomService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

// rooms:[room_id] hash: roomAdminId, createdAt, updatedAt
func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

func (rs *RoomService) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := rs.RDB.HGet(ctx, roomKey, roomAdminIdField).Result()
	if err != nil {
		return false, err
	}

	return r == adminId.String(), nil
}

func (rs *RoomService) RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error) {
	if len(userIds) == 0 {
		return false, fmt.Errorf("at least one user id must be specified")
	}

	roomSubscriberIds := getRoomSubscriberIdsKey(roomId)
	tempUserIdsKey := "temp:" + uuid.New().String()

	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}

	if err := rs.RDB.SAdd(ctx, tempUserIdsKey, userIdStrs...).Err(); err != nil {
		return false, err
	}

	r, err := rs.RDB.SInter(ctx, roomSubscriberIds, tempUserIdsKey).Result()
	if err != nil {
		return false, err
	}

	if err := rs.RDB.Del(ctx, tempUserIdsKey).Err(); err != nil {
		return false, err
	}

	return len(r) == len(userIds), nil
}

func (rs *RoomService) RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)

	r, err := rs.RDB.Exists(ctx, roomKey).Result()
	if err != nil {
		return false, err
	}

	return r > 0, nil
}

func (rss *RoomSubscriberService) RemoveRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error {
	roomSubscriberKey := getRoomSubscriberKey(roomId, userId)
	err := rss.RDB.Del(ctx, roomSubscriberKey).Err()
	if err != nil {
		return err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	return rss.RDB.SRem(ctx, roomSubscriberIdsKey, userId.String()).Err()
}

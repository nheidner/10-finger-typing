package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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

type RoomService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func (rs *RoomService) RoomHasActiveGame(ctx context.Context, roomId uuid.UUID) (bool, error) {
	currentGameKey := getCurrentGameKey(roomId)
	statusField := "status"

	r, err := rs.RDB.HGet(ctx, currentGameKey, statusField).Result()
	switch {
	case err == redis.Nil:
		return false, nil
	case err != nil:
		return false, err
	}

	gameStatus, err := strconv.Atoi(r)
	if err != nil {
		return false, err
	}

	return GameStatus(gameStatus) == StartedGameStatus, nil
}

func (rs *RoomService) RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error) {
	roomKey := getRoomKey(roomId)
	adminIdField := "adminId"

	r, err := rs.RDB.HGet(ctx, roomKey, adminIdField).Result()
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

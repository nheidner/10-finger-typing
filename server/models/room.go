package models

import (
	custom_errors "10-typing/errors"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Room struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Subscribers []*User         `json:"subscribers" gorm:"many2many:user_rooms"`
	Tokens      []Token         `json:"-"`
	Games       []Game          `json:"-"`
}

type CreateRoomInput struct {
	UserIds []uint   `json:"userIds"`
	Emails  []string `json:"emails" binding:"dive,email"`
}

type RoomService struct {
	DB  *gorm.DB
	RDB *redis.Client
}

// rooms:[roomId] hash {id, ... }
// rooms:[roomId]:subscribers set of userIds
// rooms:[roomId]:active_game hash {}
// rooms:[roomId]:active_game:users set of userIds
// conns:[userId] set of connection ids

// create also in redis
func (rs *RoomService) Create(tx *gorm.DB, input CreateRoomInput) (*Room, error) {
	db := tx
	if db == nil {
		db = rs.DB.Begin()
	}

	var room Room
	if err := db.Create(&room).Error; err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	// subscribers
	for _, userId := range input.UserIds {
		join := map[string]any{"room_id": room.ID, "user_id": userId}

		if err := db.Table("user_rooms").Create(&join).Error; err != nil {
			return returnAndRollBackIfNeeded(tx, err)
		}
	}

	if tx == nil {
		db.Commit()
	}

	if err := db.Preload("Subscribers").Find(&room).Error; err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	if err := rs.CreateInRedis(context.Background(), &room); err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	return &room, nil
}

func (rs *RoomService) CreateInRedis(ctx context.Context, room *Room) error {
	// add room
	roomKey := getRoomKey(room.ID)
	roomValue := map[string]any{
		"createdAt": room.CreatedAt.Unix(),
		"updatedAt": room.UpdatedAt.Unix(),
	}
	if err := rs.RDB.HSet(ctx, roomKey, roomValue).Err(); err != nil {
		return err
	}

	// add room subscriber ids
	roomSubscriberIdsKey := getRoomSubscriberIdsKey(room.ID)
	roomSubscriberIdsValue := make([]string, 0, len(room.Subscribers))
	for _, subscriber := range room.Subscribers {
		subscriberId := strconv.Itoa(int(subscriber.ID))
		roomSubscriberIdsValue = append(roomSubscriberIdsValue, subscriberId)
	}

	if err := rs.RDB.SAdd(ctx, roomSubscriberIdsKey, roomSubscriberIdsValue).Err(); err != nil {
		return err
	}

	// add room subscribers
	for _, subscriber := range room.Subscribers {
		roomSubscriberKey := getRoomSubscriberKey(room.ID, strconv.Itoa(int(subscriber.ID)))
		roomSubscriberValue := map[string]any{
			"username": subscriber.Username,
		}

		if err := rs.RDB.HSet(ctx, roomSubscriberKey, roomSubscriberValue).Err(); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (rs *RoomService) FindInRedis(ctx context.Context, roomId uuid.UUID, userId uint) (*Room, error) {
	roomKey := getRoomKey(roomId)

	f, err := rs.RDB.HGetAll(ctx, roomKey).Result()
	if err != nil {
		return nil, err
	}

	roomSubscriberIdsKey := getRoomSubscriberIdsKey(roomId)
	s, err := rs.RDB.SMembers(ctx, roomSubscriberIdsKey).Result()
	if err != nil {
		return nil, err
	}

	if !contains(s, strconv.Itoa(int(userId))) {
		return nil, fmt.Errorf("user is not subscribed to room")
	}

	roomSubscribers := make([]*User, 0, len(s))
	for _, roomSubscriberId := range s {

		roomSubscriberKey := getRoomSubscriberKey(roomId, roomSubscriberId)

		rs, err := rs.RDB.HGetAll(ctx, roomSubscriberKey).Result()
		if err != nil {
			return nil, err
		}

		roomSubscriberIdUint, err := strconv.Atoi(roomSubscriberId)
		if err != nil {
			return nil, err
		}

		subscriber := User{
			ID:       uint(roomSubscriberIdUint),
			Username: rs["username"],
		}

		roomSubscribers = append(roomSubscribers, &subscriber)
	}

	createdAtInt, err := strconv.Atoi(f["createdAt"])
	if err != nil {
		return nil, err
	}
	updatedAtInt, err := strconv.Atoi(f["createdAt"])
	if err != nil {
		return nil, err
	}

	return &Room{
		ID:          roomId,
		CreatedAt:   time.Unix(int64(createdAtInt), int64((createdAtInt%1000)*1e6)),
		UpdatedAt:   time.Unix(int64(updatedAtInt), int64((updatedAtInt%1000)*1e6)),
		Subscribers: roomSubscribers,
	}, nil

}

func contains(s []string, item string) bool {
	for _, a := range s {
		if a == item {
			return true
		}
	}

	return false
}

func returnAndRollBackIfNeeded(tx *gorm.DB, err error) (*Room, error) {
	if tx == nil {
		tx.Rollback()
	}

	return nil, err
}

func (rs *RoomService) Find(roomId uuid.UUID, userId uint) (*Room, error) {
	room, _ := rs.FindInRedis(context.Background(), roomId, userId)
	if room != nil {
		return room, nil
	}

	if result := rs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.id = ?", roomId).
		Where("ur.user_id = ?", userId).
		Find(room); (result.Error != nil) || (result.RowsAffected == 0) {
		badRequestError := custom_errors.HTTPError{Message: "no room found", Status: http.StatusBadRequest, Details: result.Error.Error()}
		return nil, badRequestError
	}

	// update the cache

	return room, nil
}

func (rs *RoomService) HasUnstartedGames(roomId uuid.UUID) (bool, error) {
	ctx := context.Background()
	roomUnstartedGamesKey := getUnstartedGamesKey(roomId)

	unstartedGamesSum, err := rs.RDB.SCard(ctx, roomUnstartedGamesKey).Result()
	if err != nil {
		return false, err
	}

	return unstartedGamesSum != 0, nil
}

func (rs *RoomService) DeleteAll() error {
	return rs.DB.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

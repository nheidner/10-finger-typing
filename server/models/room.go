package models

import (
	custom_errors "10-typing/errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Room struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Subscribers []*User         `json:"-" gorm:"many2many:user_rooms"`
	TextId      uint            `json:"textId"`
}

type CreateRoomInput struct {
	UserIds []uint `json:"usernames"`
	TextId  uint   `json:"-"`
}

type RoomService struct {
	DB *gorm.DB
}

func (rs *RoomService) Create(input CreateRoomInput) (*Room, error) {
	tx := rs.DB.Begin()

	room := Room{
		TextId: input.TextId,
	}
	if err := tx.Create(&room).Error; err != nil {
		tx.Rollback()
		badRequestError := custom_errors.HTTPError{Message: "error creating room", Status: http.StatusBadRequest, Details: err.Error()}
		return nil, badRequestError
	}

	for _, userId := range input.UserIds {
		join := map[string]any{"room_id": room.ID, "user_id": userId}

		if err := tx.Table("user_rooms").Create(&join).Error; err != nil {
			tx.Rollback()
			badRequestError := custom_errors.HTTPError{Message: "error creating user room association", Status: http.StatusBadRequest, Details: err.Error()}
			return nil, badRequestError
		}
	}

	tx.Commit()

	return &room, nil
}

func (rs *RoomService) Find(roomId uuid.UUID, textId, userId uint) (*Room, error) {
	var room Room

	if result := rs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Joins("INNER JOIN users ON ur.user_id = users.id").
		Where("rooms.id = ?", roomId).
		Where("rooms.text_id = ?", textId).
		Where("users.id = ?", userId).
		First(&room); (result.Error != nil) || (result.RowsAffected == 0) {
		badRequestError := custom_errors.HTTPError{Message: "no room found", Status: http.StatusBadRequest, Details: result.Error.Error()}
		return nil, badRequestError
	}

	return &room, nil
}

func (rs *RoomService) DeleteAll() error {
	return rs.DB.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

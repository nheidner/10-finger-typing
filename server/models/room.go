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
	TextId  uint   `json:"textId"`
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

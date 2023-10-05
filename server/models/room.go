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
	Texts       []*Text         `json:"-" gorm:"many2many:text_rooms"`
	Tokens      []Token         `json:"-"`
}

type FindRoomQuery struct {
	TextId string `form:"textId" binding:"required"`
}

type NewRoomUser struct {
	Email string `json:"email" binding:"email"`
	Name  string `json:"name"`
}

type CreateRoomInput struct {
	UserIds      []uint        `json:"userIds"`
	NewRoomUsers []NewRoomUser `json:"newRoomUsers"`
	TextIds      []uint        `json:"textIds"`
}

type RoomService struct {
	DB *gorm.DB
}

func (rs *RoomService) Create(tx *gorm.DB, input CreateRoomInput) (*Room, error) {
	db := tx
	if tx == nil {
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

	// texts
	for _, textId := range input.TextIds {
		join := map[string]any{"room_id": room.ID, "text_id": textId}

		if err := db.Table("text_rooms").Create(&join).Error; err != nil {
			return returnAndRollBackIfNeeded(tx, err)
		}
	}

	if tx == nil {
		db.Commit()
	}

	if err := db.Preload("Subscribers").First(&room, room.ID).Error; err != nil {
		return returnAndRollBackIfNeeded(tx, err)
	}

	return &room, nil
}

func returnAndRollBackIfNeeded(tx *gorm.DB, err error) (*Room, error) {
	if tx == nil {
		tx.Rollback()
	}

	return nil, err
}

// finds room that is connected to user and text
func (rs *RoomService) Find(roomId uuid.UUID, textId, userId uint) (*Room, error) {
	var room Room

	if result := rs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.id = ?", roomId).
		Where("rooms.text_id = ?", textId).
		Where("ur.user_id = ?", userId).
		First(&room); (result.Error != nil) || (result.RowsAffected == 0) {
		badRequestError := custom_errors.HTTPError{Message: "no room found", Status: http.StatusBadRequest, Details: result.Error.Error()}
		return nil, badRequestError
	}

	return &room, nil
}

func (rs *RoomService) DeleteAll() error {
	return rs.DB.Exec("TRUNCATE rooms RESTART IDENTITY CASCADE").Error
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time       `json:"createdAt"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Room      Room            `json:"-"`
	RoomID    uuid.UUID       `json:"-"`
	IsUsed    bool            `json:"-"`
}

type TokenService struct {
	DB *gorm.DB
}

func (ts *TokenService) Create(tx *gorm.DB, roomId uuid.UUID) (*Token, error) {
	token := Token{
		RoomID: roomId,
	}

	db := ts.getDbOrTx(tx)

	if err := db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

func (ts *TokenService) FindById(id uuid.UUID) (*Token, error) {
	return nil, nil
}

func (ts *TokenService) getDbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return ts.DB
}

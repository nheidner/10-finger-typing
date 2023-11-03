package repositories

import (
	"10-typing/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenDbRepository struct {
	db *gorm.DB
}

func NewTokenDbRepository(db *gorm.DB) *TokenDbRepository {
	return &TokenDbRepository{db}
}

func (tr *TokenDbRepository) Create(roomId uuid.UUID) (*models.Token, error) {
	token := models.Token{
		RoomID: roomId,
	}

	if err := tr.db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

package repositories

import "gorm.io/gorm"

type GameDbRepository struct {
	db *gorm.DB
}

func NewGameDbRepository(db *gorm.DB) *GameDbRepository {
	return &GameDbRepository{db}
}

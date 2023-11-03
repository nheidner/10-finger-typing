package repositories

import (
	"10-typing/models"

	"gorm.io/gorm"
)

type UserDbRepository struct {
	db *gorm.DB
}

func NewUserDbRepository(db *gorm.DB) *UserDbRepository {
	return &UserDbRepository{db}
}

func (ur *UserDbRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	if err := ur.db.Where("email = ?", email).Find(&user).Error; err != nil || user.Email == "" {
		return nil, err
	}

	return &user, nil
}

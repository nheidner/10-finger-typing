package repositories

import (
	"10-typing/models"
	"time"

	"github.com/google/uuid"
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

	result := ur.db.Where("email = ?", email).Find(&user)
	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, nil
	}

	return &user, nil
}

func (ur *UserDbRepository) FindUsers(username, usernameSubstr string) ([]models.User, error) {
	var users []models.User
	findUsersDbQuery := ur.db

	if username != "" {
		findUsersDbQuery = findUsersDbQuery.Where("username = ?", username)
	}

	if usernameSubstr != "" {
		findUsersDbQuery = findUsersDbQuery.Where("username ILIKE ?", "%"+usernameSubstr+"%")
	}

	findUsersDbQuery.Find(&users)

	if findUsersDbQuery.Error != nil {
		return nil, findUsersDbQuery.Error
	}

	return users, nil
}

func (ur *UserDbRepository) FindOneById(userId uuid.UUID) (*models.User, error) {
	user := models.User{
		ID: userId,
	}

	result := ur.db.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (ur *UserDbRepository) FindUserByValidSessionTokenHash(tokenHash string) (*models.User, error) {
	var user models.User

	result := ur.db.
		Joins("INNER JOIN sessions ON users.id = sessions.user_id").
		Where("sessions.token_hash = ? AND sessions.created_at > ?", tokenHash, time.Now().Add(-models.SessionDuration*time.Second)).
		Find(&user)

	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, nil
	}

	return &user, nil
}

func (ur *UserDbRepository) Create(newUser models.User) (*models.User, error) {
	if err := ur.db.Create(&newUser).Error; err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (ur *UserDbRepository) Verify(userId uuid.UUID) error {
	return ur.db.Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error
}

func (ur *UserDbRepository) DeleteAll() error {
	return ur.db.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error
}

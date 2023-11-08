package sql_repo

import (
	"10-typing/models"
	"time"

	"github.com/google/uuid"
)

func (repo *SQLRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User

	result := repo.db.Where("email = ?", email).Find(&user)
	switch {
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		return nil, nil
	}

	return &user, nil
}

func (repo *SQLRepository) FindUsers(username, usernameSubstr string) ([]models.User, error) {
	var users []models.User
	findUsersDbQuery := repo.db

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

func (repo *SQLRepository) FindUserById(userId uuid.UUID) (*models.User, error) {
	user := models.User{
		ID: userId,
	}

	result := repo.db.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (repo *SQLRepository) FindUserByValidSessionTokenHash(tokenHash string) (*models.User, error) {
	var user models.User

	result := repo.db.
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

func (repo *SQLRepository) CreateUser(newUser models.User) (*models.User, error) {
	if err := repo.db.Create(&newUser).Error; err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (repo *SQLRepository) VerifyUser(userId uuid.UUID) error {
	return repo.db.Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error
}

func (repo *SQLRepository) DeleteAllUsers() error {
	return repo.db.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error
}

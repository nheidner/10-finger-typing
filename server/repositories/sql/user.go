package sql_repo

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"

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

func (repo *SQLRepository) CreateUserAndCache(cacheRepo repositories.CacheRepository, newUser models.User) (*models.User, error) {
	var ctx = context.Background()

	createdUser, err := repo.createUser(newUser)
	if err != nil {
		return nil, err
	}

	if err = cacheRepo.SetUser(ctx, *createdUser); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (repo *SQLRepository) VerifyUserAndCache(cacheRepo repositories.CacheRepository, userId uuid.UUID) error {
	var ctx = context.Background()

	if err := repo.verifyUser(userId); err != nil {
		return err
	}

	userKeyExists, err := cacheRepo.UserExists(ctx, userId)
	if err != nil {
		return err
	}
	if !userKeyExists {
		return nil
	}

	return cacheRepo.VerifyUser(ctx, userId)
}

func (repo *SQLRepository) DeleteAllUsers() error {
	return repo.db.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error
}

func (repo *SQLRepository) verifyUser(userId uuid.UUID) error {
	return repo.db.Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error
}

func (repo *SQLRepository) createUser(newUser models.User) (*models.User, error) {
	if err := repo.db.Create(&newUser).Error; err != nil {
		return nil, err
	}

	return &newUser, nil
}

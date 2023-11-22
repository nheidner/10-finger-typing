package sql_repo

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *SQLRepository) FindUserByEmail(email string) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserByEmail"
	var user models.User

	if err := repo.db.Where("email = ?", email).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, repositories.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) FindUsers(username, usernameSubstr string) ([]models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUsers"
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
		return nil, errors.E(op, findUsersDbQuery.Error)
	}

	return users, nil
}

func (repo *SQLRepository) FindUserById(userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserById"
	user := models.User{
		ID: userId,
	}

	if err := repo.db.First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, repositories.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) CreateUserAndCache(cacheRepo repositories.CacheRepository, newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateUserAndCache"
	var ctx = context.Background()

	createdUser, err := repo.createUser(newUser)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = cacheRepo.SetUser(ctx, *createdUser); err != nil {
		return nil, errors.E(op, err)
	}

	return createdUser, nil
}

func (repo *SQLRepository) VerifyUserAndCache(cacheRepo repositories.CacheRepository, userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.VerifyUserAndCache"
	var ctx = context.Background()

	if err := repo.verifyUser(userId); err != nil {
		return errors.E(op, err)
	}

	userKeyExists, err := cacheRepo.UserExists(ctx, userId)
	if err != nil {
		return errors.E(op, err)
	}
	if !userKeyExists {
		// if user is not in cache, then it also doesn't have to be verified in the cache
		return nil
	}

	if err := cacheRepo.VerifyUser(ctx, userId); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) DeleteAllUsers() error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllUsers"

	if err := repo.db.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) verifyUser(userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.verifyUser"

	if err := repo.db.Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) createUser(newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.createUser"

	if err := repo.db.Create(&newUser).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &newUser, nil
}

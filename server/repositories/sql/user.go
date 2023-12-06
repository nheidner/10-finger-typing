package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *SQLRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserByEmail"
	var user models.User

	if err := repo.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) FindUsers(ctx context.Context, username, usernameSubstr string) ([]models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUsers"
	var users []models.User
	findUsersDbQuery := repo.db.WithContext(ctx)

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

func (repo *SQLRepository) FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserById"
	user := models.User{
		ID: userId,
	}

	if err := repo.db.WithContext(ctx).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) CreateUserAndCache(ctx context.Context, cacheRepo common.CacheRepository, newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateUserAndCache"

	createdUser, err := repo.createUser(ctx, newUser)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = cacheRepo.SetUser(ctx, *createdUser); err != nil {
		return nil, errors.E(op, err)
	}

	return createdUser, nil
}

func (repo *SQLRepository) VerifyUserAndCache(ctx context.Context, cacheRepo common.CacheRepository, userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.VerifyUserAndCache"

	if err := repo.verifyUser(ctx, userId); err != nil {
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

func (repo *SQLRepository) DeleteAllUsers(ctx context.Context) error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllUsers"

	if err := repo.db.WithContext(ctx).Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) verifyUser(ctx context.Context, userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.verifyUser"

	if err := repo.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) createUser(ctx context.Context, newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.createUser"

	if err := repo.db.WithContext(ctx).Create(&newUser).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &newUser, nil
}

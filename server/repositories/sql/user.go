package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (repo *SQLRepository) FindUserByEmail(ctx context.Context, tx common.Transaction, email string) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserByEmail"
	db := repo.dbConn(tx)
	var user models.User

	if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) FindUsers(ctx context.Context, tx common.Transaction, username, usernameSubstr string) ([]models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUsers"
	db := repo.dbConn(tx)
	var users []models.User
	findUsersDbQuery := db.WithContext(ctx)

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

func (repo *SQLRepository) FindUserById(ctx context.Context, tx common.Transaction, userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindUserById"
	db := repo.dbConn(tx)
	user := models.User{
		ID: userId,
	}

	if err := db.WithContext(ctx).First(&user).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, errors.E(op, common.ErrNotFound)
		default:
			return nil, errors.E(op, err)
		}
	}

	return &user, nil
}

func (repo *SQLRepository) CreateUserAndCache(ctx context.Context, tx common.Transaction, cacheRepo common.CacheRepository, newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateUserAndCache"

	createdUser, err := repo.createUser(ctx, tx, newUser)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = cacheRepo.SetUser(ctx, nil, *createdUser); err != nil {
		return nil, errors.E(op, err)
	}

	return createdUser, nil
}

func (repo *SQLRepository) VerifyUserAndCache(ctx context.Context, tx common.Transaction, cacheRepo common.CacheRepository, userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.VerifyUserAndCache"

	if err := repo.verifyUser(ctx, tx, userId); err != nil {
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

	if err := cacheRepo.VerifyUser(ctx, nil, userId); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) DeleteAllUsers(ctx context.Context, tx common.Transaction) error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllUsers"
	db := repo.dbConn(tx)

	if err := db.WithContext(ctx).Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) verifyUser(ctx context.Context, tx common.Transaction, userId uuid.UUID) error {
	const op errors.Op = "sql_repo.SQLRepository.verifyUser"
	db := repo.dbConn(tx)

	if err := db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("is_verified", true).Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *SQLRepository) createUser(ctx context.Context, tx common.Transaction, newUser models.User) (*models.User, error) {
	const op errors.Op = "sql_repo.SQLRepository.createUser"
	db := repo.dbConn(tx)

	if err := db.WithContext(ctx).Create(&newUser).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &newUser, nil
}

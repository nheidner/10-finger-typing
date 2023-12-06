package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	userUsernameField     = "username"
	userPasswordHashField = "password_hash"
	userFirstNameField    = "first_name"
	userLastNameField     = "last_name"
	userEmailField        = "email"
	userIsVerifiedField   = "is_verified"
)

// users:[userid] hash of user data
func getUserKey(userId uuid.UUID) string {
	return "users:" + userId.String()
}

func getUserEmailKey(email string) string {
	return "user_emails:" + email
}

// if not found, queries db
func (repo *RedisRepository) GetUserByEmailInCacheOrDB(ctx context.Context, dbRepo common.DBRepository, email string) (*models.User, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetUserByEmailInCacheOrDB"
	userEmailKey := getUserEmailKey(email)

	userIdStr, err := repo.redisClient.Get(ctx, userEmailKey).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	user, err := repo.GetUserByIdInCacheOrDB(ctx, dbRepo, userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

// if not found, queries db
func (repo *RedisRepository) GetUserByIdInCacheOrDB(ctx context.Context, dbRepo common.DBRepository, userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetUserByIdInCacheOrDB"
	user, err := repo.getUser(ctx, userId)
	switch {
	case errors.Is(err, common.ErrNotFound):
		break
	case err != nil:
		return nil, errors.E(op, err)
	case err == nil:
		return user, nil
	}

	user, err = dbRepo.FindUserById(ctx, userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = repo.SetUser(ctx, *user); err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

// read: read first in cache, if not exists, read from db and write to cache (in case, keys got evicted)
func (repo *RedisRepository) GetUserBySessionTokenHashInCacheOrDB(
	ctx context.Context,
	dbRepo common.DBRepository,
	tokenHash string,
) (*models.User, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetUserBySessionTokenHashInCacheOrDB"

	userId, err := repo.getUserIdBySessionTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.E(op, err)
	}

	user, err := repo.getUser(ctx, userId)
	switch {
	case errors.Is(err, common.ErrNotFound):
		break
	case err != nil:
		return nil, errors.E(op, err)
	case err == nil:
		return user, nil
	}

	user, err = dbRepo.FindUserById(ctx, userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = repo.SetUser(ctx, *user); err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

func (repo *RedisRepository) UserExists(ctx context.Context, userId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.UserExists"
	userKey := getUserKey(userId)

	r, err := repo.redisClient.Exists(ctx, userKey).Result()
	if err != nil {
		errors.E(op, err)
	}

	return r > 0, nil
}

func (repo *RedisRepository) SetUser(ctx context.Context, user models.User) error {
	const op errors.Op = "redis_repo.RedisRepository.SetUser"
	userEmailKey := getUserEmailKey(user.Email)

	if err := repo.redisClient.Set(ctx, userEmailKey, user.ID.String(), 0).Err(); err != nil {
		return errors.E(op, err)
	}

	userKey := getUserKey(user.ID)

	if err := repo.redisClient.HSet(ctx, userKey, map[string]any{
		userUsernameField:     user.Username,
		userPasswordHashField: user.PasswordHash,
		userFirstNameField:    user.FirstName,
		userLastNameField:     user.LastName,
		userEmailField:        user.Email,
		userIsVerifiedField:   user.IsVerified,
	}).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) VerifyUser(ctx context.Context, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.VerifyUser"
	userKey := getUserKey(userId)

	if err := repo.redisClient.HSet(ctx, userKey, userIsVerifiedField, true).Err(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) DeleteAllUsers(ctx context.Context) error {
	const op errors.Op = "redis_repo.RedisRepository.DeleteAllUsers"

	if err := deleteKeysByPattern(ctx, repo, "users:*"); err != nil {
		return errors.E(op, err)
	}

	if err := deleteKeysByPattern(ctx, repo, "user_emails:*"); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (repo *RedisRepository) getUserIdBySessionTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	const op errors.Op = "redis_repo.RedisRepository.getUserIdBySessionTokenHash"
	sessionKey := getSessionKey(tokenHash)

	userIdStr, err := repo.redisClient.Get(ctx, sessionKey).Result()
	switch {
	case err == redis.Nil:
		return uuid.Nil, errors.E(op, common.ErrNotFound)
	case err != nil:
		return uuid.Nil, errors.E(op, err)
	}

	userId, err := uuid.Parse(userIdStr)
	if err != err {
		return uuid.Nil, errors.E(op, err)
	}

	return userId, nil
}

func (repo *RedisRepository) getUser(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "redis_repo.RedisRepository.getUser"
	userKey := getUserKey(userId)

	r, err := repo.redisClient.HGetAll(ctx, userKey).Result()
	switch {
	case err != nil:
		return nil, errors.E(op, err)
	case len(r) == 0:
		return nil, errors.E(op, common.ErrNotFound)
	}

	isVerifiedStr := r[userIsVerifiedField]

	return &models.User{
		ID:           userId,
		Username:     r[userUsernameField],
		PasswordHash: r[userPasswordHashField],
		FirstName:    r[userFirstNameField],
		LastName:     r[userLastNameField],
		Email:        r[userEmailField],
		IsVerified:   isVerifiedStr == "1",
	}, nil
}

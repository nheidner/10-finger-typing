package redis_repo

import (
	"10-typing/models"
	"10-typing/repositories"
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
func (repo *RedisRepository) GetUserByEmailInCacheOrDB(ctx context.Context, dbRepo repositories.DBRepository, email string) (*models.User, error) {
	userEmailKey := getUserEmailKey(email)

	userIdStr, err := repo.redisClient.Get(ctx, userEmailKey).Result()
	if err != nil {
		return nil, err
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, err
	}

	return repo.GetUserByIdInCacheOrDB(ctx, dbRepo, userId)
}

// if not found, queries db
func (repo *RedisRepository) GetUserByIdInCacheOrDB(ctx context.Context, dbRepo repositories.DBRepository, userId uuid.UUID) (*models.User, error) {
	user, err := repo.getUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	user, err = dbRepo.FindUserById(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	if err = repo.SetUser(ctx, *user); err != nil {
		return nil, err
	}

	return user, nil
}

// read: read first in cache, if not exists, read from db and write to cache (in case, keys got evicted)
func (repo *RedisRepository) GetUserBySessionTokenHashInCacheOrDB(
	ctx context.Context,
	dbRepo repositories.DBRepository,
	tokenHash string,
) (*models.User, error) {
	userId, err := repo.getUserIdBySessionTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if userId == uuid.Nil {
		return nil, nil
	}

	user, err := repo.getUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	user, err = dbRepo.FindUserById(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	if err = repo.SetUser(ctx, *user); err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *RedisRepository) UserExists(ctx context.Context, userId uuid.UUID) (bool, error) {
	userKey := getUserKey(userId)

	r, err := repo.redisClient.Exists(ctx, userKey).Result()

	return r > 0, err
}

func (repo *RedisRepository) SetUser(ctx context.Context, user models.User) error {
	userEmailKey := getUserEmailKey(user.Email)

	if err := repo.redisClient.Set(ctx, userEmailKey, user.ID.String(), 0).Err(); err != nil {
		return err
	}

	userKey := getUserKey(user.ID)

	return repo.redisClient.HSet(ctx, userKey, map[string]any{
		userUsernameField:     user.Username,
		userPasswordHashField: user.PasswordHash,
		userFirstNameField:    user.FirstName,
		userLastNameField:     user.LastName,
		userEmailField:        user.Email,
		userIsVerifiedField:   user.IsVerified,
	}).Err()
}

func (repo *RedisRepository) VerifyUser(ctx context.Context, userId uuid.UUID) error {
	userKey := getUserKey(userId)

	return repo.redisClient.HSet(ctx, userKey, userIsVerifiedField, true).Err()
}

func (repo *RedisRepository) DeleteAllUsers(ctx context.Context) error {
	if err := deleteKeysByPattern(ctx, repo, "users:*"); err != nil {
		return err
	}

	if err := deleteKeysByPattern(ctx, repo, "user_emails:*"); err != nil {
		return err
	}

	return nil
}

func (repo *RedisRepository) getUserIdBySessionTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	sessionKey := getSessionKey(tokenHash)

	userIdStr, err := repo.redisClient.Get(ctx, sessionKey).Result()
	switch {
	case err == redis.Nil:
		return uuid.Nil, nil
	case err != nil:
		return uuid.Nil, err
	}

	return uuid.Parse(userIdStr)
}

func (repo *RedisRepository) getUser(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	userKey := getUserKey(userId)

	r, err := repo.redisClient.HGetAll(ctx, userKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
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

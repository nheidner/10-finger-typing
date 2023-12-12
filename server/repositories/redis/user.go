package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// if not found, queries db
func (repo *RedisRepository) GetUserByEmailInCacheOrDB(ctx context.Context, dbRepo common.DBRepository, email string) (*models.User, error) {
	const op errors.Op = "redis_repo.RedisRepository.GetUserByEmailInCacheOrDB"
	var userEmailKey = getUserEmailKey(email)
	var cmd redis.Cmdable = repo.redisClient

	userIdStr, err := cmd.Get(ctx, userEmailKey).Result()
	switch {
	case err == redis.Nil:
		return nil, errors.E(op, common.ErrNotFound)
	case err != nil:
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

	user, err = dbRepo.FindUserById(ctx, nil, userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = repo.SetUser(ctx, nil, *user); err != nil {
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

	user, err = dbRepo.FindUserById(ctx, nil, userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = repo.SetUser(ctx, nil, *user); err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

func (repo *RedisRepository) UserExists(ctx context.Context, userId uuid.UUID) (bool, error) {
	const op errors.Op = "redis_repo.RedisRepository.UserExists"
	var userKey = getUserKey(userId)
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.Exists(ctx, userKey).Result()
	if err != nil {
		errors.E(op, err)
	}

	return r > 0, nil
}

func (repo *RedisRepository) SetUser(ctx context.Context, tx common.Transaction, user models.User) error {
	const op errors.Op = "redis_repo.RedisRepository.SetUser"
	userEmailKey := getUserEmailKey(user.Email)
	var userKey = getUserKey(user.ID)

	// PIPELINE start if no outer pipeline exists
	cmd, innerTx := repo.beginPipelineIfNoOuterTransactionExists(tx)

	cmd.Set(ctx, userEmailKey, user.ID.String(), 0)

	cmd.HSet(ctx, userKey, map[string]any{
		userUsernameField:     user.Username,
		userPasswordHashField: user.PasswordHash,
		userFirstNameField:    user.FirstName,
		userLastNameField:     user.LastName,
		userEmailField:        user.Email,
		userIsVerifiedField:   user.IsVerified,
	})

	// PIPELINE commit
	if innerTx != nil {
		if err := innerTx.Commit(ctx); err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (repo *RedisRepository) VerifyUser(ctx context.Context, tx common.Transaction, userId uuid.UUID) error {
	const op errors.Op = "redis_repo.RedisRepository.VerifyUser"
	var userKey = getUserKey(userId)
	var cmd = repo.cmdable(tx)

	if err := cmd.HSet(ctx, userKey, userIsVerifiedField, true).Err(); err != nil {
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
	var cmd redis.Cmdable = repo.redisClient

	userIdStr, err := cmd.Get(ctx, sessionKey).Result()
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
	var userKey = getUserKey(userId)
	var cmd redis.Cmdable = repo.redisClient

	r, err := cmd.HGetAll(ctx, userKey).Result()
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

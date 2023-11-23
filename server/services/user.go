package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"10-typing/rand"
	"10-typing/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	minBytesPerToken = 32
	userPwPepper     = "secret-random-string"
)

type UserService struct {
	dbRepo               common.DBRepository
	cacheRepo            common.CacheRepository
	sessionBytesPerToken int
}

func NewUserService(dbRepo common.DBRepository, cacheRepo common.CacheRepository, sessionBytesPerToken int) *UserService {
	return &UserService{dbRepo, cacheRepo, sessionBytesPerToken}
}

func (us *UserService) FindUsers(ctx context.Context, username, usernameSubstr string) ([]models.User, error) {
	const op errors.Op = "services.UserService.FindUsers"

	users, err := us.dbRepo.FindUsers(username, usernameSubstr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return users, nil
}

func (us *UserService) FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	const op errors.Op = "services.UserService.FindUserById"

	user, err := us.dbRepo.FindUserById(userId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

func (us *UserService) Create(ctx context.Context, email, username, firstName, lastName, password string) (*models.User, error) {
	const op errors.Op = "services.UserService.Create"

	hashedPassword, err := us.hashedPassword(password)
	if err != nil {
		return nil, errors.E(op, err)
	}

	newUser := models.User{
		Email:        email,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		IsVerified:   false,
		PasswordHash: hashedPassword,
	}

	user, err := us.dbRepo.CreateUserAndCache(us.cacheRepo, newUser)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return user, nil
}

func (us *UserService) VerifyUser(ctx context.Context, userId uuid.UUID) error {
	const op errors.Op = "services.UserService.VerifyUser"

	if err := us.dbRepo.VerifyUserAndCache(us.cacheRepo, userId); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (us *UserService) Login(ctx context.Context, email, password string) (user *models.User, sessionToken string, err error) {
	const op errors.Op = "services.UserService.Login"

	user, err = us.dbRepo.FindUserByEmail(email)
	switch {
	case errors.Is(err, common.ErrNotFound):
		return nil, "", errors.E(op, err, http.StatusBadRequest)
	case err != nil:
		return nil, "", errors.E(op, err)
	case !user.IsVerified:
		err := fmt.Errorf("user not verified")
		return nil, "", errors.E(op, err, http.StatusBadRequest, errors.Messages{"message": "user not verified"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		return nil, "", errors.E(op, err, http.StatusBadRequest, errors.Messages{"message": "password is not correct"})
	}

	token, err := us.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	return user, token, nil
}

func (us *UserService) DeleteSession(ctx context.Context, token string) error {
	const op errors.Op = "services.UserService.DeleteSession"

	tokenHash := utils.HashSessionToken(token)

	if err := us.cacheRepo.DeleteSession(ctx, tokenHash); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (us *UserService) hashedPassword(password string) (hashedPassword string, err error) {
	const op errors.Op = "services.UserService.hashedPassword"

	pwBytes := []byte(password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", errors.E(op, err)
	}

	return string(hashedBytes), nil
}

func (us *UserService) createSession(ctx context.Context, userId uuid.UUID) (token string, err error) {
	const op errors.Op = "services.UserService.createSession"

	bytesPerToken := us.sessionBytesPerToken
	if bytesPerToken < minBytesPerToken {
		bytesPerToken = minBytesPerToken
	}
	token, err = rand.String(bytesPerToken)
	if err != nil {
		return "", errors.E(op, err)
	}

	tokenHash := utils.HashSessionToken(token)

	if err := us.cacheRepo.SetSession(ctx, tokenHash, userId); err != nil {
		return "", errors.E(op, err)
	}

	return token, nil
}

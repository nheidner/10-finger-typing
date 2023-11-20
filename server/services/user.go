package services

import (
	"10-typing/models"
	"10-typing/rand"
	"10-typing/repositories"
	"10-typing/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	minBytesPerToken = 32
	userPwPepper     = "secret-random-string"
)

type UserService struct {
	dbRepo               repositories.DBRepository
	cacheRepo            repositories.CacheRepository
	sessionBytesPerToken int
}

func NewUserService(dbRepo repositories.DBRepository, cacheRepo repositories.CacheRepository, sessionBytesPerToken int) *UserService {
	return &UserService{dbRepo, cacheRepo, sessionBytesPerToken}
}

func (us *UserService) FindUsers(username, usernameSubstr string) ([]models.User, error) {
	return us.dbRepo.FindUsers(username, usernameSubstr)
}

func (us *UserService) FindUserById(userId uuid.UUID) (*models.User, error) {
	return us.dbRepo.FindUserById(userId)
}

func (us *UserService) Create(email, username, firstName, lastName, password string) (*models.User, error) {
	hashedPassword, err := us.hashedPassword(password)
	if err != nil {
		return nil, err
	}

	newUser := models.User{
		Email:        email,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		IsVerified:   false,
		PasswordHash: hashedPassword,
	}

	return us.dbRepo.CreateUserAndCache(us.cacheRepo, newUser)
}

func (us *UserService) VerifyUser(userId uuid.UUID) error {
	return us.dbRepo.VerifyUserAndCache(us.cacheRepo, userId)
}

func (us *UserService) Login(email, password string) (user *models.User, sessionToken string, err error) {
	user, err = us.dbRepo.FindUserByEmail(email)
	switch {
	case err != nil:
		return nil, "", err
	case user == nil:
		return nil, "", errors.New("user not found")
	case !user.IsVerified:
		return nil, "", errors.New("user not verified")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		return nil, "", errors.New("invalid password: " + err.Error())
	}

	token, err := us.createSession(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (us *UserService) DeleteSession(token string) error {
	tokenHash := utils.HashSessionToken(token)

	return us.cacheRepo.DeleteSession(context.Background(), tokenHash)
}

func (us *UserService) hashedPassword(password string) (hashedPassword string, err error) {
	pwBytes := []byte(password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func (us *UserService) createSession(userId uuid.UUID) (token string, err error) {
	bytesPerToken := us.sessionBytesPerToken
	if bytesPerToken < minBytesPerToken {
		bytesPerToken = minBytesPerToken
	}
	token, err = rand.String(bytesPerToken)
	if err != nil {
		return "", err
	}

	tokenHash := utils.HashSessionToken(token)

	return token, us.cacheRepo.SetSession(context.Background(), tokenHash, userId)
}

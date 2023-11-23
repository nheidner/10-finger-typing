package common

import (
	"10-typing/models"
	"context"

	"github.com/google/uuid"
)

type DBRepository interface {
	BeginTx() Transaction
	RoomDBRepository
	ScoreDBRepository
	TextDBRepository
	TokenDBRepository
	UserDBRepository
	UserRoomDBRepository
}

type RoomDBRepository interface {
	FindRoomWithUsers(roomId uuid.UUID) (*models.Room, error)
	FindRoom(roomId uuid.UUID) (*models.Room, error)
	CreateRoom(newRoom models.Room) (*models.Room, error)
	SoftDeleteRoom(roomId uuid.UUID) error
	DeleteAllRooms() error
}

type ScoreDBRepository interface {
	FindScores(userId, gameId uuid.UUID, username string, sortOptions []models.SortOption) ([]models.Score, error)
	CreateScore(score models.Score) (*models.Score, error)
	DeleteAllScores() error
}

type TextDBRepository interface {
	FindNewTextForUser(
		userId uuid.UUID, language string,
		punctuation bool,
		specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
	) (*models.Text, error)
	FindAllTextIds() ([]uuid.UUID, error)
	FindTextById(textId uuid.UUID) (*models.Text, error)
	CreateTextAndCache(ctx context.Context, cacheRepo CacheRepository, text models.Text) (*models.Text, error)
}

type TokenDBRepository interface {
	CreateToken(roomId uuid.UUID) (*models.Token, error)
}

type UserDBRepository interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUsers(username, usernameSubstr string) ([]models.User, error)
	FindUserById(userId uuid.UUID) (*models.User, error)
	CreateUserAndCache(cacheRepo CacheRepository, newUser models.User) (*models.User, error)
	VerifyUserAndCache(cacheRepo CacheRepository, userId uuid.UUID) error
	DeleteAllUsers() error
}

type UserRoomDBRepository interface {
	CreateUserRoom(userId, roomId uuid.UUID) error
}

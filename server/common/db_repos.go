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
	FindRoomWithUsers(ctx context.Context, roomId uuid.UUID) (*models.Room, error)
	FindRoom(ctx context.Context, roomId uuid.UUID) (*models.Room, error)
	CreateRoom(ctx context.Context, newRoom models.Room) (*models.Room, error)
	SoftDeleteRoom(ctx context.Context, roomId uuid.UUID) error
	DeleteAllRooms(ctx context.Context) error
}

type ScoreDBRepository interface {
	FindScores(ctx context.Context, userId, gameId uuid.UUID, username string, sortOptions []models.SortOption) ([]models.Score, error)
	CreateScore(ctx context.Context, score models.Score) (*models.Score, error)
	DeleteAllScores(ctx context.Context) error
}

type TextDBRepository interface {
	FindNewTextForUser(ctx context.Context,
		userId uuid.UUID, language string,
		punctuation bool,
		specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
	) (*models.Text, error)
	FindAllTextIds(ctx context.Context) ([]uuid.UUID, error)
	FindTextById(ctx context.Context, textId uuid.UUID) (*models.Text, error)
	CreateTextAndCache(ctx context.Context, cacheRepo CacheRepository, text models.Text) (*models.Text, error)
}

type TokenDBRepository interface {
	CreateToken(ctx context.Context, roomId uuid.UUID) (*models.Token, error)
}

type UserDBRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUsers(ctx context.Context, username, usernameSubstr string) ([]models.User, error)
	FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	CreateUserAndCache(ctx context.Context, cacheRepo CacheRepository, newUser models.User) (*models.User, error)
	VerifyUserAndCache(ctx context.Context, cacheRepo CacheRepository, userId uuid.UUID) error
	DeleteAllUsers(ctx context.Context) error
}

type UserRoomDBRepository interface {
	CreateUserRoom(ctx context.Context, userId, roomId uuid.UUID) error
}

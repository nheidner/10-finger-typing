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
	FindRoomWithUsers(ctx context.Context, tx Transaction, roomId uuid.UUID) (*models.Room, error)
	FindRoom(ctx context.Context, tx Transaction, roomId uuid.UUID) (*models.Room, error)
	CreateRoom(ctx context.Context, tx Transaction, newRoom models.Room) (*models.Room, error)
	SoftDeleteRoom(ctx context.Context, tx Transaction, roomId uuid.UUID) error
	DeleteAllRooms(ctx context.Context, tx Transaction) error
}

type ScoreDBRepository interface {
	FindScores(ctx context.Context, tx Transaction, userId, gameId uuid.UUID, username string, sortOptions []models.SortOption) ([]models.Score, error)
	CreateScore(ctx context.Context, tx Transaction, score models.Score) (*models.Score, error)
	DeleteAllScores(ctx context.Context, tx Transaction) error
}

type TextDBRepository interface {
	FindNewTextForUser(ctx context.Context,
		tx Transaction,
		userId uuid.UUID, language string,
		punctuation bool,
		specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
	) (*models.Text, error)
	FindAllTextIds(ctx context.Context, tx Transaction) ([]uuid.UUID, error)
	FindTextById(ctx context.Context, tx Transaction, textId uuid.UUID) (*models.Text, error)
	CreateTextAndCache(ctx context.Context, tx Transaction, cacheRepo CacheRepository, text models.Text) (*models.Text, error)
}

type TokenDBRepository interface {
	CreateToken(ctx context.Context, tx Transaction, roomId uuid.UUID) (*models.Token, error)
}

type UserDBRepository interface {
	FindUserByEmail(ctx context.Context, tx Transaction, email string) (*models.User, error)
	FindUsers(ctx context.Context, tx Transaction, username, usernameSubstr string) ([]models.User, error)
	FindUserById(ctx context.Context, tx Transaction, userId uuid.UUID) (*models.User, error)
	CreateUserAndCache(ctx context.Context, tx Transaction, cacheRepo CacheRepository, newUser models.User) (*models.User, error)
	VerifyUserAndCache(ctx context.Context, tx Transaction, cacheRepo CacheRepository, userId uuid.UUID) error
	DeleteAllUsers(ctx context.Context, tx Transaction) error
}

type UserRoomDBRepository interface {
	CreateUserRoom(ctx context.Context, tx Transaction, userId, roomId uuid.UUID) error
}

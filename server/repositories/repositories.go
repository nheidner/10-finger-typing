package repositories

import (
	"10-typing/errors"
	"10-typing/models"
	"context"

	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
)

type Transaction interface {
	Commit()
	Rollback()
}

type DBRepository interface {
	BeginTx() Transaction
	RoomDBRepository
	ScoreDBRepository
	TextDBRepository
	TokenDBRepository
	UserDBRepository
	UserRoomDBRepository
}

type CacheRepository interface {
	GameCacheRepository
	RoomCacheRepository
	RoomStreamCacheRepository
	RoomSubscriberCacheRepository
	TextCacheRepository
	UserNotificationCacheRepository
	UserCacheRepository
	SessionCacheRepository
}

type OpenAiRepository interface {
	GenerateTypingText(language string, punctuation bool, specialCharacters, numbers int) (string, error)
}

type EmailTransactionRepository interface {
	InviteNewUserToRoom(email string, token uuid.UUID) error
	InviteUserToRoom(email, username string) error
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

type GameCacheRepository interface {
	GetCurrentGameUserIds(ctx context.Context, roomId uuid.UUID) ([]uuid.UUID, error)
	GetCurrentGameUsersNumber(ctx context.Context, roomId uuid.UUID) (int, error)
	GetCurrentGameStatus(ctx context.Context, roomId uuid.UUID) (models.GameStatus, error)
	GetCurrentGameId(ctx context.Context, roomId uuid.UUID) (uuid.UUID, error)
	GetCurrentGame(ctx context.Context, roomId uuid.UUID) (*models.Game, error)
	SetNewCurrentGame(ctx context.Context, newGameId, textId, roomId uuid.UUID, userIds ...uuid.UUID) error
	SetGameUser(ctx context.Context, roomId, userId uuid.UUID) error
	SetCurrentGameStatus(ctx context.Context, roomId uuid.UUID, gameStatus models.GameStatus) error
	IsCurrentGame(ctx context.Context, roomId, gameId uuid.UUID) (bool, error)
	IsCurrentGameUser(ctx context.Context, roomId, userId uuid.UUID) (bool, error)
}

type RoomCacheRepository interface {
	GetRoomInCacheOrDb(ctx context.Context, dbRepo DBRepository, roomId uuid.UUID) (*models.Room, error)
	GetRoomGameDurationSec(ctx context.Context, roomId uuid.UUID) (gameDurationSec int, err error)
	SetRoom(ctx context.Context, room models.Room) error
	RoomHasAdmin(ctx context.Context, roomId, adminId uuid.UUID) (bool, error)
	RoomHasSubscribers(ctx context.Context, roomId uuid.UUID, userIds ...uuid.UUID) (bool, error)
	RoomExists(ctx context.Context, roomId uuid.UUID) (bool, error)
	DeleteRoom(ctx context.Context, roomId uuid.UUID) error
	DeleteAllRooms(ctx context.Context) error
}

type RoomStreamCacheRepository interface {
	// call PublishPushMessage with type and payload and not with push message type
	PublishPushMessage(ctx context.Context, roomId uuid.UUID, pushMessage models.PushMessage) error
	PublishAction(ctx context.Context, roomId uuid.UUID, action models.StreamActionType) error
	GetPushMessages(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[[]byte]
	GetAction(ctx context.Context, roomId uuid.UUID, startTime time.Time) <-chan models.StreamSubscriptionResult[models.StreamActionType]
}

type RoomSubscriberCacheRepository interface {
	GetRoomSubscriberStatus(ctx context.Context, roomId, userId uuid.UUID) (numberRoomSubscriberConns int64, roomSubscriberStatusHasBeenUpdated bool, err error)
	GetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID) (models.SubscriberGameStatus, error)
	GetRoomSubscribers(ctx context.Context, roomId uuid.UUID) ([]models.RoomSubscriber, error)
	SetRoomSubscriberGameStatus(ctx context.Context, roomId, userId uuid.UUID, status models.SubscriberGameStatus) error
	SetRoomSubscriberConnection(ctx context.Context, roomId, userId, newConnectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error)
	DeleteRoomSubscriber(ctx context.Context, roomId, userId uuid.UUID) error
	DeleteRoomSubscriberConnection(ctx context.Context, roomId, userId, connectionId uuid.UUID) (roomSubscriberStatusHasBeenUpdated bool, err error)
}

type TextCacheRepository interface {
	SetTextId(ctx context.Context, textIds ...uuid.UUID) error
	TextIdsKeyExists(ctx context.Context) (bool, error)
	TextIdExists(ctx context.Context, textId uuid.UUID) (bool, error)
	DeleteTextIdsKey(ctx context.Context) error
}

type UserNotificationCacheRepository interface {
	PublishUserNotification(ctx context.Context, userId uuid.UUID, userNotification models.UserNotification) error
	GetUserNotification(ctx context.Context, userId uuid.UUID, startId string) chan models.StreamSubscriptionResult[*models.UserNotification]
}

type UserCacheRepository interface {
	GetUserByEmailInCacheOrDB(ctx context.Context, dbRepo DBRepository, email string) (*models.User, error)
	GetUserByIdInCacheOrDB(ctx context.Context, dbRepo DBRepository, userId uuid.UUID) (*models.User, error)
	GetUserBySessionTokenHashInCacheOrDB(ctx context.Context, dbRepo DBRepository, tokenHash string) (*models.User, error)
	UserExists(ctx context.Context, userId uuid.UUID) (bool, error)
	SetUser(ctx context.Context, user models.User) error
	VerifyUser(ctx context.Context, userId uuid.UUID) error
	DeleteAllUsers(ctx context.Context) error
}

type SessionCacheRepository interface {
	SetSession(ctx context.Context, tokenHash string, userId uuid.UUID) error
	DeleteSession(ctx context.Context, tokenHash string) error
	DeleteAllSessions(ctx context.Context) error
}

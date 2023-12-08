package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"

	"context"
	"fmt"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type RoomService struct {
	dbRepo               common.DBRepository
	cacheRepo            common.CacheRepository
	emailTransactionRepo common.EmailTransactionRepository
	logger               common.Logger
}

func NewRoomService(
	dbRepo common.DBRepository,
	cacheRepo common.CacheRepository,
	emailTransactionRepo common.EmailTransactionRepository,
	logger common.Logger,
) *RoomService {
	return &RoomService{
		dbRepo,
		cacheRepo,
		emailTransactionRepo,
		logger,
	}
}

func (rs *RoomService) CreateRoom(ctx context.Context, userIds []uuid.UUID, emails []string, gameDurationSec int, authenticatedUser models.User) (*models.Room, error) {
	const op errors.Op = "services.RoomService.CreateRoom"

	// validate
	if (len(userIds) == 0) && (len(emails) == 0) {
		err := fmt.Errorf("you cannot create a room just for yourself")
		return nil, errors.E(op, err, http.StatusBadRequest)
	}

	for _, userId := range userIds {
		if userId == authenticatedUser.ID {
			err := fmt.Errorf("you cannot create a room for yourself with yourself")
			return nil, errors.E(op, err, http.StatusBadRequest)
		}
	}

	for _, email := range emails {
		if email == authenticatedUser.Email {
			err := fmt.Errorf("you cannot create a room for yourself with yourself")
			return nil, errors.E(op, err, http.StatusBadRequest)
		}
	}

	// userIds are automatically validated when rows in user_room table are created
	userIds = append(userIds, authenticatedUser.ID)

	// validate that emails are not already existing users
	var allEmails []string

	for _, email := range emails {
		user, err := rs.cacheRepo.GetUserByEmailInCacheOrDB(ctx, rs.dbRepo, email)
		if err != nil {
			return nil, errors.E(op, err, http.StatusInternalServerError)
		}

		if user == nil {
			allEmails = append(allEmails, email)
			continue
		}

		userIds = append(userIds, user.ID)
	}

	tx := rs.dbRepo.BeginTx()

	// create room
	room, err := rs.createRoomWithSubscribers(ctx, tx, userIds, allEmails, authenticatedUser.ID, gameDurationSec)
	if err != nil {
		err := errors.E(op, err, http.StatusInternalServerError)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, errors.E(op, rollbackErr)
		}

		return nil, err
	}

	// send notifications
	for _, roomSubscriber := range room.Users {
		if roomSubscriber.ID == authenticatedUser.ID {
			continue
		}

		userNotification := models.UserNotification{
			Type: models.RoomInvitation,
			Payload: map[string]any{
				"by":     authenticatedUser.Username,
				"roomId": room.ID,
			},
		}

		if err = rs.cacheRepo.PublishUserNotification(ctx, roomSubscriber.ID, userNotification); err != nil {
			err := errors.E(op, err, http.StatusInternalServerError)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, errors.E(op, rollbackErr)
			}

			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E(op, err)
	}

	// create tokens and send invites to non registered users
	for _, email := range emails {
		token, err := rs.dbRepo.CreateToken(ctx, room.ID)
		if err != nil {
			return nil, errors.E(op, err)
		}

		err = rs.emailTransactionRepo.InviteNewUserToRoom(email, token.ID)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// send invites to registered users
	for _, roomSubscriber := range room.Users {
		if roomSubscriber.ID == authenticatedUser.ID {
			continue
		}

		err = rs.emailTransactionRepo.InviteUserToRoom(roomSubscriber.Email, roomSubscriber.Username)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	return room, nil
}

func (rs *RoomService) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "services.RoomService.DeleteRoom"

	if err := rs.dbRepo.SoftDeleteRoom(ctx, roomId); err != nil {
		return errors.E(op, err)
	}

	if err := rs.cacheRepo.DeleteRoom(ctx, roomId); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (rs *RoomService) LeaveRoom(ctx context.Context, roomId, userId uuid.UUID) error {
	const op errors.Op = "services.RoomService.LeaveRoom"

	isAdmin, err := rs.cacheRepo.RoomHasAdmin(ctx, roomId, userId)
	if err != nil {
		return errors.E(op, err)
	}

	if isAdmin {
		// first need to send terminate action message so that all websocket that remained connected, disconnect
		if err := rs.cacheRepo.PublishAction(ctx, roomId, models.TerminateAction); err != nil {
			return errors.E(op, err)
		}

		if err := rs.DeleteRoom(ctx, roomId); err != nil {
			return errors.E(op, err)
		}

		return nil
	}

	if err = rs.cacheRepo.DeleteRoomSubscriber(ctx, roomId, userId); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// reads from connection and handles incoming ping and cursor messages.
//
// gets initial_state data and sends it as message to client.
//
// subscribes to room redis stream and sends messages to client.
func (rs *RoomService) RoomConnect(ctx context.Context, c *gin.Context, roomId uuid.UUID, user *models.User) error {
	const op errors.Op = "services.RoomService.RoomConnect"

	room, err := rs.cacheRepo.GetRoomInCacheOrDb(ctx, rs.dbRepo, roomId)
	switch {
	case errors.Is(err, common.ErrNotFound):
		return errors.E(op, err, http.StatusNotFound)
	case err != nil:
		return errors.E(op, err, http.StatusInternalServerError)
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return errors.E(op, err, http.StatusBadRequest)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	roomSubscription := newRoomSubscription(conn, room.ID, user.ID, rs.cacheRepo, rs.logger)
	defer roomSubscription.close(ctx)

	timeStamp := time.Now()
	errCh := make(chan error)

	go func() {
		err := roomSubscription.handleMessages(ctx)
		rs.logger.Error(errors.E(op, err))

		select {
		case errCh <- err:
		default:
		}
	}()

	go func() {
		err := roomSubscription.handleRoomSubscriberStatus(ctx)

		if err != nil {
			rs.logger.Error(errors.E(op, err))

			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		err := roomSubscription.sendInitialState(ctx, *room)

		if err != nil {
			rs.logger.Error(errors.E(op, err))

			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		err := roomSubscription.subscribe(ctx, timeStamp)
		rs.logger.Error(errors.E(op, err))

		select {
		case errCh <- err:
		default:
		}
	}()

	<-errCh

	return nil
}

func (rs *RoomService) createRoomWithSubscribers(ctx context.Context, tx common.Transaction, userIds []uuid.UUID, emails []string, adminId uuid.UUID, gameDurationSec int) (*models.Room, error) {
	const op errors.Op = "services.RoomService.createRoomWithSubscribers"

	newRoom := models.Room{
		AdminId: adminId,
	}
	if gameDurationSec != 0 {
		newRoom.GameDurationSec = gameDurationSec
	}

	createdRoom, err := rs.dbRepo.CreateRoom(ctx, tx, newRoom)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// room subscribers
	for _, userId := range userIds {
		if err := rs.dbRepo.CreateUserRoom(ctx, tx, userId, createdRoom.ID); err != nil {
			return nil, errors.E(op, err)
		}
	}

	createdRoom, err = rs.dbRepo.FindRoom(ctx, tx, createdRoom.ID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err := rs.cacheRepo.SetRoom(ctx, *createdRoom); err != nil {
		return nil, errors.E(op, err)
	}

	return createdRoom, nil
}

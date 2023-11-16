package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type RoomService struct {
	dbRepo               repositories.DBRepository
	cacheRepo            repositories.CacheRepository
	emailTransactionRepo repositories.EmailTransactionRepository
}

func NewRoomService(
	dbRepo repositories.DBRepository,
	cacheRepo repositories.CacheRepository,
	emailTransactionRepo repositories.EmailTransactionRepository,
) *RoomService {
	return &RoomService{
		dbRepo,
		cacheRepo,
		emailTransactionRepo,
	}
}

func (rs *RoomService) Find(roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	var ctx = context.Background()

	room, err := rs.cacheRepo.GetRoom(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	if room == nil {
		room, err = rs.dbRepo.FindRoomByUser(roomId, userId)
		if err != nil {
			return nil, err
		}

		if err = rs.cacheRepo.SetRoom(ctx, *room); err != nil {
			// no error should be returned
			log.Println(err)
		}
	}

	return room, nil
}

func (rs *RoomService) CreateRoom(userIds []uuid.UUID, emails []string, authenticatedUser models.User) (*models.Room, error) {
	// validate
	if (len(userIds) == 0) && (len(emails) == 0) {
		return nil, fmt.Errorf("you cannot create a room just for yourself")
	}

	for _, userId := range userIds {
		if userId == authenticatedUser.ID {
			return nil, fmt.Errorf("you cannot create a room for yourself with yourself")
		}
	}

	for _, email := range emails {
		if email == authenticatedUser.Email {
			return nil, fmt.Errorf("you cannot create a room for yourself with yourself")
		}
	}

	userIds = append(userIds, authenticatedUser.ID)

	// TODO: validate userIds

	// validate that emails are not already existing users
	var allEmails []string

	for _, email := range emails {
		user, err := rs.dbRepo.FindUserByEmail(email)
		if err != nil {
			return nil, err
		}

		if user == nil {
			allEmails = append(allEmails, email)
			continue
		}

		userIds = append(userIds, user.ID)
	}

	// create room
	room, err := rs.createRoomWithSubscribers(userIds, allEmails, authenticatedUser.ID)
	if err != nil {
		return nil, err
	}

	// create tokens and send invites to non registered users
	for _, email := range emails {
		token, err := rs.dbRepo.CreateToken(room.ID)
		if err != nil {
			return nil, err
		}

		err = rs.emailTransactionRepo.InviteNewUserToRoom(email, token.ID)
		if err != nil {
			return nil, err
		}
	}

	// send invites to registered users
	for _, roomSubscriber := range room.Users {
		if roomSubscriber.ID == authenticatedUser.ID {
			continue
		}

		err = rs.emailTransactionRepo.InviteUserToRoom(roomSubscriber.Email, roomSubscriber.Username)
		if err != nil {
			return nil, err
		}
	}

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

		if err = rs.cacheRepo.PublishUserNotification(context.Background(), roomSubscriber.ID, userNotification); err != nil {
			return nil, err
		}
	}

	return room, nil
}

func (rs *RoomService) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	if err := rs.dbRepo.SoftDeleteRoom(roomId); err != nil {
		return err
	}

	return rs.cacheRepo.DeleteRoom(ctx, roomId)
}

func (rs *RoomService) LeaveRoom(roomId, userId uuid.UUID) error {
	var ctx = context.Background()

	isAdmin, err := rs.cacheRepo.RoomHasAdmin(ctx, roomId, userId)
	if err != nil {
		return err
	}

	if isAdmin {
		// first need to send terminate action message so that all websocket that remained connected, disconnect
		if err := rs.cacheRepo.PublishAction(ctx, roomId, models.TerminateAction); err != nil {
			log.Println("terminate action failed:", err)
			return err
		}

		if err := rs.DeleteRoom(ctx, roomId); err != nil {
			log.Println("failed to remove room subscriber:", err)
			return err
		}

		return nil
	}

	if err = rs.cacheRepo.DeleteRoomSubscriber(ctx, roomId, userId); err != nil {
		log.Println("failed to remove room subscriber:", err)
		return err
	}

	return nil
}

// reads from connection and handles incoming ping and cursor messages.
//
// gets initial_state data and sends it as message to client.
//
// subscribes to room redis stream and sends messages to client.
func (rs *RoomService) RoomConnect(ctx context.Context, userId uuid.UUID, room *models.Room, conn *websocket.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	roomSubscription := newRoomSubscription(conn, room.ID, userId, rs.cacheRepo)
	defer roomSubscription.close(ctx)

	timeStamp := time.Now()
	errCh := make(chan error)

	go func() {
		err := roomSubscription.handleMessages(ctx)
		if err != nil {
			err = errors.New("error handling messages " + err.Error())
			log.Println(err)
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		err := roomSubscription.handleRoomSubscriberStatus(ctx)
		if err != nil {
			err = errors.New("error room subscriber status " + err.Error())
			log.Println(err)
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		err := roomSubscription.sendInitialState(ctx, *room)
		if err != nil {
			err = errors.New("error sending initial state " + err.Error())
			log.Println(err)
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		err := roomSubscription.subscribe(ctx, timeStamp)
		if err != nil {
			err = errors.New("error subscribing " + err.Error())
			log.Println(err)
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	err := <-errCh

	return err
}

func (rs *RoomService) createRoomWithSubscribers(userIds []uuid.UUID, emails []string, adminId uuid.UUID) (*models.Room, error) {
	createdRoom, err := rs.dbRepo.CreateRoom(models.Room{
		AdminId: adminId,
	})
	if err != nil {
		return nil, err
	}

	// room subscribers
	for _, userId := range userIds {
		if err := rs.dbRepo.CreateUserRoom(userId, createdRoom.ID); err != nil {
			return nil, err
		}
	}

	createdRoom, err = rs.dbRepo.FindRoom(createdRoom.ID)
	if err != nil {
		return nil, err
	}

	if err := rs.cacheRepo.SetRoom(context.Background(), *createdRoom); err != nil {
		return nil, err
	}

	return createdRoom, nil
}

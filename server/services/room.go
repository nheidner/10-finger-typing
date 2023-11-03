package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type RoomService struct {
	roomDbRepo              *repositories.RoomDbRepository
	roomRedisRepo           *repositories.RoomRedisRepository
	userRoomDbRepo          *repositories.UserRoomDbRepository
	roomStreamRedisRepo     *repositories.RoomStreamRedisRepository
	roomSubscriberRedisRepo *repositories.RoomSubscriberRedisRepository
	userDbRepo              *repositories.UserDbRepository
	tokenDbRepo             *repositories.TokenDbRepository
	emailTransactionRepo    *repositories.EmailTransactionRepository
	gameRedisRpo            *repositories.GameRedisRepository
}

func NewRoomService(
	roomDbRepo *repositories.RoomDbRepository,
	roomRedisRepo *repositories.RoomRedisRepository,
	userRoomDbRepo *repositories.UserRoomDbRepository,
	roomStreamRedisRepo *repositories.RoomStreamRedisRepository,
	roomSubscriberRedisRepo *repositories.RoomSubscriberRedisRepository,
	userDbRepo *repositories.UserDbRepository,
	tokenDbRepo *repositories.TokenDbRepository,
	emailTransactionRepo *repositories.EmailTransactionRepository,
	gameRedisRpo *repositories.GameRedisRepository,
) *RoomService {
	return &RoomService{
		roomDbRepo,
		roomRedisRepo,
		userRoomDbRepo,
		roomStreamRedisRepo,
		roomSubscriberRedisRepo,
		userDbRepo,
		tokenDbRepo,
		emailTransactionRepo,
		gameRedisRpo,
	}
}

func (rs *RoomService) Find(roomId uuid.UUID, userId uuid.UUID) (*models.Room, error) {
	var ctx = context.Background()

	room, err := rs.roomRedisRepo.FindInRedis(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	if room == nil {
		room, err = rs.roomDbRepo.FindInDb(roomId, userId)
		if err != nil {
			return nil, err
		}

		if err = rs.roomRedisRepo.CreateRoomInRedis(ctx, *room); err != nil {
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
		user, err := rs.userDbRepo.FindByEmail(email)
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
		token, err := rs.tokenDbRepo.Create(room.ID)
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

	return room, nil
}

func (rs *RoomService) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	if err := rs.roomDbRepo.SoftDeleteRoomFromDB(roomId); err != nil {
		return err
	}

	return rs.roomRedisRepo.DeleteRoomFromRedis(ctx, roomId)
}

func (rs *RoomService) LeaveRoom(roomId, userId uuid.UUID) error {
	var ctx = context.Background()

	isAdmin, err := rs.roomRedisRepo.RoomHasAdmin(ctx, roomId, userId)
	if err != nil {
		return err
	}

	if isAdmin {
		// first need to send terminate action message so that all websocket that remained connected, disconnect
		if err := rs.roomStreamRedisRepo.PublishAction(ctx, roomId, models.TerminateAction); err != nil {
			log.Println("terminate action failed:", err)
			return err
		}

		if err := rs.DeleteRoom(ctx, roomId); err != nil {
			log.Println("failed to remove room subscriber:", err)
			return err
		}

		return nil
	}

	if err = rs.roomSubscriberRedisRepo.RemoveRoomSubscriber(ctx, roomId, userId); err != nil {
		log.Println("failed to remove room subscriber:", err)
		return err
	}

	return nil
}

func (rs *RoomService) RoomConnect(userId uuid.UUID, room *models.Room, conn *websocket.Conn, timeStamp time.Time) error {
	var ctx = context.Background()

	roomSubscription := newRoomSubscription(conn, room.ID, userId, rs.roomSubscriberRedisRepo, rs.roomStreamRedisRepo)
	defer roomSubscription.close(ctx)

	err := roomSubscription.initRoomSubscriber(ctx)
	if err != nil {
		log.Println("Failed to initialise room subscriber:", err)
		return err
	}

	existingRoomSubscribers, err := rs.roomSubscriberRedisRepo.GetRoomSubscribers(ctx, room.ID)
	if err != nil {
		log.Println("Failed to get room subscribers:", err)
		return err
	}

	currentGame, err := rs.gameRedisRpo.GetCurrentGameFromRedis(ctx, room.ID)
	if err != nil {
		log.Println("Failed to get current room:", err)
		return err
	}

	room.Subscribers = existingRoomSubscribers
	room.CurrentGame = currentGame

	err = wsjson.Write(ctx, roomSubscription.conn, room)
	if err != nil {
		log.Println("Failed to initialise room subscriber:", err)
		return err
	}

	err = roomSubscription.subscribe(ctx, timeStamp)
	if err != nil {
		log.Println("Error subscribing to room stream:", err)
	}

	return nil
}

func (rs *RoomService) createRoomWithSubscribers(userIds []uuid.UUID, emails []string, adminId uuid.UUID) (*models.Room, error) {
	var newRoom = &models.Room{
		AdminId: adminId,
	}

	if err := rs.roomDbRepo.Create(newRoom); err != nil {
		return nil, err
	}

	// room subscribers
	for _, userId := range userIds {
		if err := rs.userRoomDbRepo.Create(userId, newRoom.ID); err != nil {
			return nil, err
		}
	}

	newRoom, err := rs.roomDbRepo.FindRoomWithUsers(newRoom.ID)
	if err != nil {
		return nil, err
	}

	if err := rs.roomRedisRepo.CreateRoomInRedis(context.Background(), *newRoom); err != nil {
		return nil, err
	}

	return newRoom, nil
}

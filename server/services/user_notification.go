package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserNotificationService struct {
	cacheRepo repositories.CacheRepository
}

func NewUserNotificationService(cacheRepo repositories.CacheRepository) *UserNotificationService {
	return &UserNotificationService{cacheRepo}
}

func (us *UserNotificationService) FindRealtimeUserNotification(userId uuid.UUID, lastId string) (*models.UserNotification, error) {
	t := time.NewTimer(20 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	userNotificationCh := make(chan *models.UserNotification)
	errorCh := make(chan error)

	defer cancel()
	defer t.Stop()

	go func() {
		userNotification, err := us.cacheRepo.GetUserNotification(ctx, userId, lastId)
		if err != nil {
			select {
			case errorCh <- err:
			default:
			}

			return
		}

		select {
		case userNotificationCh <- userNotification:
		default:
		}
	}()

	select {
	case <-t.C:
		return nil, nil
	case userNotification := <-userNotificationCh:
		// TODO: save user ID in DB
		return userNotification, nil
	case err := <-errorCh:
		return nil, err
	}
}

package services

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

const maxRequestDurationSecs = 20

type UserNotificationService struct {
	cacheRepo repositories.CacheRepository
}

func NewUserNotificationService(cacheRepo repositories.CacheRepository) *UserNotificationService {
	return &UserNotificationService{cacheRepo}
}

func (us *UserNotificationService) FindRealtimeUserNotification(ctx context.Context, userId uuid.UUID, lastId string) (*models.UserNotification, error) {
	const op errors.Op = "services.UserNotificationService.FindRealtimeUserNotification"

	userNotificationResultCh := us.cacheRepo.GetUserNotification(ctx, userId, lastId)

	t := time.NewTimer(maxRequestDurationSecs * time.Second)
	defer t.Stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, errors.E(op, ctx.Err())
	case userNotificationResult := <-userNotificationResultCh:
		if userNotificationResult.Error != nil {
			return nil, errors.E(op, userNotificationResult.Error)
		}

		return userNotificationResult.Value, nil
	case <-t.C:
		return nil, nil
	}
}

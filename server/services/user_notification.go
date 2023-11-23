package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"time"

	"github.com/google/uuid"
)

const maxRequestDurationSecs = 20

type UserNotificationService struct {
	cacheRepo common.CacheRepository
	logger    common.Logger
}

func NewUserNotificationService(cacheRepo common.CacheRepository, logger common.Logger) *UserNotificationService {
	return &UserNotificationService{cacheRepo, logger}
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
		return nil, errors.E(op, common.ErrNotFound)
	}
}

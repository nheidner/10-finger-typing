package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"

	"github.com/google/uuid"
)

type ScoreService struct {
	dbRepo common.DBRepository
	logger common.Logger
}

func NewScoreService(dbRepo common.DBRepository, logger common.Logger) *ScoreService {
	return &ScoreService{dbRepo, logger}
}

func (ss *ScoreService) Create(
	ctx context.Context,
	gameId, userId, textId uuid.UUID,
	wordsTyped int,
	timeElapsed float64,
	errorsJSON models.ErrorsJSON,
) (*models.Score, error) {
	const op errors.Op = "services.ScoreService.Create"

	numberErrors := 0
	for _, value := range errorsJSON {
		numberErrors += value
	}

	var newScore = models.Score{
		WordsTyped:   wordsTyped,
		TimeElapsed:  timeElapsed,
		Errors:       errorsJSON,
		UserId:       userId,
		GameId:       gameId,
		NumberErrors: numberErrors,
		TextId:       textId,
	}

	createdScore, err := ss.dbRepo.CreateScore(ctx, newScore)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return createdScore, nil
}

func (ss *ScoreService) FindScores(
	ctx context.Context,
	userId, gameId uuid.UUID,
	username string,
	sortOptions []models.SortOption,
) ([]models.Score, error) {
	const op errors.Op = "services.ScoreService.FindScores"

	scores, err := ss.dbRepo.FindScores(ctx, userId, gameId, username, sortOptions)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return scores, nil
}

package services

import (
	"10-typing/errors"
	"10-typing/models"
	"10-typing/repositories"

	"github.com/google/uuid"
)

type ScoreService struct {
	dbRepo repositories.DBRepository
}

func NewScoreService(dbRepo repositories.DBRepository) *ScoreService {
	return &ScoreService{dbRepo}
}

func (ss *ScoreService) Create(
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

	createdScore, err := ss.dbRepo.CreateScore(newScore)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return createdScore, nil
}

func (ss *ScoreService) FindScores(
	userId, gameId uuid.UUID,
	username string,
	sortOptions []models.SortOption,
) ([]models.Score, error) {
	const op errors.Op = "services.ScoreService.FindScores"

	scores, err := ss.dbRepo.FindScores(userId, gameId, username, sortOptions)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return scores, nil
}

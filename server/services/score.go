package services

import (
	"10-typing/models"
	"10-typing/repositories"

	"github.com/google/uuid"
)

type ScoreService struct {
	scoreRepo *repositories.ScoreDbRepository
}

func NewScoreService(scoreRepo *repositories.ScoreDbRepository) *ScoreService {
	return &ScoreService{scoreRepo}
}

func (ss *ScoreService) Create(
	gameId, userId, textId uuid.UUID,
	wordsTyped int,
	timeElapsed float64,
	errorsJSON models.ErrorsJSON,
) (*models.Score, error) {
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

	return ss.scoreRepo.Create(newScore)
}

func (ss *ScoreService) FindScores(
	userId, gameId uuid.UUID,
	username string,
	sortOptions []models.SortOption,
) ([]models.Score, error) {
	return ss.scoreRepo.FindScores(userId, gameId, username, sortOptions)
}

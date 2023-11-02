package repositories

import (
	"10-typing/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScoreDbRepository struct {
	db *gorm.DB
}

func NewScoreDbRepository(db *gorm.DB) *ScoreDbRepository {
	return &ScoreDbRepository{db}
}

func (sr *ScoreDbRepository) FindScores(userId, gameId uuid.UUID, username string, sortOptions []models.SortOption) ([]models.Score, error) {
	var scores []models.Score

	findScoresDbQuery := sr.db
	if userId != uuid.Nil {
		findScoresDbQuery = findScoresDbQuery.Where("user_id = ?", userId)
	}
	if gameId != uuid.Nil {
		findScoresDbQuery = findScoresDbQuery.Where("game_id = ?", gameId)
	}
	if username != "" {
		findScoresDbQuery = findScoresDbQuery.Joins("INNER JOIN users ON scores.user_id = users.id").Where("users.username = ?", username)
	}

	for _, sortOption := range sortOptions {
		findScoresDbQuery = findScoresDbQuery.Order(clause.OrderByColumn{Column: clause.Column{Name: sortOption.Column}, Desc: sortOption.Order == "desc"})
	}
	if len(sortOptions) == 0 {
		findScoresDbQuery = findScoresDbQuery.Order("created_at desc")
	}

	findScoresDbQuery.Find(&scores)

	if findScoresDbQuery.Error != nil {
		return nil, findScoresDbQuery.Error
	}

	return scores, nil
}

func (sr *ScoreDbRepository) Create(score models.Score) (*models.Score, error) {
	omittedFiels := []string{"WordsPerMinute", "Accuracy"}

	if score.GameId == uuid.Nil {
		omittedFiels = append(omittedFiels, "GameId")
	}

	createResult := sr.db.
		Omit(omittedFiels...).
		Clauses(clause.Returning{
			Columns: []clause.Column{
				{Name: "id"},
				{Name: "words_per_minute"},
				{Name: "words_typed"},
				{Name: "time_elapsed"},
				{Name: "accuracy"},
				{Name: "number_errors"},
				{Name: "errors"},
			}}).
		Create(&score)

	if createResult.Error != nil {
		return nil, createResult.Error
	}
	if createResult.RowsAffected == 0 {
		return nil, errors.New("no rows affected")
	}

	return &score, nil
}

func (sr *ScoreDbRepository) DeleteAll() error {
	return sr.db.Exec("TRUNCATE scores RESTART IDENTITY CASCADE").Error
}

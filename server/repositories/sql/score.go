package sql_repo

import (
	"10-typing/errors"
	"10-typing/models"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

func (repo *SQLRepository) FindScores(userId, gameId uuid.UUID, username string, sortOptions []models.SortOption) ([]models.Score, error) {
	const op errors.Op = "sql_repo.SQLRepository.FindScores"
	var scores []models.Score

	findScoresDbQuery := repo.db
	if userId != uuid.Nil {
		findScoresDbQuery = findScoresDbQuery.Where("user_id = ?", userId)
	}
	if gameId != uuid.Nil {
		findScoresDbQuery = findScoresDbQuery.Where("game_id = ?", gameId)
	}
	if username != "" {
		findScoresDbQuery = findScoresDbQuery.Joins("INNER JOIN users ON scores.user_id = users.id").
			Where("users.username = ?", username)
	}

	for _, sortOption := range sortOptions {
		findScoresDbQuery = findScoresDbQuery.Order(clause.OrderByColumn{Column: clause.Column{Name: sortOption.Column}, Desc: sortOption.Order == "desc"})
	}
	if len(sortOptions) == 0 {
		findScoresDbQuery = findScoresDbQuery.Order("created_at desc")
	}

	findScoresDbQuery.Find(&scores)

	if findScoresDbQuery.Error != nil {
		return nil, errors.E(op, findScoresDbQuery.Error)
	}

	return scores, nil
}

func (repo *SQLRepository) CreateScore(score models.Score) (*models.Score, error) {
	const op errors.Op = "sql_repo.SQLRepository.CreateScore"
	omittedFiels := []string{"WordsPerMinute", "Accuracy"}

	if score.GameId == uuid.Nil {
		omittedFiels = append(omittedFiels, "GameId")
	}

	if err := repo.db.
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
				{Name: "user_id"},
			}}).
		Create(&score).Error; err != nil {
		return nil, errors.E(op, err)
	}

	return &score, nil
}

func (repo *SQLRepository) DeleteAllScores() error {
	const op errors.Op = "sql_repo.SQLRepository.DeleteAllScores"

	if err := repo.db.Exec("TRUNCATE scores RESTART IDENTITY CASCADE").Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

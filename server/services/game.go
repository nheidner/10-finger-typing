package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	gameDurationSeconds           = 30
	waitForResultsDurationSeconds = 5
	countdownDurationSeconds      = 3
)

type GameService struct {
	gameRedisRepo       *repositories.GameRedisRepository
	roomStreamRedisRepo *repositories.RoomStreamRedisRepository
	scoreDbRepo         *repositories.ScoreDbRepository
	textRedisRepo       *repositories.TextRedisRepository
}

func NewGameService(
	gameRedisRepo *repositories.GameRedisRepository,
	roomStreamRedisRepo *repositories.RoomStreamRedisRepository,
	scoreDbRepo *repositories.ScoreDbRepository,
	textRedisRepo *repositories.TextRedisRepository,
) *GameService {
	return &GameService{gameRedisRepo, roomStreamRedisRepo, scoreDbRepo, textRedisRepo}
}

func (gs *GameService) SetNewCurrentGame(userId, roomId, textId uuid.UUID) (uuid.UUID, error) {
	var ctx = context.Background()

	// validate
	textExists, err := gs.textRedisRepo.TextExists(ctx, textId)
	if err != nil {
		return uuid.Nil, err
	}
	if !textExists {
		return uuid.Nil, errors.New("text does not exist")
	}

	currentGameStatus, err := gs.gameRedisRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return uuid.Nil, err
	}
	roomHasActiveGame := currentGameStatus == models.StartedGameStatus
	if roomHasActiveGame {
		return uuid.Nil, errors.New("room has active game at the moment")
	}

	var gameId = uuid.New()

	if err := gs.gameRedisRepo.SetNewCurrentGameInRedis(ctx, gameId, textId, roomId, userId); err != nil {
		return uuid.Nil, err
	}

	return gameId, nil
}

func (gs *GameService) UserFinishesGame(
	roomId,
	userId,
	textId uuid.UUID,
	wordsTyped int,
	timeElapsed float64,
	errorsJSON models.ErrorsJSON,
) error {
	var ctx = context.Background()

	// check if game status is not already finished
	currentGameStatus, err := gs.gameRedisRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		log.Println("error setting current game status: ", err)
		return err
	case currentGameStatus != models.StartedGameStatus:
		log.Println("game has not the correct status: ", err)
		return errors.New("game has not the correct status")
	}

	// get game id
	gameId, err := gs.gameRedisRepo.GetCurrentGameId(ctx, roomId)
	if err != nil {
		log.Println("error when getting current game id from redis: ", err)
		return err
	}

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

	_, err = gs.scoreDbRepo.Create(newScore)
	if err != nil {
		log.Println("error when creating a new score: ", err)
		return err
	}

	// post action on stream
	err = gs.roomStreamRedisRepo.PublishAction(ctx, roomId, models.GameUserScoreAction)
	if err != nil {
		log.Println("error when publishing the game user score action: ", err)
		return err
	}

	return nil
}

func (gs *GameService) AddUserToGame(roomId, userId uuid.UUID) error {
	var ctx = context.Background()

	currentGameStatus, err := gs.gameRedisRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return err
	}
	if currentGameStatus > models.CountdownGameStatus {
		// log.Println("game is not in unstarted state", err)
		return err
	}

	err = gs.gameRedisRepo.AddGameUserInRedis(ctx, roomId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (gs *GameService) InitiateGameIfReady(roomId uuid.UUID) error {
	var ctx = context.Background()

	numberGameUsers, err := gs.gameRedisRepo.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return err
	}

	if numberGameUsers == 2 {
		// start countdown
		err := gs.gameRedisRepo.SetCurrentGameStatusInRedis(ctx, roomId, models.CountdownGameStatus)
		if err != nil {
			return err
		}

		countdownPushMessage := repositories.PushMessage{
			Type: models.CountdownStart,
			Payload: map[string]any{
				"duration": countdownDurationSeconds,
			},
		}

		err = gs.roomStreamRedisRepo.PublishPushMessage(ctx, roomId, countdownPushMessage)
		if err != nil {
			return err
		}

		go func() {
			time.Sleep(countdownDurationSeconds*time.Second + gameDurationSeconds*time.Second)
			if err = gs.handleGameResults(roomId); err != nil {
				log.Println(err)
			}
		}()
	}

	return nil
}

func (gs *GameService) handleGameResults(roomId uuid.UUID) error {
	var ctx = context.Background()

	numberGameUsers, err := gs.gameRedisRepo.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return errors.New("error getting current game users:" + err.Error())
	}
	ctx, cancel := context.WithCancel(ctx)
	allResultsReceivedCh := gs.getAllResultsReceived(ctx, numberGameUsers, roomId)

	timer := time.NewTimer(waitForResultsDurationSeconds * time.Second)

	select {
	case <-timer.C:
		cancel()
	case <-allResultsReceivedCh:
		timer.Stop()
		cancel()
	}

	// set game status to Finished
	err = gs.gameRedisRepo.SetCurrentGameStatusInRedis(ctx, roomId, models.FinishedGameStatus)
	if err != nil {
		return errors.New("error setting game status to finished:" + err.Error())
	}

	gameId, err := gs.gameRedisRepo.GetCurrentGameId(ctx, roomId)
	if err != nil {
		return errors.New("error when getting current game id from redis: " + err.Error())
	}

	scores, err := gs.scoreDbRepo.FindScores(
		uuid.Nil, gameId,
		"",
		[]models.SortOption{{Column: "words_per_minute", Order: "desc"}},
	)

	if err != nil {
		return errors.New("error findind scores:" + err.Error())
	}
	scorePushMessage := repositories.PushMessage{
		Type:    models.GameScores,
		Payload: scores,
	}

	return gs.roomStreamRedisRepo.PublishPushMessage(ctx, roomId, scorePushMessage)
}

func (gs *GameService) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	allReceived := make(chan struct{})

	go func() {
		defer close(allReceived)
		actionCh, errCh := gs.roomStreamRedisRepo.GetAction(ctx, roomId, time.Time{})

		for resultsCount := 0; resultsCount < playersNumber; {
			select {
			case action, ok := <-actionCh:
				if !ok {
					return
				}

				if action == models.GameUserScoreAction {
					resultsCount++
					continue
				}

			case err, ok := <-errCh:
				if !ok {
					return
				}

				log.Println(err)
				return
			case <-ctx.Done():
				return
			}

		}

		allReceived <- struct{}{}
	}()

	return allReceived
}

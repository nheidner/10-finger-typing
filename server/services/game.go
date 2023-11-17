package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	waitForResultsDurationSeconds = 10
	countdownDurationSeconds      = 3
)

type GameService struct {
	dbRepo    repositories.DBRepository
	cacheRepo repositories.CacheRepository
}

func NewGameService(
	dbRepo repositories.DBRepository,
	cacheRepo repositories.CacheRepository,
) *GameService {
	return &GameService{dbRepo, cacheRepo}
}

func (gs *GameService) SetNewCurrentGame(userId, roomId, textId uuid.UUID) (uuid.UUID, error) {
	var ctx = context.Background()

	// validate
	textExists, err := gs.cacheRepo.TextIdExists(ctx, textId)
	if err != nil {
		return uuid.Nil, err
	}
	if !textExists {
		return uuid.Nil, errors.New("text does not exist")
	}

	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return uuid.Nil, err
	}
	roomHasActiveGame := currentGameStatus == models.StartedGameStatus
	if roomHasActiveGame {
		return uuid.Nil, errors.New("room has active game at the moment")
	}

	var gameId = uuid.New()

	if err := gs.cacheRepo.SetNewCurrentGame(ctx, gameId, textId, roomId, userId); err != nil {
		return uuid.Nil, err
	}

	if err := gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.UnstartedGameStatus); err != nil {
		return uuid.Nil, err
	}

	// TODO: send new_game message

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
	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		log.Println("error setting current game status: ", err)
		return err
	case currentGameStatus != models.StartedGameStatus:
		log.Println("game has not the correct status: ", err)
		return errors.New("game has not the correct status")
	}

	// get game id
	gameId, err := gs.cacheRepo.GetCurrentGameId(ctx, roomId)
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

	_, err = gs.dbRepo.CreateScore(newScore)
	if err != nil {
		log.Println("error when creating a new score: ", err)
		return err
	}

	// post action on stream
	err = gs.cacheRepo.PublishAction(ctx, roomId, models.GameUserScoreAction)
	if err != nil {
		log.Println("error when publishing the game user score action: ", err)
		return err
	}

	return nil
}

func (gs *GameService) AddUserToGame(roomId, userId uuid.UUID) error {
	var ctx = context.Background()

	// comment out
	isCurrentGameUser, err := gs.cacheRepo.IsCurrentGameUser(ctx, roomId, userId)
	if err != nil {
		return err
	}
	if isCurrentGameUser {
		return errors.New("game was already started")
	}

	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return err
	}
	if currentGameStatus > models.CountdownGameStatus {
		// log.Println("game is not in unstarted state", err)
		return err
	}

	err = gs.cacheRepo.SetGameUser(ctx, roomId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (gs *GameService) InitiateGameIfReady(roomId uuid.UUID) error {
	var ctx = context.Background()

	numberGameUsers, err := gs.cacheRepo.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return err
	}

	if numberGameUsers != 2 {
		return nil
	}

	// comment out
	gameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return err
	}

	if gameStatus != models.UnstartedGameStatus {
		return nil
	}

	// start countdown
	err = gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.CountdownGameStatus)
	if err != nil {
		return err
	}

	countdownPushMessage := models.PushMessage{
		Type:    models.CountdownStart,
		Payload: countdownDurationSeconds,
	}

	err = gs.cacheRepo.PublishPushMessage(ctx, roomId, countdownPushMessage)
	if err != nil {
		return err
	}

	gameDurationSec, err := gs.cacheRepo.GetRoomGameDurationSec(ctx, roomId)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(countdownDurationSeconds * time.Second)

		// after blocking for countdown duration, set game status to "started"
		err = gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.StartedGameStatus)
		if err != nil {
			log.Println("error setting game status to finished:", err.Error())
			return
		}

		time.Sleep(time.Duration(gameDurationSec) * time.Second)

		if err = gs.handleGameResults(roomId); err != nil {
			log.Println("error handling game results", err)
		}
	}()

	return nil
}

func (gs *GameService) handleGameResults(roomId uuid.UUID) error {
	var ctx = context.Background()

	numberGameUsers, err := gs.cacheRepo.GetCurrentGameUsersNumber(ctx, roomId)
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

	ctx = context.Background()

	// set game status to finished
	err = gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.FinishedGameStatus)
	if err != nil {
		return errors.New("error setting game status to finished:" + err.Error())
	}

	gameId, err := gs.cacheRepo.GetCurrentGameId(ctx, roomId)
	if err != nil {
		return errors.New("error when getting current game id from redis: " + err.Error())
	}

	scores, err := gs.dbRepo.FindScores(uuid.Nil, gameId, "", []models.SortOption{{Column: "words_per_minute", Order: "desc"}})

	fmt.Println("scores :>>", scores)

	if err != nil {
		return errors.New("error findind scores:" + err.Error())
	}
	scorePushMessage := models.PushMessage{
		Type:    models.GameScores,
		Payload: scores,
	}

	return gs.cacheRepo.PublishPushMessage(ctx, roomId, scorePushMessage)
}

func (gs *GameService) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	allReceived := make(chan struct{})

	go func() {
		defer close(allReceived)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		actionResultCh := gs.cacheRepo.GetAction(ctx, roomId, time.Time{})

		for resultsCount := 0; resultsCount < playersNumber; {
			select {
			case actionResult, ok := <-actionResultCh:
				if !ok {
					return
				}
				if actionResult.Error != nil {
					log.Println("error from action result channel: ", actionResult.Error)
					return
				}
				if actionResult.Value == models.GameUserScoreAction {
					resultsCount++
					continue
				}
			case <-ctx.Done():
				return
			}
		}

		allReceived <- struct{}{}
	}()

	return allReceived
}

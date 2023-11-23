package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"context"
	"fmt"

	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	waitForResultsDurationSeconds = 10
	countdownDurationSeconds      = 3
)

type GameService struct {
	dbRepo    common.DBRepository
	cacheRepo common.CacheRepository
	logger    common.Logger
}

func NewGameService(
	dbRepo common.DBRepository,
	cacheRepo common.CacheRepository,
	logger common.Logger,
) *GameService {
	return &GameService{dbRepo, cacheRepo, logger}
}

func (gs *GameService) SetNewCurrentGame(ctx context.Context, userId, roomId, textId uuid.UUID) (uuid.UUID, error) {
	const op errors.Op = "services.GameService.SetNewCurrentGame"

	// validate
	textExists, err := gs.cacheRepo.TextIdExists(ctx, textId)
	switch {
	case err != nil:
		return uuid.Nil, errors.E(op, err)
	case !textExists:
		err := fmt.Errorf("text does not exist")
		return uuid.Nil, errors.E(op, err, http.StatusBadRequest)
	}

	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case errors.Is(err, common.ErrNotFound):
		break
	case err != nil:
		return uuid.Nil, errors.E(op, err)
	case currentGameStatus == models.StartedGameStatus:
		err := fmt.Errorf("room has active game")
		return uuid.Nil, errors.E(op, err, http.StatusBadRequest)
	}

	var gameId = uuid.New()

	if err := gs.cacheRepo.SetNewCurrentGame(ctx, gameId, textId, roomId, userId); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	if err := gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.UnstartedGameStatus); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	// TODO: send new_game message

	return gameId, nil
}

func (gs *GameService) UserFinishesGame(
	ctx context.Context,
	roomId,
	userId,
	textId uuid.UUID,
	wordsTyped int,
	timeElapsed float64,
	errorsJSON models.ErrorsJSON,
) error {
	const op errors.Op = "services.GameService.UserFinishesGame"

	// check if game status is not already finished
	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case currentGameStatus != models.StartedGameStatus:
		err := fmt.Errorf("game has not the correct status")
		return errors.E(op, err, http.StatusBadRequest)
	}

	// get game id
	gameId, err := gs.cacheRepo.GetCurrentGameId(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
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
		return errors.E(op, err)
	}

	// post action on stream
	err = gs.cacheRepo.PublishAction(ctx, roomId, models.GameUserScoreAction)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (gs *GameService) AddUserToGame(ctx context.Context, roomId, userId uuid.UUID) error {
	const op errors.Op = "services.GameService.AddUserToGame"

	// comment out
	isCurrentGameUser, err := gs.cacheRepo.IsCurrentGameUser(ctx, roomId, userId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case isCurrentGameUser:
		err := fmt.Errorf("user already joined game")
		return errors.E(op, err, http.StatusBadRequest)
	}

	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case currentGameStatus > models.CountdownGameStatus:
		err := fmt.Errorf("game was already started")
		return errors.E(op, err, http.StatusBadRequest)
	}

	err = gs.cacheRepo.SetGameUser(ctx, roomId, userId)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (gs *GameService) InitiateGameIfReady(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "services.GameService.InitiateGameIfReady"

	numberGameUsers, err := gs.cacheRepo.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	if numberGameUsers != 2 {
		return nil
	}

	// comment out
	gameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	if gameStatus != models.UnstartedGameStatus {
		return nil
	}

	// start countdown
	err = gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.CountdownGameStatus)
	if err != nil {
		return errors.E(op, err)
	}

	countdownPushMessage := models.PushMessage{
		Type:    models.CountdownStart,
		Payload: countdownDurationSeconds,
	}

	err = gs.cacheRepo.PublishPushMessage(ctx, roomId, countdownPushMessage)
	if err != nil {
		return errors.E(op, err)
	}

	gameDurationSec, err := gs.cacheRepo.GetRoomGameDurationSec(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	go gs.handleGameDuration(context.Background(), gameDurationSec, roomId)

	return nil
}

func (gs *GameService) handleGameDuration(ctx context.Context, gameDurationSec int, roomId uuid.UUID) {
	const op errors.Op = "services.GameService.handleGameDuration"

	time.Sleep(countdownDurationSeconds * time.Second)

	// after blocking for countdown duration, set game status to "started"
	err := gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.StartedGameStatus)
	if err != nil {
		gs.logger.Error(errors.E(op, err))
		return
	}

	time.Sleep(time.Duration(gameDurationSec) * time.Second)

	if err = gs.handleGameResults(ctx, roomId); err != nil {
		gs.logger.Error(errors.E(op, err))
	}
}

func (gs *GameService) handleGameResults(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "services.GameService.handleGameResults"

	numberGameUsers, err := gs.cacheRepo.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	allResultsReceivedCh := gs.getAllResultsReceived(cancelCtx, numberGameUsers, roomId)

	timer := time.NewTimer(waitForResultsDurationSeconds * time.Second)

	select {
	case <-timer.C:
	case <-allResultsReceivedCh:
		timer.Stop()
	}
	cancel()

	// set game status to finished
	err = gs.cacheRepo.SetCurrentGameStatus(ctx, roomId, models.FinishedGameStatus)
	if err != nil {
		return errors.E(op, err)
	}

	gameId, err := gs.cacheRepo.GetCurrentGameId(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	scores, err := gs.dbRepo.FindScores(uuid.Nil, gameId, "", []models.SortOption{{Column: "words_per_minute", Order: "desc"}})
	if err != nil {
		return errors.E(op, err)
	}

	scorePushMessage := models.PushMessage{
		Type:    models.GameScores,
		Payload: scores,
	}

	if err := gs.cacheRepo.PublishPushMessage(ctx, roomId, scorePushMessage); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (gs *GameService) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	const op errors.Op = "services.GameService.getAllResultsReceived"
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
					gs.logger.Error(errors.E(op, actionResult.Error))
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

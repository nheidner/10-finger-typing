package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"
	"10-typing/utils"
	"context"
	"fmt"

	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	waitForResultsDurationSeconds = 10
	countdownDurationSeconds      = 5
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

func (gs *GameService) CreateNewCurrentGame(ctx context.Context, userId, roomId, textId uuid.UUID) (uuid.UUID, error) {
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

	// cleanup
	if err := gs.cacheRepo.DeleteCurrentGameScores(ctx, roomId); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	// create game
	if err := gs.cacheRepo.SetRoomSubscriberGameStatusForAllRoomSubscribers(ctx, roomId, models.UnstartedSubscriberGameStatus); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	if err := gs.cacheRepo.SetNewCurrentGame(ctx, nil, gameId, textId, roomId, userId); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	newGame, err := gs.cacheRepo.GetCurrentGame(ctx, roomId)
	if err != nil {
		return uuid.Nil, errors.E(op, err)
	}

	if err := gs.cacheRepo.PublishPushMessage(ctx, nil, roomId, models.PushMessage{
		Type:    models.NewGame,
		Payload: newGame,
	}); err != nil {
		return uuid.Nil, errors.E(op, err)
	}

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

	// validate: check if game status is not already finished
	currentGameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case currentGameStatus != models.StartedGameStatus:
		err := fmt.Errorf("game has not the correct status")
		return errors.E(op, err, http.StatusBadRequest)
	}

	if err := gs.cacheRepo.SetRoomSubscriberGameStatus(ctx, nil, roomId, userId, models.FinishedSubscriberGameStatus); err != nil {
		return errors.E(op, err)
	}

	userFinishedGamePushMessage := models.PushMessage{
		Type:    models.UserFinishedGame,
		Payload: userId,
	}
	if err := gs.cacheRepo.PublishPushMessage(ctx, nil, roomId, userFinishedGamePushMessage); err != nil {
		return errors.E(op, err)
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

	createdScore, err := gs.dbRepo.CreateScore(ctx, nil, newScore)
	if err != nil {
		return errors.E(op, err)
	}

	// post action on stream
	err = gs.cacheRepo.PublishAction(ctx, nil, roomId, models.GameUserScoreAction)
	if err != nil {
		return errors.E(op, err)
	}

	// the error should only be logged but not returned because the score is already saved in the DB
	if err := gs.cacheRepo.SetCurrentGameScore(ctx, nil, roomId, *createdScore); err != nil {
		gs.logger.Error(errors.E(op, err))
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

	// PIPELINE start
	tx := gs.cacheRepo.BeginPipeline()
	if err := gs.cacheRepo.SetCurrentGameUser(ctx, tx, roomId, userId); err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	if err := gs.cacheRepo.SetRoomSubscriberGameStatus(ctx, tx, roomId, userId, models.StartedSubscriberGameStatus); err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	userStartedGamePushMessage := models.PushMessage{
		Type:    models.UserStartedGame,
		Payload: userId,
	}
	if err := gs.cacheRepo.PublishPushMessage(ctx, tx, roomId, userStartedGamePushMessage); err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	// PIPELINE commit
	if err := tx.Commit(ctx); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (gs *GameService) InitiateGameIfReady(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "services.GameService.InitiateGameIfReady"

	numberGameUsers, err := gs.cacheRepo.GetCurrentGameUsersNumber(ctx, roomId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case numberGameUsers != 2:
		return nil
	}

	gameStatus, err := gs.cacheRepo.GetCurrentGameStatus(ctx, roomId)
	switch {
	case err != nil:
		return errors.E(op, err)
	case gameStatus != models.UnstartedGameStatus:
		return nil
	}

	// start countdown
	err = gs.cacheRepo.SetCurrentGameStatus(ctx, nil, roomId, models.CountdownGameStatus)
	if err != nil {
		return errors.E(op, err)
	}

	gameDurationSec, err := gs.cacheRepo.GetRoomGameDurationSec(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	ctx = context.Background()

	go gs.countdown(ctx, roomId, countdownDurationSeconds)
	go func() {
		defer gs.cleanupGame(ctx, roomId)

		if err = gs.handleGameDuration(ctx, gameDurationSec, roomId); err != nil {
			gs.logger.Error(errors.E(op, err))
			return
		}

		if err = gs.handleGameResults(ctx, roomId); err != nil {
			gs.logger.Error(errors.E(op, err))
		}
	}()

	return nil
}

func (gs *GameService) countdown(ctx context.Context, roomId uuid.UUID, countdownDurationSeconds int) {
	const op errors.Op = "services.GameService.countdown"

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for countdownDurationSeconds > 0 {
		countdownPushMessage := models.PushMessage{
			Type:    models.Countdown,
			Payload: countdownDurationSeconds,
		}
		err := gs.cacheRepo.PublishPushMessage(ctx, nil, roomId, countdownPushMessage)
		if err != nil {
			gs.logger.Error(errors.E(op, err))
			return
		}

		countdownDurationSeconds--
		<-ticker.C
	}
}

func (gs *GameService) handleGameDuration(ctx context.Context, gameDurationSec int, roomId uuid.UUID) error {
	const op errors.Op = "services.GameService.handleGameDuration"

	time.Sleep(countdownDurationSeconds * time.Second)

	// after blocking for countdown duration, set game status to "started"
	if err := gs.cacheRepo.SetCurrentGameStatus(ctx, nil, roomId, models.StartedGameStatus); err != nil {
		return errors.E(op, err)
	}

	gameStartedPushMessage := models.PushMessage{
		Type: models.GameStarted,
	}
	if err := gs.cacheRepo.PublishPushMessage(ctx, nil, roomId, gameStartedPushMessage); err != nil {
		return errors.E(op, err)
	}

	time.Sleep(time.Duration(gameDurationSec) * time.Second)

	return nil
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

	currentGameScores, err := gs.cacheRepo.GetCurrentGameScores(ctx, roomId)
	if err != nil {
		return errors.E(op, err)
	}

	scorePushMessage := models.PushMessage{
		Type:    models.GameScores,
		Payload: currentGameScores,
	}
	if err := gs.cacheRepo.PublishPushMessage(ctx, nil, roomId, scorePushMessage); err != nil {
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

func (gs *GameService) cleanupGame(ctx context.Context, roomId uuid.UUID) error {
	const op errors.Op = "services.GameService.cleanupGame"

	// PIPELINE start
	tx := gs.cacheRepo.BeginPipeline()

	// set game status to finished
	err := gs.cacheRepo.SetCurrentGameStatus(ctx, tx, roomId, models.FinishedGameStatus)
	if err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	if err := gs.cacheRepo.DeleteAllCurrentGameUsers(ctx, tx, roomId); err != nil {
		err := errors.E(op, err)
		return utils.RollbackAndErr(op, err, tx)
	}

	// PIPELINE commit
	if err := tx.Commit(ctx); err != nil {
		return errors.E(op, err)
	}

	return nil
}

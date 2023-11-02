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
	gameRedisRepo  *repositories.GameRedisRepository
	roomStreamRepo *repositories.RoomStreamRepository
	scoreDbRepo    *repositories.ScoreDbRepository
}

func NewGameService(
	gameRedisRepo *repositories.GameRedisRepository,
	roomStreamRepo *repositories.RoomStreamRepository,
	scoreDbRepo *repositories.ScoreDbRepository,
) *GameService {
	return &GameService{gameRedisRepo, roomStreamRepo, scoreDbRepo}
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
			Type: repositories.CountdownStart,
			Payload: map[string]any{
				"duration": countdownDurationSeconds,
			},
		}

		err = gs.roomStreamRepo.PublishPushMessage(ctx, roomId, countdownPushMessage)
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
		Type:    repositories.GameScores,
		Payload: scores,
	}

	return gs.roomStreamRepo.PublishPushMessage(ctx, roomId, scorePushMessage)
}

func (gs *GameService) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	allReceived := make(chan struct{})

	go func() {
		defer close(allReceived)
		actionCh, errCh := gs.roomStreamRepo.GetAction(ctx, roomId, time.Time{})

		for resultsCount := 0; resultsCount < playersNumber; {
			select {
			case action, ok := <-actionCh:
				if !ok {
					return
				}

				if action == repositories.GameUserScoreAction {
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

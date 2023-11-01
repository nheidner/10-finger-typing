package controllers

import (
	"10-typing/models"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	gameDurationSeconds           = 30
	waitForResultsDurationSeconds = 5
	countdownDurationSeconds      = 3
)

type Games struct {
	GameService           *models.GameService
	RoomService           *models.RoomService
	TextService           *models.TextService
	RoomSubscriberService *models.RoomSubscriberService
	ScoreService          *models.ScoreService
	RoomStreamService     *models.RoomStreamService
}

func (g *Games) CreateGame(c *gin.Context) {
	user, textId, roomId, err := processCreateGameHTTPParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var game = models.Game{
		ID: uuid.New(),
	}
	var ctx = context.Background()

	// validate
	textExists, err := g.TextService.TextExists(ctx, textId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !textExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text does not exist"})
		return
	}

	currentGameStatus, err := g.GameService.GetCurrentGameStatus(ctx, roomId)
	if err != nil {
		log.Println("current game status error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	roomHasActiveGame := currentGameStatus == models.StartedGameStatus
	if roomHasActiveGame {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room has active game at the moment"})
		return
	}

	if err := g.GameService.SetNewCurrentGame(ctx, game.ID, textId, roomId, user.ID); err != nil {
		log.Println("SetNewCurrentGame", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{
		"id": game.ID.String(),
	}})
}

func processCreateGameHTTPParams(c *gin.Context) (user *models.User, textId, roomId uuid.UUID, err error) {
	var input models.CreateGameInput

	user, err = getUserFromContext(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	roomId, err = getRoomIdFromPath(c)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	return user, input.TextId, roomId, nil
}

func (g *Games) StartGame(c *gin.Context) {
	user, roomId, err := processStartGameHTTPParams(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	currentGameStatus, err := g.GameService.GetCurrentGameStatus(c.Request.Context(), roomId)
	if err != nil {
		log.Println("RoomHasActiveGame error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if currentGameStatus > models.CountdownGameStatus {
		log.Println("game is not in unstarted state", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = g.GameService.AddGameUser(c.Request.Context(), roomId, user.ID)
	if err != nil {
		log.Println("error adding new game user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	numberGameUsers, err := g.GameService.GetCurrentGameUsersNumber(c.Request.Context(), roomId)
	if err != nil {
		log.Println("error adding new game user: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	if numberGameUsers == 2 {
		// start countdown
		err := g.GameService.SetCurrentGameStatus(c.Request.Context(), roomId, models.CountdownGameStatus)
		if err != nil {
			log.Println("error setting current game status: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
			return
		}

		countdownPushMessage := models.PushMessage{
			Type: models.CountdownStart,
			Payload: map[string]any{
				"duration": countdownDurationSeconds,
			},
		}

		err = g.RoomStreamService.PublishPushMessage(c.Request.Context(), roomId, countdownPushMessage)
		if err != nil {
			log.Println("error publishing to stream:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
			return
		}

		go func() {
			time.Sleep(countdownDurationSeconds*time.Second + gameDurationSeconds*time.Second)
			if err = g.handleGameResults(context.Background(), roomId); err != nil {
				log.Panicln(err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"data": "game started"})
}

func (g *Games) FinishGame(c *gin.Context) {
	user, roomId, scoreInput, err := processFinishGameHTTPParams(c)
	if err != nil {
		log.Println("error processing http params: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	// check if game status is not already finished
	currentGameStatus, err := g.GameService.GetCurrentGameStatus(c.Request.Context(), roomId)
	switch {
	case err != nil:
		log.Println("error setting current game status: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	case currentGameStatus != models.StartedGameStatus:
		log.Println("game has not the correct status: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error"})
		return
	}

	// get game id
	gameId, err := g.GameService.GetCurrentGameId(c.Request.Context(), roomId)
	if err != nil {
		log.Println("error when getting current game id from redis: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	scoreInput.GameId = gameId
	scoreInput.UserId = user.ID

	_, err = g.ScoreService.Create(*scoreInput)
	if err != nil {
		log.Println("error when creating a new score: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	// post action on stream
	err = g.RoomStreamService.PublishAction(c.Request.Context(), roomId, models.GameUserScoreAction)
	if err != nil {
		log.Println("error when publishing the game user score action: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "score added"})
}

func processStartGameHTTPParams(c *gin.Context) (user *models.User, roomId uuid.UUID, err error) {
	user, err = getUserFromContext(c)
	if err != nil {
		return nil, uuid.Nil, err
	}

	roomId, err = getRoomIdFromPath(c)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return user, roomId, err
}

func processFinishGameHTTPParams(c *gin.Context) (*models.User, uuid.UUID, *models.CreateScoreInput, error) {
	var input models.CreateScoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, uuid.Nil, nil, err
	}

	user, err := getUserFromContext(c)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	roomId, err := getRoomIdFromPath(c)
	if err != nil {
		return nil, uuid.Nil, nil, err
	}

	return user, roomId, &input, err
}

// create extra function
func (g *Games) handleGameResults(ctx context.Context, roomId uuid.UUID) error {
	numberGameUsers, err := g.GameService.GetCurrentGameUsersNumber(ctx, roomId)
	if err != nil {
		return errors.New("error getting current game users:" + err.Error())
	}
	ctx, cancel := context.WithCancel(ctx)
	allResultsReceivedCh := g.getAllResultsReceived(ctx, numberGameUsers, roomId)

	timer := time.NewTimer(waitForResultsDurationSeconds * time.Second)

	select {
	case <-timer.C:
		cancel()
	case <-allResultsReceivedCh:
		timer.Stop()
		cancel()
	}

	// set game status to Finished
	err = g.GameService.SetCurrentGameStatus(ctx, roomId, models.FinishedGameStatus)
	if err != nil {
		return errors.New("error setting game status to finished:" + err.Error())
	}

	gameId, err := g.GameService.GetCurrentGameId(ctx, roomId)
	if err != nil {
		return errors.New("error when getting current game id from redis: " + err.Error())
	}

	scores, err := g.ScoreService.FindScores(&models.FindScoresQuery{
		SortOptions: []models.SortOption{{Column: "words_per_minute", Order: "desc"}},
		GameId:      gameId,
	})
	if err != nil {
		return errors.New("error findind scores:" + err.Error())
	}
	scorePushMessage := models.PushMessage{
		Type:    models.GameScores,
		Payload: scores,
	}

	return g.RoomStreamService.PublishPushMessage(ctx, roomId, scorePushMessage)
}

func (g *Games) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	allReceived := make(chan struct{})

	go func() {
		defer close(allReceived)
		actionCh, errCh := g.RoomStreamService.GetAction(ctx, roomId, time.Time{})

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

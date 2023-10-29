package controllers

import (
	"10-typing/models"
	"context"
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

	// if gameUsers.length == 2, then start countdown
	if numberGameUsers == 2 {
		err := g.GameService.SetCurrentGameStatus(c.Request.Context(), roomId, models.CountdownGameStatus)
		if err != nil {
			log.Println("error setting current game status: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
			return
		}

		countdownMessage := models.WSMessage{Type: "countdown_start", Payload: map[string]any{
			"duration": countdownDurationSeconds,
		}}

		err = g.RoomService.Publish(c.Request.Context(), roomId, countdownMessage)
		if err != nil {
			log.Println("error publishing to stream:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error"})
			return
		}

		// create extra function
		go func() {
			time.Sleep(time.Second*countdownDurationSeconds + gameDurationSeconds*time.Second - 1*time.Second)

			numberGameUsers, err := g.GameService.GetCurrentGameUsersNumber(c.Request.Context(), roomId)
			if err != nil {
				log.Println("error getting current game users:", err)
			}
			ctx, cancel := context.WithCancel(c.Request.Context())
			allResultsReceivedCh := g.getAllResultsReceived(ctx, numberGameUsers, roomId)

			gameFinishedCh := time.NewTimer(waitForResultsDurationSeconds * time.Second)

			select {
			case <-gameFinishedCh.C:
				cancel()
			case <-allResultsReceivedCh:
				gameFinishedCh.Stop()
				cancel()
			}

			// publish game results
			// store game/score in DB
		}()
	}

	c.JSON(http.StatusOK, gin.H{"data": "game started"})
}

func (g *Games) getAllResultsReceived(ctx context.Context, playersNumber int, roomId uuid.UUID) <-chan struct{} {
	allReceived := make(chan struct{})

	go func() {
		defer close(allReceived)
		roomActionMessageCh, errCh := g.RoomSubscriberService.GetActions(ctx, roomId, time.Time{})

		for resultsCount := 0; resultsCount < playersNumber; {
			select {
			case roomActionMessage, ok := <-roomActionMessageCh:
				if !ok {
					return
				}
				roomAction := roomActionMessage["action"]
				if roomAction == "result" {
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

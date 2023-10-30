package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (g *Games) IsCurrentGameUser(c *gin.Context) {
	roomId, err := getRoomIdFromPath(c)
	if err != nil {
		log.Println("roomid parameter could not be parsed", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	user, err := getUserFromContext(c)
	if err != nil {
		log.Println("user id could not be retrieved from the request context", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	isCurrentGameUser, err := g.GameService.IsCurrentGameUser(c.Request.Context(), roomId, user.ID)
	switch {
	case err != nil:
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	case !isCurrentGameUser:
		err = errors.New("user:" + user.ID.String() + "is not current game user of room" + roomId.String())
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Next()
}

// checks if gameid parameter identifies the current game that the roomid parameter identifies
func (g *Games) IsCurrentGame(c *gin.Context) {
	roomId, err := getRoomIdFromPath(c)
	if err != nil {
		log.Println("roomid parameter could not be parsed", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	gameId, err := getGameIdFromPath(c)
	if err != nil {
		log.Println("gameid parameter could not be parsed", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	isCurrentGame, err := g.GameService.IsCurrentGame(c.Request.Context(), roomId, gameId)
	switch {
	case err != nil:
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error"})
		return
	case !isCurrentGame:
		err = errors.New("game:" + gameId.String() + "is not current game of room" + roomId.String())
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Next()
}

package main

import (
	"10-typing/models"
	"os"
)

func main() {
	models.ConnectDatabase()

	userService := models.UserService{
		DB: models.DB,
	}
	sessionService := models.SessionService{
		DB: models.DB,
	}
	scoreService := models.ScoreService{
		DB: models.DB,
	}
	textService := models.TextService{
		DB:  models.DB,
		RDB: models.RedisClient,
	}
	roomService := models.RoomService{
		DB:  models.DB,
		RDB: models.RedisClient,
	}

	err := userService.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = sessionService.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = scoreService.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = textService.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = roomService.DeleteAll()
	if err != nil {
		os.Exit(1)
	}
}

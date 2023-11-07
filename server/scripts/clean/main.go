package main

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"
	"os"
)

func main() {
	models.ConnectDatabase()

	textDbRepo := repositories.NewTextDbRepository(models.DB)
	textRedisRepo := repositories.NewTextRedisRepository(models.RedisClient)
	roomDbRepo := repositories.NewRoomDbRepository(models.DB)
	roomRedisRepo := repositories.NewRoomRedisRepository(models.RedisClient)
	scoreDbRepo := repositories.NewScoreDbRepository(models.DB)
	sessionDbRepo := repositories.NewSessionDbRepository(models.DB)
	userDbRepo := repositories.NewUserDbRepository(models.DB)

	err := userDbRepo.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = sessionDbRepo.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = scoreDbRepo.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	err = textDbRepo.DeleteAll()
	if err != nil {
		os.Exit(1)
		return
	}

	var ctx = context.Background()

	err = textRedisRepo.DeleteAllFromRedis(ctx)
	if err != nil {
		os.Exit(1)
	}

	err = roomDbRepo.DeleteAll()
	if err != nil {
		os.Exit(1)
	}

	err = roomRedisRepo.DeleteAllFromRedis(ctx)
	if err != nil {
		os.Exit(1)
	}
}

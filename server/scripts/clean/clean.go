package main

import (
	"10-typing/models"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"
	"context"
	"os"
)

func main() {
	models.ConnectDatabase()

	var ctx = context.Background()
	cacheRepo := redis_repo.NewRedisRepository(models.RedisClient)
	dbRepo := sql_repo.NewSQLRepository(models.DB)

	err := dbRepo.DeleteAllUsers()
	if err != nil {
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteAllUsers(ctx)
	if err != nil {
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteAllSessions(ctx)
	if err != nil {
		os.Exit(1)
		return
	}

	err = dbRepo.DeleteAllScores()
	if err != nil {
		os.Exit(1)
		return
	}

	err = dbRepo.DeleteAllTexts()
	if err != nil {
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteTextIdsKey(ctx)
	if err != nil {
		os.Exit(1)
	}

	err = dbRepo.DeleteAllRooms()
	if err != nil {
		os.Exit(1)
	}

	err = cacheRepo.DeleteAllRooms(ctx)
	if err != nil {
		os.Exit(1)
	}
}
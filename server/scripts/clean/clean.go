package main

import (
	"10-typing/models"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"
	"context"
	"log"
	"os"
)

func main() {
	models.ConnectDatabase()

	var ctx = context.Background()
	cacheRepo := redis_repo.NewRedisRepository(models.RedisClient)
	dbRepo := sql_repo.NewSQLRepository(models.DB)

	err := dbRepo.DeleteAllUsers(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteAllUsers(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteAllSessions(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}

	err = dbRepo.DeleteAllScores(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}

	err = dbRepo.DeleteAllTexts(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
		return
	}

	err = cacheRepo.DeleteTextIdsKey(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	err = dbRepo.DeleteAllRooms(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	err = cacheRepo.DeleteAllRooms(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

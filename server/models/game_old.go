package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func (gs *GameService) CreateOld(tx *gorm.DB, input CreateGameInput, roomId uuid.UUID, userId uuid.UUID) (*Game, error) {
	db := gs.getDbOrTx(tx)

	newGame := Game{
		TextId: input.TextId,
		RoomId: roomId,
		Status: UnstartedGameStatus,
	}

	if err := db.Create(&newGame).Error; err != nil {
		return nil, err
	}

	newSubscriber := Subscriber{
		UserId:         userId,
		StartTimeStamp: nil,
		Status:         UnactiveSubscriberStatus,
	}

	err := gs.addGameSubscriberAndGameStatus(context.Background(), &newGame, &newSubscriber)
	if err != nil {
		return nil, err
	}

	newGame.Subscribers = append(newGame.Subscribers, newSubscriber)

	return &newGame, nil
}

func (gs *GameService) Find(gameId, roomId uuid.UUID, userId uuid.UUID) (*Game, error) {
	game := Game{
		ID: gameId,
	}

	result := gs.DB.
		Joins("INNER JOIN user_rooms ur ON ur.room_id = games.room_id").
		Where("games.room_id = ?", roomId).
		Where("ur.user_id = ?", userId).
		Find(&game)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no game found")
	}

	ctx := context.Background()

	gameStatus, err := gs.getGameStatus(ctx, gameId)
	if err != nil {
		return nil, err
	}

	gameSubscribers, err := gs.getSubscribers(ctx, gameId)
	if err != nil {
		return nil, err
	}

	game.Status = gameStatus
	game.Subscribers = gameSubscribers

	return &game, nil
}

func (gs *GameService) addGameSubscriberAndGameStatus(ctx context.Context, newGame *Game, newSubscriber *Subscriber) error {
	_, err := gs.RDB.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := addGameSubscriber(ctx, pipe, newGame.RoomId, newGame.ID, newSubscriber); err != nil {
			return err
		}

		return addGameStatus(ctx, pipe, newGame.ID, newGame.Status)
	})

	return err
}

func (gs *GameService) getGameStatus(ctx context.Context, gameId uuid.UUID) (GameStatus, error) {
	gameStatusKey := getGameStatusKey(gameId)
	gameStatusStr, err := gs.RDB.Get(ctx, gameStatusKey).Result()
	if err != nil {
		return UnstartedGameStatus, err
	}

	gameStatus, err := strconv.Atoi(gameStatusStr)
	if err != nil {
		return UnstartedGameStatus, err
	}

	return GameStatus(gameStatus), nil
}

func (gs *GameService) getSubscribers(ctx context.Context, gameId uuid.UUID) ([]Subscriber, error) {
	gameUserIdsKey := getGameUserIdsKey(gameId)

	subscriberIds, err := gs.RDB.SMembers(ctx, gameUserIdsKey).Result()
	if err != nil {
		return nil, err
	}

	subscribers := []Subscriber{}
	for _, subscriberId := range subscriberIds {
		subscriber, err := gs.getSubscriber(ctx, gameId, subscriberId)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, *subscriber)
	}

	return subscribers, nil
}

func (gs *GameService) getSubscriber(ctx context.Context, gameId uuid.UUID, subscriberIdStr string) (*Subscriber, error) {
	userDataKey := getUserDataKey(gameId, subscriberIdStr)

	f, err := gs.RDB.HGetAll(ctx, userDataKey).Result()
	if err != nil {
		return nil, err
	}

	subscriber, err := mapRedisHashToSubscriber(f)
	if err != nil {
		return nil, err
	}

	subscriberId, err := uuid.Parse(subscriberIdStr)
	if err != nil {
		return nil, err
	}

	subscriber.UserId = subscriberId

	return subscriber, nil
}

func mapRedisHashToSubscriber(hashData map[string]string) (*Subscriber, error) {
	var subscriber Subscriber

	if val, ok := hashData["status"]; ok {
		statusInt, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}

		subscriber.Status = SubscriberStatus(statusInt)
	}

	if val, ok := hashData["startTime"]; ok {
		startTimeInt, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}

		startTime := time.Unix(int64(startTimeInt), 0)
		subscriber.StartTimeStamp = &startTime
	}

	return &subscriber, nil
}

func addGameStatus(ctx context.Context, pipe redis.Pipeliner, gameId uuid.UUID, status GameStatus) error {
	value := strconv.Itoa(int(status))
	key := getGameStatusKey(gameId)

	return pipe.Set(ctx, key, value, 0).Err()
}

func addGameSubscriber(ctx context.Context, pipe redis.Pipeliner, roomId, gameId uuid.UUID, subscriber *Subscriber) error {
	userIdStr := subscriber.UserId.String()
	roomUnstartedGamesKey := getUnstartedGamesKey(roomId)
	gameUserIdsKey := getGameUserIdsKey(gameId)
	userDataKey := getUserDataKey(gameId, userIdStr)

	err := pipe.SAdd(ctx, roomUnstartedGamesKey, gameId.String()).Err()
	if err != nil {
		return err
	}

	err = pipe.SAdd(ctx, gameUserIdsKey, userIdStr).Err()
	if err != nil {
		return err
	}

	userDataFields := map[string]any{
		"status": strconv.Itoa(int(subscriber.Status)),
	}

	if subscriber.StartTimeStamp != nil {
		userDataFields["startTime"] = subscriber.StartTimeStamp.Unix()
	}

	return pipe.HSet(ctx, userDataKey, userDataFields).Err()
}

func (gs *GameService) getDbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return gs.DB
}

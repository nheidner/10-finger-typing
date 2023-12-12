package main

import (
	"10-typing/errors"
	"10-typing/models"
	open_ai_repo "10-typing/repositories/open_ai"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"
	"10-typing/services"
	"10-typing/zerologger"
	"context"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	userService  *services.UserService
	scoreService *services.ScoreService
	textService  *services.TextService
)

func init() {
	cacheRepo := redis_repo.NewRedisRepository(models.RedisClient)
	dbRepo := sql_repo.NewSQLRepository(models.DB)
	openAiRepo := open_ai_repo.NewOpenAiRepository("")

	zl := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	logger := zerologger.New(zl)

	userService = services.NewUserService(dbRepo, cacheRepo, logger, 32)
	scoreService = services.NewScoreService(dbRepo, logger)
	textService = services.NewTextService(dbRepo, cacheRepo, openAiRepo, logger)
}

func main() {
	var ctx = context.Background()

	users, err := seedUsers(ctx)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	fakeUsers, err := seedFakeUsers(ctx, 5)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	fakeTexts, err := seedFakeTexts(ctx, 20)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	allUsers := append(users, fakeUsers...)

	_, err = seedFakeScores(ctx, allUsers, fakeTexts, 100)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func seedUsers(ctx context.Context) ([]*models.User, error) {
	const op errors.Op = "main.seedUsers"

	userData := []models.User{
		{
			Username:  "niko",
			Email:     "niko@gmail.com",
			FirstName: "Nikolas",
			LastName:  "Heidner",
			Password:  "password1?",
		},
		{
			Username:  "luka",
			Email:     "luka@gmail.com",
			FirstName: "Luka",
			LastName:  "Stärk",
			Password:  "password1?",
		},
		{
			Username:  "przemi",
			Email:     "przemi@gmail.com",
			FirstName: "Przemek",
			LastName:  "Borucki",
			Password:  "password1?",
		},
	}

	users := make([]*models.User, 0, len(userData))

	for _, userInputData := range userData {
		user, err := userService.Create(
			ctx,
			userInputData.Email,
			userInputData.Username,
			userInputData.FirstName,
			userInputData.LastName,
			userInputData.Password,
		)
		if err != nil {
			return nil, errors.E(op, err)
		}

		err = userService.VerifyUser(ctx, user.ID)
		if err != nil {
			return nil, errors.E(op, err)
		}

		users = append(users, user)
	}

	return users, nil
}

func seedFakeUsers(ctx context.Context, n int) ([]*models.User, error) {
	const op errors.Op = "main.seedFakeUsers"

	users := make([]*models.User, 0, n)

	for i := 0; i < n; i++ {
		userInputData, err := generateFakeData[models.User]()
		if err != nil {
			return nil, errors.E(op, err)
		}

		user, err := userService.Create(
			ctx,
			userInputData.Email,
			userInputData.Username,
			userInputData.FirstName,
			userInputData.LastName,
			userInputData.Password,
		)
		if err != nil {
			return nil, errors.E(op, err)
		}

		err = userService.VerifyUser(ctx, user.ID)
		if err != nil {
			return nil, errors.E(op, err)
		}

		users = append(users, user)
	}

	return users, nil
}

func seedFakeTexts(ctx context.Context, n int) ([]*models.Text, error) {
	const op errors.Op = "main.seedFakeTexts"

	texts := make([]*models.Text, 0, n)

	for i := 0; i < n; i++ {
		gptText := "The quick brown fox jumps over the lazy dogs back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890. The quick brown fox jumps over the lazy dogs back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890."
		newText, err := generateFakeData[models.Text]()
		if err != nil {
			return nil, errors.E(op, err)
		}

		text, err := textService.Create(ctx, newText.Language, gptText, newText.Punctuation, newText.SpecialCharacters, newText.Numbers)
		if err != nil {
			return nil, errors.E(op, err)
		}

		texts = append(texts, text)
	}

	return texts, nil
}

func seedFakeScores(ctx context.Context, users []*models.User, texts []*models.Text, n int) ([]*models.Score, error) {
	const op errors.Op = "main.seedFakeScores"

	scores := make([]*models.Score, 0, n)

	for i := 0; i < n; i++ {
		newScore, err := generateFakeData[models.Score]()
		if err != nil {
			return nil, errors.E(op, err)
		}

		randomUsersIndex := rand.Intn(len(users))
		randomTextsIndex := rand.Intn(len(texts))
		randomUser := users[randomUsersIndex]
		randomText := texts[randomTextsIndex]

		randomCharsAmount := rand.Intn(8)
		typingErrors := make(models.ErrorsJSON, randomCharsAmount)
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for i := 0; i < randomCharsAmount; i++ {
			randomChar := string(chars[rand.Intn(len(chars))])
			chars = strings.Replace(chars, randomChar, "", 1)
			randomErrorsAmount := rand.Intn(5) + 1

			typingErrors[randomChar] = randomErrorsAmount
		}

		score, err := scoreService.Create(
			ctx,
			uuid.Nil,
			randomUser.ID,
			randomText.ID,
			newScore.WordsTyped,
			newScore.TimeElapsed,
			typingErrors,
		)
		if err != nil {
			return nil, errors.E(op, err)
		}

		scores = append(scores, score)
	}

	return scores, nil
}

func generateFakeData[T models.User | models.Text | models.Score]() (*T, error) {
	const op errors.Op = "main.generateFakeData"

	inputDataPtr := new(T)
	err := faker.FakeData(inputDataPtr)

	if err != nil {
		return nil, errors.E(op, err)
	}

	return inputDataPtr, nil
}

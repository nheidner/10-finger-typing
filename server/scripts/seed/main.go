package main

import (
	"10-typing/models"
	open_ai_repo "10-typing/repositories/open_ai"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"
	"10-typing/services"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
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

	userService = services.NewUserService(dbRepo, cacheRepo, 32)
	scoreService = services.NewScoreService(dbRepo)
	textService = services.NewTextService(dbRepo, cacheRepo, openAiRepo)
}

func main() {
	users, err := seedUsers()
	if err != nil {
		fmt.Println("error seeding users:", err)
		os.Exit(1)
	}

	fakeUsers, err := seedFakeUsers(5)
	if err != nil {
		fmt.Println("error seeding fake users:", err)
		os.Exit(1)
	}

	fakeTexts, err := seedFakeTexts(20)
	if err != nil {
		fmt.Println("error seeding fake texts:", err)
		os.Exit(1)
	}

	allUsers := append(users, fakeUsers...)

	_, err = seedFakeScores(allUsers, fakeTexts, 100)
	if err != nil {
		fmt.Println("error seeding fake scores:", err)
		os.Exit(1)
	}
}

func seedUsers() ([]*models.User, error) {
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
			userInputData.Email,
			userInputData.Username,
			userInputData.FirstName,
			userInputData.LastName,
			userInputData.Password,
		)
		if err != nil {
			return nil, err
		}

		err = userService.VerifyUser(user.ID)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func seedFakeUsers(n int) ([]*models.User, error) {
	users := make([]*models.User, 0, n)

	for i := 0; i < n; i++ {
		userInputData, err := generateFakeData[models.User]()
		if err != nil {
			return nil, err
		}

		user, err := userService.Create(
			userInputData.Email,
			userInputData.Username,
			userInputData.FirstName,
			userInputData.LastName,
			userInputData.Password,
		)
		if err != nil {
			return nil, err
		}

		err = userService.VerifyUser(user.ID)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func seedFakeTexts(n int) ([]*models.Text, error) {
	texts := make([]*models.Text, 0, n)

	for i := 0; i < n; i++ {
		gptText := "The quick brown fox jumps over the lazy dogs back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890. The quick brown fox jumps over the lazy dogs back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890."
		newText, err := generateFakeData[models.Text]()
		if err != nil {
			return nil, err
		}

		text, err := textService.Create(newText.Language, gptText, newText.Punctuation, newText.SpecialCharacters, newText.Numbers)
		if err != nil {
			return nil, err
		}

		texts = append(texts, text)
	}

	return texts, nil
}

func seedFakeScores(users []*models.User, texts []*models.Text, n int) ([]*models.Score, error) {
	scores := make([]*models.Score, 0, n)

	for i := 0; i < n; i++ {
		newScore, err := generateFakeData[models.Score]()
		if err != nil {
			return nil, errors.New("error generating fake data: " + err.Error())
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
			uuid.Nil,
			randomUser.ID,
			randomText.ID,
			newScore.WordsTyped,
			newScore.TimeElapsed,
			typingErrors,
		)
		if err != nil {
			return nil, errors.New("error creating new score: " + err.Error())
		}

		scores = append(scores, score)
	}

	return scores, nil
}

func generateFakeData[T models.User | models.Text | models.Score]() (*T, error) {
	inputDataPtr := new(T)
	err := faker.FakeData(inputDataPtr)

	if err != nil {
		return nil, err
	}

	return inputDataPtr, nil
}

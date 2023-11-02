package main

import (
	"10-typing/controllers"
	"10-typing/models"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

var (
	userService  *models.UserService
	textService  *models.TextService
	scoreService *models.ScoreService
)

func init() {
	userService = &models.UserService{DB: models.DB}
	textService = &models.TextService{DB: models.DB, RDB: models.RedisClient}
	scoreService = &models.ScoreService{DB: models.DB}
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
	usersInputData := []*models.CreateUserInput{
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

	users := make([]*models.User, 0, len(usersInputData))

	for _, userInputData := range usersInputData {
		user, err := userService.Create(*userInputData)
		if err != nil {
			return nil, err
		}

		err = userService.Verify(user.ID)
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
		userInputData, err := generateFakeData[models.CreateUserInput]()
		if err != nil {
			return nil, err
		}

		user, err := userService.Create(*userInputData)
		if err != nil {
			return nil, err
		}

		err = userService.Verify(user.ID)
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
		gptText := "The quick brown fox jumps over the lazy dog's back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890. The quick brown fox jumps over the lazy dog's back. The five boxing wizards jump quickly. Special characters: @#$%^&* (8). Numbers: 12345678. 1234567890. 1234567890."
		textInputData, err := generateFakeData[controllers.CreateTextInput]()
		if err != nil {
			return nil, err
		}

		text, err := textService.Create(context.Background(), *textInputData, gptText)
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
		scoreInputData, err := generateFakeData[controllers.CreateScoreInput]()
		if err != nil {
			return nil, errors.New("error generating fake data: " + err.Error())
		}

		randomUsersIndex := rand.Intn(len(users))
		randomTextsIndex := rand.Intn(len(texts))

		randomUser := users[randomUsersIndex]
		randomText := texts[randomTextsIndex]

		scoreInputData.UserId = randomUser.ID
		scoreInputData.TextId = randomText.ID
		scoreInputData.GameId = uuid.Nil

		randomCharsAmount := rand.Intn(8)

		typingErrors := make(models.ErrorsJSON, randomCharsAmount)
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		for i := 0; i < randomCharsAmount; i++ {
			randomChar := string(chars[rand.Intn(len(chars))])
			chars = strings.Replace(chars, randomChar, "", 1)
			randomErrorsAmount := rand.Intn(5) + 1

			typingErrors[randomChar] = randomErrorsAmount
		}

		scoreInputData.Errors = typingErrors

		score, err := scoreService.Create(*scoreInputData)
		if err != nil {
			return nil, errors.New("error creating new score: " + err.Error())
		}

		scores = append(scores, score)
	}

	return scores, nil
}

func generateFakeData[T models.CreateUserInput | controllers.CreateTextInput | controllers.CreateScoreInput]() (*T, error) {
	inputDataPtr := new(T)
	err := faker.FakeData(inputDataPtr)

	if err != nil {
		return nil, err
	}

	return inputDataPtr, nil
}

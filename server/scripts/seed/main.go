package main

import (
	"10-typing/models"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/go-faker/faker/v4"
)

var (
	userService  *models.UserService
	textService  *models.TextService
	scoreService *models.ScoreService
)

func init() {
	userService = &models.UserService{DB: models.DB}
	textService = &models.TextService{DB: models.DB}
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
		fmt.Println("error seeding fake users:", err)
		os.Exit(1)
	}

	allUsers := append(users, fakeUsers...)

	_, err = seedFakeScores(allUsers, fakeTexts, 100)
	if err != nil {
		fmt.Println("error seeding fake users:", err)
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
		textInputData, err := generateFakeData[models.CreateTextInput]()
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
		scoreInputData, err := generateFakeData[models.CreateScoreInput]()
		if err != nil {
			return nil, err
		}

		randomUsersIndex := rand.Intn(len(users))
		randomTextsIndex := rand.Intn(len(texts))

		randomUser := users[randomUsersIndex]
		randomText := texts[randomTextsIndex]

		scoreInputData.UserId = randomUser.ID
		scoreInputData.TextId = randomText.ID

		randomCharsAmount := rand.Intn(8)

		errors := make(models.ErrorsJSON, randomCharsAmount)
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		for i := 0; i < randomCharsAmount; i++ {
			randomChar := string(chars[rand.Intn(len(chars))])
			chars = strings.Replace(chars, randomChar, "", 1)
			randomErrorsAmount := rand.Intn(5) + 1

			errors[randomChar] = randomErrorsAmount
		}

		scoreInputData.Errors = errors

		score, err := scoreService.Create(*scoreInputData)
		if err != nil {
			return nil, err
		}

		scores = append(scores, score)
	}

	return scores, nil
}

func generateFakeData[T models.CreateUserInput | models.CreateTextInput | models.CreateScoreInput]() (*T, error) {
	inputDataPtr := new(T)
	err := faker.FakeData(inputDataPtr)

	if err != nil {
		return nil, err
	}

	return inputDataPtr, nil
}

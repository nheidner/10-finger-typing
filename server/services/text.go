package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"

	"github.com/google/uuid"
)

type TextService struct {
	textDbRepo    *repositories.TextDbRepository
	textRedisRepo *repositories.TextRedisRepository
	openAiRepo    *repositories.OpenAiRepository
}

func NewTextService(textDbRepo *repositories.TextDbRepository, textRedisRepo *repositories.TextRedisRepository, openAiRepo *repositories.OpenAiRepository) *TextService {
	return &TextService{textDbRepo, textRedisRepo, openAiRepo}
}

func (ts *TextService) FindNewTextForUser(
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	return ts.textDbRepo.FindNewTextByUserId(
		userId,
		language,
		punctuation,
		specialCharactersGte,
		specialCharactersLte,
		numbersGte,
		numbersLte,
	)
}

func (ts *TextService) Create(
	language string,
	punctuation bool,
	specialCharacters, numbers int,
) (*models.Text, error) {
	gptText, err := ts.openAiRepo.GenerateTypingText(language, punctuation, specialCharacters, numbers)
	if err != nil {
		return nil, err
	}

	newText := models.Text{
		Language:          language,
		Text:              gptText,
		Punctuation:       punctuation,
		SpecialCharacters: specialCharacters,
		Numbers:           numbers,
	}

	var ctx = context.Background()

	createdText, err := ts.textDbRepo.Create(newText)
	if err != nil {
		return nil, err
	}

	// create text in redis and additionally: if no text ids key exists, query all text ids from DB and write them to text ids key
	allTextsAreInRedis, err := ts.textRedisRepo.AllTextsAreInRedis(ctx)
	switch {
	case err != nil:
		return nil, err
	case !allTextsAreInRedis:
		allTextIds, err := ts.textDbRepo.GetAllTextIds()
		if err != nil {
			return nil, err
		}

		allTextIds = append(allTextIds, createdText.ID)

		err = ts.textRedisRepo.CreateInRedis(ctx, allTextIds...)

		return createdText, err
	default:
		err = ts.textRedisRepo.CreateInRedis(ctx, createdText.ID)

		return createdText, err
	}
}

func (ts *TextService) DeleteAll() error {
	err := ts.textDbRepo.DeleteAll()
	if err != nil {
		return err
	}

	return ts.textDbRepo.DeleteAll()
}

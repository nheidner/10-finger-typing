package services

import (
	"10-typing/models"
	"10-typing/repositories"
	"context"

	"github.com/google/uuid"
)

type TextService struct {
	dbRepo     repositories.DBRepository
	cacheRepo  repositories.CacheRepository
	openAiRepo repositories.OpenAiRepository
}

func NewTextService(dbRepo repositories.DBRepository, cacheRepo repositories.CacheRepository, openAiRepo repositories.OpenAiRepository) *TextService {
	return &TextService{dbRepo, cacheRepo, openAiRepo}
}

func (ts *TextService) FindNewTextForUser(
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	return ts.dbRepo.FindNewTextByUserId(
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
	language, text string,
	punctuation bool,
	specialCharacters, numbers int,
) (*models.Text, error) {
	if text == "" {
		gptText, err := ts.openAiRepo.GenerateTypingText(language, punctuation, specialCharacters, numbers)
		if err != nil {
			return nil, err
		}

		text = gptText
	}

	newText := models.Text{
		Language:          language,
		Text:              text,
		Punctuation:       punctuation,
		SpecialCharacters: specialCharacters,
		Numbers:           numbers,
	}

	var ctx = context.Background()

	createdText, err := ts.dbRepo.CreateText(newText)
	if err != nil {
		return nil, err
	}

	// create text in redis and additionally: if no text ids key exists, query all text ids from DB and write them to text ids key
	allTextsAreInRedis, err := ts.cacheRepo.AllTextsAreInCache(ctx)
	switch {
	case err != nil:
		return nil, err
	case !allTextsAreInRedis:
		allTextIds, err := ts.dbRepo.FindAllTextIds()
		if err != nil {
			return nil, err
		}

		allTextIds = append(allTextIds, createdText.ID)

		err = ts.cacheRepo.SetText(ctx, allTextIds...)

		return createdText, err
	default:
		err = ts.cacheRepo.SetText(ctx, createdText.ID)

		return createdText, err
	}
}

func (ts *TextService) DeleteAll() error {
	err := ts.cacheRepo.DeleteAllTextsFromRedis(context.Background())
	if err != nil {
		return err
	}

	return ts.dbRepo.DeleteAllTexts()
}

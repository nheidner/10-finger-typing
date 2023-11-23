package services

import (
	"10-typing/common"
	"10-typing/errors"
	"10-typing/models"

	"context"

	"github.com/google/uuid"
)

type TextService struct {
	dbRepo     common.DBRepository
	cacheRepo  common.CacheRepository
	openAiRepo common.OpenAiRepository
}

func NewTextService(dbRepo common.DBRepository, cacheRepo common.CacheRepository, openAiRepo common.OpenAiRepository) *TextService {
	return &TextService{dbRepo, cacheRepo, openAiRepo}
}

func (ts *TextService) FindNewTextForUser(
	ctx context.Context,
	userId uuid.UUID, language string,
	punctuation bool,
	specialCharactersGte, specialCharactersLte, numbersGte, numbersLte int,
) (*models.Text, error) {
	const op errors.Op = "services.TextService.FindNewTextForUser"

	text, err := ts.dbRepo.FindNewTextForUser(
		userId,
		language,
		punctuation,
		specialCharactersGte,
		specialCharactersLte,
		numbersGte,
		numbersLte,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return text, nil
}

func (ts *TextService) FindTextById(ctx context.Context, textId uuid.UUID) (*models.Text, error) {
	const op errors.Op = "services.TextService.FindTextById"

	text, err := ts.dbRepo.FindTextById(textId)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return text, nil
}

func (ts *TextService) Create(
	ctx context.Context,
	language, text string,
	punctuation bool,
	specialCharacters, numbers int,
) (*models.Text, error) {
	const op errors.Op = "services.TextService.FindNewTextForUser"

	if text == "" {
		gptText, err := ts.openAiRepo.GenerateTypingText(language, punctuation, specialCharacters, numbers)
		if err != nil {
			return nil, errors.E(op, err)
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

	createdText, err := ts.dbRepo.CreateTextAndCache(ctx, ts.cacheRepo, newText)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return createdText, nil
}

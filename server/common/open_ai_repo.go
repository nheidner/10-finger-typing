package common

type OpenAiRepository interface {
	GenerateTypingText(language string, punctuation bool, specialCharacters, numbers int) (string, error)
}

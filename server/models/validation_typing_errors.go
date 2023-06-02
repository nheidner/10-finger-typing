package models

import (
	"github.com/go-playground/validator/v10"
)

var validKeys = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

var TypingErrors validator.Func = func(fl validator.FieldLevel) bool {
	keyErrors, ok := fl.Field().Interface().(ErrorsJSON)

	if !ok {
		return false
	}

	for key := range keyErrors {
		if !keyIsValid(key) {
			return false
		}
	}

	return true
}

func keyIsValid(key string) bool {
	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}

	return false
}

package utils

import (
	"10-typing/errors"
	"10-typing/models"
	"net/http"
)

func NewCookie(name, value string) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		MaxAge:   models.SessionDurationSec,
		Path:     "/",
	}

	return &cookie
}

func SetCookie(w http.ResponseWriter, name, value string) {
	cookie := NewCookie(name, value)
	http.SetCookie(w, cookie)
}

func ReadCookie(r *http.Request, name string) (string, error) {
	const op errors.Op = "utils.ReadCookie"

	cookie, err := r.Cookie(name)
	if err != nil {
		return "", errors.New(op, err)
	}

	return cookie.Value, nil
}

func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := NewCookie(name, "")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

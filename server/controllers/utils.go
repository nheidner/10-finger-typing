package controllers

import (
	"10-typing/models"
	"fmt"
	"net/http"
)

const CookieSession = "SID"

func newCookie(name, value string) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		MaxAge:   models.SessionDuration,
		Path:     "/",
	}

	return &cookie
}

func setCookie(w http.ResponseWriter, name, value string) {
	cookie := newCookie(name, value)
	http.SetCookie(w, cookie)
}

func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}

	return cookie.Value, nil
}

func deleteCookie(w http.ResponseWriter, name string) {
	cookie := newCookie(name, "")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

// strips sensitive user information from users ex
func stripSensitiveUserInformation(users []models.User, exception *models.User) {
	for i := range users {
		if exception != nil && users[i].ID == exception.ID {
			continue
		}

		users[i].FirstName = ""
		users[i].Email = ""
		users[i].LastName = ""
		users[i].IsVerified = false
	}
}

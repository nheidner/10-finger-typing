package utils

import "10-typing/models"

// strips sensitive user information from users ex
func StripSensitiveUserInformation(users []models.User, exception *models.User) {
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

package models

import (
	custom_errors "10-typing/errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `json:"id" gorm:"primary_key"`
	Username     string `json:"username" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Password     string `json:"password" gorm:"-"`
	PasswordHash string `gorm:"not null;type:varchar(510)"`
	FirstName    string `json:"firstName" gorm:"type:varchar(255)"`
	Email        string `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
	LastName     string `json:"lastName" gorm:"type:varchar(255)"`
	IsVerified   bool   `json:"isVerified" gorm:"default:false; not null"`
	CreatedAt    int    `json:"createdAt" gorm:"autoCreateTime"`
	UpdateAt     int    `json:"updateAt" gorm:"autoUpdateTime"`
}

type CreateUserInput struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=255"`
	Password  string `json:"password" binding:"omitempty,min=6,max=255"`
	FirstName string `json:"firstName" binding:"omitempty,min=3,max=255"`
	LastName  string `json:"lastName" binding:"omitempty,min=3,max=255"`
}

type LoginUserInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserService struct {
	DB *gorm.DB
}

const userPwPepper = "secret-random-string"

func (us UserService) Create(input CreateUserInput) (*User, error) {
	pwBytes := []byte(input.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}

	user := User{
		Email:        input.Email,
		Username:     input.Username,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		IsVerified:   false,
		PasswordHash: string(hashedBytes),
	}

	if result := us.DB.Create(&user); result.Error != nil {
		badRequestError := custom_errors.HTTPError{Message: "error creating user", Status: http.StatusBadRequest, Details: result.Error.Error()}
		return nil, badRequestError
	}

	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*User, error) {
	var user User
	result := us.DB.Where("email = ?", email).Find(&user)
	if result.Error != nil {
		badRequestError := custom_errors.HTTPError{Message: "error querying user", Status: http.StatusInternalServerError, Details: result.Error.Error()}
		return nil, badRequestError
	}
	if (user == User{}) {
		badRequestError := custom_errors.HTTPError{Message: "user not found", Status: http.StatusBadRequest}
		return nil, badRequestError
	}

	if !user.IsVerified {
		badRequestError := custom_errors.HTTPError{Message: "user not verified", Status: http.StatusBadRequest, Details: fmt.Sprintf("user %s not verified", user.Email)}
		return nil, badRequestError
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		badRequestError := custom_errors.HTTPError{Message: "invalid password", Status: http.StatusBadRequest, Details: err.Error()}
		return nil, badRequestError
	}

	return &user, nil
}

package models

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"gorm.io/gorm"

	custom_errors "10-typing/errors"
	"10-typing/rand"
)

const (
	MinBytesPerToken = 32
	SessionDuration  = 60 * 60 * 24 * 7 // 1 week in seconds
)

type Session struct {
	*gorm.Model
	ID     uint `json:"id" gorm:"primary_key"`
	UserId uint `json:"userId"`
	// token that is not saved in the database
	Token     string `json:"token" gorm:"-"`
	TokenHash string `json:"tokenHash" gorm:"not null;type:varchar(510)"`
}

type SessionService struct {
	DB            *gorm.DB
	BytesPerToken int
}

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
func (ss *SessionService) Create(userId uint) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}
	session := Session{
		UserId:    userId,
		Token:     token,
		TokenHash: ss.hash(token),
	}
	createResult := ss.DB.Create(&session)
	if (createResult.Error != nil) || (createResult.RowsAffected == 0) {
		internalServerError := custom_errors.HTTPError{Message: "Internal Server Error", Status: http.StatusInternalServerError}
		return nil, internalServerError
	}

	return &session, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	deleteResult := ss.DB.Delete(&Session{}, "token_hash = ?", tokenHash)
	if (deleteResult.Error != nil) || (deleteResult.RowsAffected == 0) {
		notFoundError := custom_errors.HTTPError{Message: "Not Found", Status: http.StatusNotFound}
		return notFoundError
	}

	return nil
}

// returns user related to value of session token
// if session token is not found or expired, returns error
func (ss *SessionService) User(token string) (*User, error) {
	var user User
	tokenHash := ss.hash(token)
	queryUserResult := ss.DB.Joins("inner join sessions on users.id = sessions.user_id").
		Where("sessions.token_hash = ? AND sessions.created_at > ?", tokenHash, time.Now().Add(-SessionDuration*time.Second)).
		First(&user)
	if (queryUserResult.Error != nil) || (queryUserResult.RowsAffected == 0) {
		notFoundError := custom_errors.HTTPError{Message: "Not Found", Status: http.StatusNotFound}
		return nil, notFoundError
	}

	return &user, nil
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))

	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

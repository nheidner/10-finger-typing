package models

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"10-typing/rand"
)

const (
	MinBytesPerToken = 32
	SessionDuration  = 60 * 60 * 24 * 7 // 1 week in seconds
)

type Session struct {
	*gorm.Model
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserId    uuid.UUID `json:"userId"`
	Token     string    `json:"token" gorm:"-"` // token that is not saved in the database
	TokenHash string    `json:"tokenHash" gorm:"not null;type:varchar(510)"`
}

type SessionService struct {
	DB            *gorm.DB
	BytesPerToken int
}

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
func (ss *SessionService) Create(userId uuid.UUID) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, err
	}
	session := Session{
		UserId:    userId,
		Token:     token,
		TokenHash: ss.hash(token),
	}
	createResult := ss.DB.Create(&session)
	if createResult.Error != nil {
		return nil, createResult.Error
	}
	if createResult.RowsAffected == 0 {
		return nil, errors.New("new sessions found")
	}

	return &session, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	deleteResult := ss.DB.Delete(&Session{}, "token_hash = ?", tokenHash)
	if deleteResult.Error != nil {
		return deleteResult.Error
	}

	if deleteResult.RowsAffected == 0 {
		return errors.New("not found")
	}

	return nil
}

// returns user related to value of session token
// if session token is not found or expired, returns error
func (ss *SessionService) User(token string) (*User, error) {
	var user User
	tokenHash := ss.hash(token)
	queryUserResult := ss.DB.Joins("INNER JOIN sessions on users.id = sessions.user_id").
		Where("sessions.token_hash = ? AND sessions.created_at > ?", tokenHash, time.Now().Add(-SessionDuration*time.Second)).
		Find(&user)
	if queryUserResult.Error != nil {
		return nil, queryUserResult.Error
	}
	if queryUserResult.RowsAffected == 0 {
		return nil, errors.New("not found")
	}

	return &user, nil
}

func (ss *SessionService) DeleteAll() error {
	return ss.DB.Exec("TRUNCATE sessions RESTART IDENTITY CASCADE").Error
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))

	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

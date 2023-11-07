package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" faker:"-"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;type:varchar(255)" faker:"username"`
	Password     string    `json:"-" gorm:"-" faker:"password"`
	PasswordHash string    `json:"-" gorm:"not null;type:varchar(510)" faker:"-"`
	FirstName    string    `json:"firstName" gorm:"type:varchar(255)" faker:"first_name"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)" faker:"email"`
	LastName     string    `json:"lastName" gorm:"type:varchar(255)" faker:"last_name"`
	IsVerified   bool      `json:"isVerified" gorm:"default:false; not null" faker:"-"`
	Sessions     []Session `json:"-" faker:"-"`
	Scores       []Score   `json:"-" faker:"-"`
	Rooms        []*Room   `json:"-" gorm:"many2many:user_rooms" faker:"-"`
	RoomsAdmin   []Room    `json:"-" gorm:"foreignKey:AdminId" faker:"-"`
}

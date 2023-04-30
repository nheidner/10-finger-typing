package models

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

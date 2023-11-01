package models

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Password     string    `json:"-" gorm:"-"`
	PasswordHash string    `json:"-" gorm:"not null;type:varchar(510)"`
	FirstName    string    `json:"firstName" gorm:"type:varchar(255)"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
	LastName     string    `json:"lastName" gorm:"type:varchar(255)"`
	IsVerified   bool      `json:"isVerified" gorm:"default:false; not null"`
	Sessions     []Session `json:"-"`
	Scores       []Score   `json:"-"`
	Rooms        []*Room   `json:"-" gorm:"many2many:user_rooms"`
	RoomsAdmin   []Room    `json:"-" gorm:"foreignKey:AdminId"`
}

type CreateUserInput struct {
	Email     string `json:"email" binding:"required,email" faker:"email"`
	Username  string `json:"username" binding:"required,min=3,max=255" faker:"username"`
	Password  string `json:"password" binding:"omitempty,min=6,max=255" faker:"password"`
	FirstName string `json:"firstName" binding:"omitempty,min=3,max=255" faker:"first_name"`
	LastName  string `json:"lastName" binding:"omitempty,min=3,max=255" faker:"last_name"`
}

type LoginUserInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type FindUsersQuery struct {
	Username    string `form:"username"`
	UsernameSub string `form:"username_contains"`
}

type UserService struct {
	DB *gorm.DB
}

const userPwPepper = "secret-random-string"

func (us UserService) FindUsers(query FindUsersQuery) ([]User, error) {
	var users []User
	findUsersDbQuery := us.DB

	if query.Username != "" {
		findUsersDbQuery = findUsersDbQuery.Where("username = ?", query.Username)
	}

	if query.UsernameSub != "" {
		findUsersDbQuery = findUsersDbQuery.Where("username ILIKE ?", "%"+query.UsernameSub+"%")
	}

	findUsersDbQuery.Find(&users)

	if findUsersDbQuery.Error != nil {
		return nil, findUsersDbQuery.Error
	}
	return users, nil
}

func (us *UserService) FindByEmail(email string) (*User, error) {
	var user User

	if err := us.DB.Where("email = ?", email).Find(&user).Error; err != nil || user.Email == "" {
		return nil, err
	}

	return &user, nil
}

func (us UserService) FindOneById(id uuid.UUID) (*User, error) {
	user := User{
		ID: id,
	}
	result := us.DB.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (us UserService) Create(input CreateUserInput) (*User, error) {
	pwBytes := []byte(input.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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
		return nil, result.Error
	}

	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*User, error) {
	var user User
	result := us.DB.Where("email = ?", email).Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	user.Sessions = []Session{}
	if user.Email == "" {
		return nil, errors.New("user not found")
	}

	if !user.IsVerified {
		return nil, errors.New("user not verified")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		return nil, errors.New("invalid password: " + err.Error())
	}

	return &user, nil
}

func (us *UserService) Verify(userId uuid.UUID) error {
	return us.DB.Model(&User{}).Where("id = ?", userId).Update("is_verified", true).Error
}

func (us *UserService) DeleteAll() error {
	return us.DB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error
}

// strips sensitive user information from users ex
func StripSensitiveUserInformation(users []User, exception *User) {
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

package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "db"
	user     = "typing"
	password = "password"
	dbname   = "typing"
	port     = "5432"
)

var DB *gorm.DB

func init() {
	ConnectDatabase()
}

func ConnectDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		})

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Berlin",
		host, port, user, password, dbname)
	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic("Failed to connect to database!")
	}

	err = db.AutoMigrate(&User{}, &Session{}, &Text{}, &Game{}, &Score{}, &Room{}, &Token{})
	if err != nil {
		panic("Failed to migrate database!")
	}

	DB = db
}

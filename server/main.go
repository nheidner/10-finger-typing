package main

import (
	"10-typing/controllers"
	"10-typing/models"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("typingerrors", models.TypingErrors)
	}

	models.ConnectDatabase()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	// Setup our model services
	userService := models.UserService{
		DB: models.DB,
	}
	sessionService := models.SessionService{
		DB: models.DB,
	}
	scoreService := models.ScoreService{
		DB: models.DB,
	}
	textService := models.TextService{
		DB: models.DB,
	}

	// Setup our controllers
	userController := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	scoreController := controllers.Scores{
		ScoreService: &scoreService,
	}
	textController := controllers.Texts{
		TextService: &textService,
	}

	api := router.Group("/api")

	api.Use(cors.New(cors.Config{
		// todo AllowOrigins based on production or development environment
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// USERS
	api.GET("/users", userController.AuthRequired, userController.FindUsers)
	api.GET("/users/:userid", userController.AuthRequired, userController.FindUser)
	api.GET("/users/:userid/scores", userController.AuthRequired, scoreController.FindScoresByUser)
	api.POST("/users/:userid/scores", userController.AuthRequired, userController.UserIdUrlParamMatchesAuthorizedUser, scoreController.CreateScore)
	api.GET("/users/:userid/text", userController.AuthRequired, userController.UserIdUrlParamMatchesAuthorizedUser, textController.FindText)
	api.POST("/users", userController.CreateUser)

	// USER
	api.GET("/user", userController.AuthRequired, userController.CurrentUser)
	api.POST("/user/login", userController.Login)
	api.POST("/user/logout", userController.AuthRequired, userController.Logout)

	// SCORES
	api.GET("/scores", userController.AuthRequired, scoreController.FindScores)

	// TEXTS
	api.POST("/texts", userController.AuthRequired, textController.CreateText)

	// websocket
	api.GET("/ws", userController.AuthRequired, controllers.Websocket)

	router.Run()
}

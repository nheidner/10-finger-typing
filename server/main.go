package main

import (
	"10-typing/controllers"
	"10-typing/models"
	"os"
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
		DB:  models.DB,
		RDB: models.RedisClient,
	}
	roomService := models.RoomService{
		DB:  models.DB,
		RDB: models.RedisClient,
	}
	tokenService := models.TokenService{
		DB: models.DB,
	}
	gameService := models.GameService{
		DB:  models.DB,
		RDB: models.RedisClient,
	}
	openAiService := models.OpenAiService{
		ApiKey: os.Getenv("OPENAI_API_KEY"),
	}
	emailTransactionService := models.EmailTransactionService{
		ApiKey: os.Getenv("POSTMARK_API_KEY"),
	}
	roomSubscriberService := models.RoomSubscriberService{
		RDB: models.RedisClient,
	}
	roomStreamService := models.RoomStreamService{
		RDB: models.RedisClient,
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
		TextService:   &textService,
		OpenAiService: &openAiService,
	}
	roomController := controllers.Rooms{
		RoomService:             &roomService,
		TokenService:            &tokenService,
		UserService:             &userService,
		EmailTransactionService: &emailTransactionService,
		RoomSubscriberService:   &roomSubscriberService,
		GameService:             &gameService,
		RoomStreamService:       &roomStreamService,
	}
	gameController := controllers.Games{
		GameService:           &gameService,
		RoomService:           &roomService,
		TextService:           &textService,
		RoomSubscriberService: &roomSubscriberService,
		RoomStreamService:     &roomStreamService,
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
	// why use the userId here -> without a user id the middleware function UserIdUrlParamMatchesAuthorizedUser would be unnecessary
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

	// ROOMS
	api.GET("/rooms/:roomid/ws", userController.AuthRequired, roomController.IsRoomMember, roomController.ConnectToRoom)
	api.POST("/rooms", userController.AuthRequired, roomController.CreateRoom)
	api.POST("/rooms/:roomid/leave", userController.AuthRequired, roomController.IsRoomMember, roomController.LeaveRoom)
	api.POST("/rooms/:roomid/games", userController.AuthRequired, roomController.IsRoomAdmin, gameController.CreateGame)
	api.POST("/rooms/:roomid/start_game", userController.AuthRequired, roomController.IsRoomMember, gameController.StartGame)
	api.POST("/rooms/:roomid/game/score", userController.AuthRequired, roomController.IsRoomMember, gameController.IsCurrentGameUser, gameController.FinishGame)

	router.Run()
}

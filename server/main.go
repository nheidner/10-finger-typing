package main

import (
	"10-typing/controllers"
	"10-typing/middlewares"

	"10-typing/models"
	"10-typing/repositories"
	"10-typing/services"
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

	// Setup repos
	emailTransactionRepo := repositories.NewEmailTransactionRepository(os.Getenv("POSTMARK_API_KEY"))
	gameRedisRepo := repositories.NewGameRedisRepository(models.RedisClient)
	openAiRepo := repositories.NewOpenAiRepository(os.Getenv("OPENAI_API_KEY"))
	roomDbRepo := repositories.NewRoomDbRepository(models.DB)
	roomRedisRepo := repositories.NewRoomRedisRepository(models.RedisClient)
	roomStreamRedisRepo := repositories.NewRoomStreamRedisRepository(models.RedisClient)
	roomSubscriberRedisRepo := repositories.NewRoomSubscriberRedisRepository(models.RedisClient)
	scoreDbRepo := repositories.NewScoreDbRepository(models.DB)
	sessionDbRepo := repositories.NewSessionDbRepository(models.DB)
	textDbRepo := repositories.NewTextDbRepository(models.DB)
	textRedisRepo := repositories.NewTextRedisRepository(models.RedisClient)
	tokenDbRepo := repositories.NewTokenDbRepository(models.DB)
	userDbRepo := repositories.NewUserDbRepository(models.DB)
	userRoomDbRepo := repositories.NewUserRoomDbRepository(models.DB)

	// Setup services
	gameService := services.NewGameService(gameRedisRepo, roomStreamRedisRepo, scoreDbRepo, textRedisRepo)
	roomService := services.NewRoomService(roomDbRepo, roomRedisRepo, userRoomDbRepo, roomStreamRedisRepo, roomSubscriberRedisRepo, userDbRepo, tokenDbRepo, emailTransactionRepo, gameRedisRepo)
	scoreService := services.NewScoreService(scoreDbRepo)
	textService := services.NewTextService(textDbRepo, textRedisRepo, openAiRepo)
	userService := services.NewUserService(userDbRepo, sessionDbRepo, 32)

	// Setup controllers
	gameController := controllers.NewGameController(gameService)
	roomController := controllers.NewRoomController(roomService)
	scoreController := controllers.NewScoreController(scoreService)
	textController := controllers.NewTextController(textService)
	userController := controllers.NewUserController(userService)

	api := router.Group("/api")

	api.Use(cors.New(cors.Config{
		// todo AllowOrigins based on production or development environment
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authRequiredMiddleware := middlewares.AuthRequired(userDbRepo)

	// USERS
	api.GET("/users", authRequiredMiddleware, userController.FindUsers)
	api.GET("/users/:userid", authRequiredMiddleware, userController.FindUser)
	api.GET("/users/:userid/scores", authRequiredMiddleware, scoreController.FindScoresByUser)
	api.POST("/users/:userid/scores", authRequiredMiddleware, middlewares.UserIdUrlParamMatchesAuthorizedUser(), scoreController.CreateScore)
	// why use the userId here -> without a user id the middleware function UserIdUrlParamMatchesAuthorizedUser would be unnecessary
	api.GET("/users/:userid/text", authRequiredMiddleware, middlewares.UserIdUrlParamMatchesAuthorizedUser(), textController.FindText)
	api.POST("/users", userController.CreateUser)

	// USER
	api.GET("/user", authRequiredMiddleware, userController.CurrentUser)
	api.POST("/user/login", userController.Login)
	api.POST("/user/logout", authRequiredMiddleware, userController.Logout)

	// SCORES
	api.GET("/scores", authRequiredMiddleware, scoreController.FindScores)

	// TEXTS
	api.POST("/texts", authRequiredMiddleware, textController.CreateText)

	// ROOMS
	api.GET("/rooms/:roomid/ws", authRequiredMiddleware, middlewares.IsRoomMember(roomRedisRepo), roomController.ConnectToRoom)
	api.POST("/rooms", authRequiredMiddleware, roomController.CreateRoom)
	api.POST("/rooms/:roomid/leave", authRequiredMiddleware, middlewares.IsRoomMember(roomRedisRepo), roomController.LeaveRoom)
	api.POST("/rooms/:roomid/games", authRequiredMiddleware, middlewares.IsRoomAdmin(roomRedisRepo), gameController.CreateGame)
	api.POST("/rooms/:roomid/start_game", authRequiredMiddleware, middlewares.IsRoomMember(roomRedisRepo), gameController.StartGame)
	api.POST("/rooms/:roomid/game/score",
		authRequiredMiddleware,
		middlewares.IsRoomMember(roomRedisRepo),
		middlewares.IsCurrentGameUser(gameRedisRepo),
		gameController.FinishGame,
	)

	router.Run()
}

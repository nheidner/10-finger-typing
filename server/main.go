package main

import (
	"10-typing/controllers"
	"10-typing/middlewares"
	email_transaction_repo "10-typing/repositories/email_transaction"
	open_ai_repo "10-typing/repositories/open_ai"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"
	"10-typing/zerologger"
	"runtime"

	"10-typing/models"
	"10-typing/services"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	// Zerolog configuration
	logFile, err := os.Create("logfile.log")
	if err != nil {
		panic("Error creating logfile.log: >> " + err.Error())
	}
	defer logFile.Close()

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)
	logger := zerologger.New(multi)

	logger.Info("GOMAXPROCS: >>", runtime.GOMAXPROCS(0))

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("typingerrors", models.TypingErrors)
	}

	// Setup repos
	dbRepo := sql_repo.NewSQLRepository(models.DB)
	cacheRepo := redis_repo.NewRedisRepository(models.RedisClient)
	emailTransactionRepo := email_transaction_repo.NewEmailTransactionRepository(os.Getenv("POSTMARK_API_KEY"))
	openAiRepo := open_ai_repo.NewOpenAiRepository(os.Getenv("OPENAI_API_KEY"))

	// Setup services
	gameService := services.NewGameService(dbRepo, cacheRepo, logger)
	roomService := services.NewRoomService(dbRepo, cacheRepo, emailTransactionRepo, logger)
	scoreService := services.NewScoreService(dbRepo, logger)
	textService := services.NewTextService(dbRepo, cacheRepo, openAiRepo, logger)
	userService := services.NewUserService(dbRepo, cacheRepo, logger, 32)
	userNoticationService := services.NewUserNotificationService(cacheRepo, logger)

	// Setup controllers
	gameController := controllers.NewGameController(gameService, logger)
	roomController := controllers.NewRoomController(roomService, logger)
	scoreController := controllers.NewScoreController(scoreService, logger)
	textController := controllers.NewTextController(textService, logger)
	userController := controllers.NewUserController(userService, logger)
	userNoticationController := controllers.NewUserNotificationController(userNoticationService, logger)

	cors := cors.New(cors.Config{
		// todo AllowOrigins based on production or development environment
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	router := gin.New()
	router.Use(middlewares.GinZerologLogger(logger), gin.Recovery(), cors)
	api := router.Group("/api")

	authRequiredMiddleware := middlewares.AuthRequired(cacheRepo, dbRepo, logger)
	isRoomMemberMiddleware := middlewares.IsRoomMember(cacheRepo, logger)
	isRoomAdminMiddleware := middlewares.IsRoomAdmin(cacheRepo, logger)
	isCurrentGameUserMiddleware := middlewares.IsCurrentGameUser(cacheRepo, logger)
	userIdUrlParamMatchesAuthorizedUserMiddleware := middlewares.UserIdUrlParamMatchesAuthorizedUser(logger)

	// USERS
	api.GET("/users", authRequiredMiddleware, userController.FindUsers)
	api.GET("/users/:userid", authRequiredMiddleware, userController.FindUser)
	api.GET("/users/:userid/scores", authRequiredMiddleware, scoreController.FindScoresByUser)
	api.POST("/users/:userid/scores", authRequiredMiddleware, userIdUrlParamMatchesAuthorizedUserMiddleware, scoreController.CreateScore)
	// why use the userId here -> without a user id the middleware function UserIdUrlParamMatchesAuthorizedUser would be unnecessary
	api.GET("/users/:userid/text", authRequiredMiddleware, userIdUrlParamMatchesAuthorizedUserMiddleware, textController.FindNewTextForUser)
	api.POST("/users", userController.CreateUser)

	// USER
	api.GET("/user", authRequiredMiddleware, userController.CurrentUser)
	api.POST("/user/login", userController.Login)
	api.POST("/user/logout", authRequiredMiddleware, userController.Logout)

	// NOTIFICATIONS
	api.GET("/notification/realtime", authRequiredMiddleware, userNoticationController.FindRealtimeUserNotification)

	// SCORES
	api.GET("/scores", authRequiredMiddleware, scoreController.FindScores)

	// TEXTS
	api.POST("/texts", authRequiredMiddleware, textController.CreateText)
	api.GET("/texts/:textid", authRequiredMiddleware, textController.FindTextById)

	// ROOMS
	api.GET("/rooms/:roomid/ws", authRequiredMiddleware, isRoomMemberMiddleware, roomController.ConnectToRoom)
	// TODO: get new text for room
	// api.GET("/rooms/:roomid/text", authRequiredMiddleware, isRoomAdminMiddleware)
	api.POST("/rooms", authRequiredMiddleware, roomController.CreateRoom)
	api.POST("/rooms/:roomid/leave", authRequiredMiddleware, isRoomMemberMiddleware, roomController.LeaveRoom)
	api.POST("/rooms/:roomid/game", authRequiredMiddleware, isRoomAdminMiddleware, gameController.CreateNewCurrentGame)
	api.POST("/rooms/:roomid/start-game", authRequiredMiddleware, isRoomMemberMiddleware, gameController.StartGame)
	api.POST("/rooms/:roomid/current-game/score",
		authRequiredMiddleware,
		isRoomMemberMiddleware,
		isCurrentGameUserMiddleware,
		gameController.FinishGame,
	)

	router.Run()
}

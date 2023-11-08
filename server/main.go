package main

import (
	"10-typing/controllers"
	"10-typing/middlewares"
	email_transaction_repo "10-typing/repositories/email_transaction"
	open_ai_repo "10-typing/repositories/open_ai"
	redis_repo "10-typing/repositories/redis"
	sql_repo "10-typing/repositories/sql"

	"10-typing/models"
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
	dbRepo := sql_repo.NewSQLRepository(models.DB)
	cacheRepo := redis_repo.NewRedisRepository(models.RedisClient)
	emailTransactionRepo := email_transaction_repo.NewEmailTransactionRepository(os.Getenv("POSTMARK_API_KEY"))
	openAiRepo := open_ai_repo.NewOpenAiRepository(os.Getenv("OPENAI_API_KEY"))

	// Setup services
	gameService := services.NewGameService(dbRepo, cacheRepo)
	roomService := services.NewRoomService(dbRepo, cacheRepo, emailTransactionRepo)
	scoreService := services.NewScoreService(dbRepo)
	textService := services.NewTextService(dbRepo, cacheRepo, openAiRepo)
	userService := services.NewUserService(dbRepo, 32)

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

	authRequiredMiddleware := middlewares.AuthRequired(dbRepo)

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
	api.GET("/rooms/:roomid/ws", authRequiredMiddleware, middlewares.IsRoomMember(cacheRepo), roomController.ConnectToRoom)
	api.POST("/rooms", authRequiredMiddleware, roomController.CreateRoom)
	api.POST("/rooms/:roomid/leave", authRequiredMiddleware, middlewares.IsRoomMember(cacheRepo), roomController.LeaveRoom)
	api.POST("/rooms/:roomid/games", authRequiredMiddleware, middlewares.IsRoomAdmin(cacheRepo), gameController.CreateGame)
	api.POST("/rooms/:roomid/start_game", authRequiredMiddleware, middlewares.IsRoomMember(cacheRepo), gameController.StartGame)
	api.POST("/rooms/:roomid/game/score",
		authRequiredMiddleware,
		middlewares.IsRoomMember(cacheRepo),
		middlewares.IsCurrentGameUser(cacheRepo),
		gameController.FinishGame,
	)

	router.Run()
}

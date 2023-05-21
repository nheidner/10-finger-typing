package main

import (
	"10-typing/controllers"
	"10-typing/models"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

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

	// Setup our controllers
	userController := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}

	api := router.Group("/api")

	api.Use(cors.New(cors.Config{
		// todo AllowOrigins based on production or development environment
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.AuthRequired, userController.FindUser)
	api.POST("/users/login", userController.Login)
	api.POST("/users/logout", userController.AuthRequired, userController.Logout)
	api.GET("/user", userController.AuthRequired, userController.CurrentUser)

	router.Run()
}

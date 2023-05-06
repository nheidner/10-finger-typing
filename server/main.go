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

	api := router.Group("/api")

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	// Setup our model services
	userService := models.UserService{
		DB: models.DB,
	}

	// Setup our controllers
	userController := controllers.Users{
		UserService: &userService,
	}

	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.FindUser)
	api.POST("/users/login", userController.Login)

	router.Run()
}

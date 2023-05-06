package main

import (
	"10-typing/controllers"
	"10-typing/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	models.ConnectDatabase()

	api := r.Group("/api")

	r.GET("/", func(c *gin.Context) {
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

	api.POST("/users", userController.CreateUser)
	api.GET("/users/:id", userController.FindUser)
	api.POST("/users/login", userController.Login)

	r.Run()
}

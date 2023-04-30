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

	api.GET("/books", controllers.FindBooks)
	api.POST("/users", controllers.CreateUser)
	api.POST("/users/login", controllers.Authenticate)

	r.Run()
}

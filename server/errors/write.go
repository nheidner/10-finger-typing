package errors

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientReporter interface {
	Message() Messages
	Status() int
}

func WriteError(c *gin.Context, err error) {
	if cr, ok := err.(ClientReporter); ok {
		status := cr.Status()
		message := cr.Message()
		log.Print(err)

		c.JSON(status, message)
		return
	}

	log.Println(err)
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
}

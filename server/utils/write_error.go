package utils

import (
	"10-typing/common"
	"10-typing/errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, err error, logger common.Logger) {
	if cr, ok := err.(interface {
		Message() errors.Messages
		Status() int
	}); ok {
		status := cr.Status()
		message := cr.Message()
		log.Print(err)

		c.JSON(status, message)
		return
	}

	logger.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
}

package middlewares

import (
	"10-typing/common"
	"time"

	"github.com/gin-gonic/gin"
)

func GinZerologLogger(logger common.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		logger.RequestInfo(method, path, clientIP, statusCode, latency)
	}
}

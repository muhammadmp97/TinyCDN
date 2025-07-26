package middlewares

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LogSlowRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Milliseconds()

		if c.FullPath() == "/g/:domain" && duration > 500 {
			log.Printf("⚠️ SLOW: %dms - %s", duration, c.Request.RequestURI)
		}
	}
}

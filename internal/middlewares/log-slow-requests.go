package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhammadmp97/TinyCDN/internal/app"
	"go.uber.org/zap"
)

func LogSlowRequests(app *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Milliseconds()

		if c.FullPath() == "/g/:domain" && duration > 500 {
			app.Logger.Warn("Slow request",
				zap.Int64("duration", duration),
				zap.String("url", c.Request.RequestURI),
			)
		}
	}
}

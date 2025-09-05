package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muhammadmp97/TinyCDN/internal/app"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/prometheus"
	"github.com/muhammadmp97/TinyCDN/internal/redis"
)

func ServeFileHandler(app *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		found, domain := models.GetDomain(c.Param("domain"))
		if !found {
			c.String(404, "Domain not found!")
			return
		}

		found, hit, file := redis.GetFile(c, app.Config, app.Redis, app.MinIO, domain, c.Query("file"), c.Request.Header)
		if !found {
			c.String(404, "File not found!")
			return
		}

		start := time.Now()
		defer func() {
			elapsed := time.Since(start).Seconds()
			prometheus.ServeLatency.WithLabelValues(domain.Name).Observe(elapsed)
		}()

		if hit {
			c.Header("Cache-Status", "HIT")
			prometheus.CacheHit.WithLabelValues(domain.Name).Inc()
		} else {
			c.Header("Cache-Status", "MISS")
			prometheus.CacheMiss.WithLabelValues(domain.Name).Inc()
		}

		if file.TTL > 0 {
			c.Header("Cache-Control", fmt.Sprintf("max-age=%d", file.TTL))
		}

		if file.Encoding == models.EncodingGZIP {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
		}

		c.Header("Server", "TinyCDN")
		c.Header("Content-Length", strconv.Itoa(file.Size))
		c.Header("Content-Type", file.Type)
		c.String(200, file.Content)
	}
}

package main

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/muhammadmp97/TinyCDN/internal/app"
	"github.com/muhammadmp97/TinyCDN/internal/config"
	"github.com/muhammadmp97/TinyCDN/internal/handlers"
	"github.com/muhammadmp97/TinyCDN/internal/middlewares"
	"github.com/muhammadmp97/TinyCDN/internal/minio"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/prometheus"
	"github.com/muhammadmp97/TinyCDN/internal/redis"
	prometheusPkg "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	prometheusPkg.MustRegister(prometheus.CacheHit)
	prometheusPkg.MustRegister(prometheus.CacheMiss)
	prometheusPkg.MustRegister(prometheus.ServeLatency)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()

	rdb := redis.NewClient(cfg)
	defer rdb.Close()

	mio, err := minio.NewClient(cfg)
	if err != nil {
		panic("Minio connection failed!")
	}

	app := app.New(cfg, logger, rdb, mio)

	loadDomains(app.Config)

	router.Use(middlewares.LogSlowRequests(app))

	router.GET("/g/:domain", handlers.ServeFileHandler(app))

	router.POST("/purge/:domain", handlers.PurgeHandler(app))

	router.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	router.Run()
}

func loadDomains(cfg *config.Config) {
	file, err := os.Open(cfg.DomainsJsonFilePath)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&models.Domains)
	if err != nil {
		panic(err.Error())
	}
}

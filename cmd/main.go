package main

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/muhammadmp97/TinyCDN/internal/handlers"
	"github.com/muhammadmp97/TinyCDN/internal/middlewares"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/prometheus"
	"github.com/muhammadmp97/TinyCDN/internal/redis"
	prometheusPkg "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	prometheusPkg.MustRegister(prometheus.CacheHit)
	prometheusPkg.MustRegister(prometheus.CacheMiss)

	router := gin.Default()
	router.Use(middlewares.LogSlowRequests())

	err := godotenv.Load()
	if err != nil {
		panic("No .env file found!")
	}

	rdb := redis.NewClient()
	defer rdb.Close()

	loadDomains()

	router.GET("/g/:domain", handlers.ServeFileHandler(rdb))

	router.POST("/purge/:domain", handlers.PurgeHandler(rdb))

	router.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	router.Run()
}

func loadDomains() {
	file, err := os.Open(os.Getenv("DOMAINS_JSON_FILE_PATH"))
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

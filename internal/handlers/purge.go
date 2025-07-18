package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/redis"
	redisPkg "github.com/redis/go-redis/v9"
)

func PurgeHandler(rdb *redisPkg.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		found, domain := models.GetDomain(c.Param("domain"))
		if !found {
			c.JSON(404, gin.H{"message": "Domain not found!"})
			return
		}

		if domain.Token != c.GetHeader("Authorization") {
			c.JSON(401, gin.H{"message": "Authorization failed!"})
			return
		}

		totalDeleted := redis.Purge(rdb, domain, c.Query("file"))

		c.JSON(200, gin.H{"total_deleted": totalDeleted})
	}
}

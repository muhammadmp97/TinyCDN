package redis

import (
	"context"

	"github.com/muhammadmp97/TinyCDN/internal/config"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
}

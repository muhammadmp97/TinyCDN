package app

import (
	"github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/config"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config *config.Config

	Redis *redis.Client
	MinIO *minio.Client
}

func New(cfg *config.Config, rdb *redis.Client, mio *minio.Client) *App {
	return &App{Config: cfg, Redis: rdb, MinIO: mio}
}

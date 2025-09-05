package app

import (
	"github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger

	Redis *redis.Client
	MinIO *minio.Client
}

func New(cfg *config.Config, logger *zap.Logger, rdb *redis.Client, mio *minio.Client) *App {
	return &App{Config: cfg, Logger: logger, Redis: rdb, MinIO: mio}
}

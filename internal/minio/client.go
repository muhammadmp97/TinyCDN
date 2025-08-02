package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/muhammadmp97/TinyCDN/internal/config"
)

func NewClient(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.MinIOAddress, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOUser, cfg.MinIOPassword, ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

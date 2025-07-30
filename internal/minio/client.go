package minio

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient() (*minio.Client, error) {
	minioClient, err := minio.New(os.Getenv("MINIO_ADDRESS"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_USER"), os.Getenv("MINIO_PASSWORD"), ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

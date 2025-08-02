package minio

import (
	"bytes"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/config"
)

func Put(c context.Context, cfg *config.Config, mio *minio.Client, objectName string, content string, contentType string) (string, error) {
	reader := bytes.NewReader([]byte(content))

	info, err := mio.PutObject(c, cfg.MinIOBucketName, objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return "", err
	}

	return info.Key, nil
}

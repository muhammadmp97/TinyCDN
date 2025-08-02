package minio

import (
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/config"
)

func Get(c context.Context, cfg *config.Config, mio *minio.Client, objectName string) (string, error) {
	object, err := mio.GetObject(c, cfg.MinIOBucketName, objectName, minio.GetObjectOptions{})

	if err != nil {
		return "", err
	}

	var sb strings.Builder
	_, err = io.Copy(&sb, object)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

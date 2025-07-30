package minio

import (
	"bytes"
	"context"
	"os"

	"github.com/minio/minio-go/v7"
)

func Put(c context.Context, mio *minio.Client, objectName string, content string, contentType string) (string, error) {
	reader := bytes.NewReader([]byte(content))

	info, err := mio.PutObject(c, os.Getenv("MINIO_BUCKET_NAME"), objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return "", err
	}

	return info.Key, nil
}

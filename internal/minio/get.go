package minio

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
)

func Get(c context.Context, mio *minio.Client, objectName string) (string, error) {
	object, err := mio.GetObject(c, os.Getenv("MINIO_BUCKET_NAME"), objectName, minio.GetObjectOptions{})

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

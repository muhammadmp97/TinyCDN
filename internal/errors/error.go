package errors

import (
	"errors"
)

var (
	ErrRedisCannotGet     = errors.New("couldn't get the file from Redis")
	ErrMinIOCannotGet     = errors.New("couldn't get the file from MinIO")
	ErrMinIOCannotPut     = errors.New("couldn't get the file from MinIO")
	ErrOriginFileNotFound = errors.New("couldn't get the file from the origin")
	ErrFileSizeLimit      = errors.New("file is too large")
)

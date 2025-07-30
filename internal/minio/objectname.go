package minio

import (
	"fmt"
	"path"
	"time"

	"github.com/muhammadmp97/TinyCDN/internal/utils"
)

func MakeObjectName(filePath string) string {
	now := time.Now().UTC()
	timestamp := now.Format("200601021504")

	ext := path.Ext(filePath)

	return fmt.Sprintf("%s_%s%s", timestamp, utils.XXHash(filePath), ext)
}

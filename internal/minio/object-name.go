package minio

import (
	"fmt"
	"time"

	"github.com/muhammadmp97/TinyCDN/internal/utils"
)

func MakeObjectName(filePath string) string {
	now := time.Now().UTC()
	timestamp := now.Format("200601021504")

	ext := utils.GetExtension(filePath)

	var directory string
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".avif":
		directory = "photos"
	case ".css", ".js":
		directory = "assets"
	case ".woff", ".woff2", ".ttf", ".otf":
		directory = "fonts"
	default:
		directory = "files"
	}

	return fmt.Sprintf("%s/%s_%s%s", directory, timestamp, utils.XXHash(filePath), ext)
}

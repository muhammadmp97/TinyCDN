package redis

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/muhammadmp97/TinyCDN/internal/app"
	"github.com/muhammadmp97/TinyCDN/internal/minio"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
	"go.uber.org/zap"
)

func GetFile(c context.Context, app *app.App, domain models.Domain, filePath string, headers http.Header) (found bool, hit bool, file models.File) {
	acceptsGzipAndIsCompressible := strings.Contains(headers.Get("Accept-Encoding"), "gzip") && utils.IsCompressible(filePath)
	encoding := models.EncodingNone
	if acceptsGzipAndIsCompressible {
		encoding = models.EncodingGZIP
	}

	redisKey := utils.MakeRedisKey(domain, filePath, acceptsGzipAndIsCompressible)
	redisFile, err := app.Redis.HGetAll(c, redisKey).Result()
	if err != nil {
		app.Logger.Error(fmt.Sprintf("Couldn't get the file from Redis: %v", err))
		return false, false, models.File{}
	}

	if len(redisFile) != 0 {
		tmpSize, _ := strconv.Atoi(redisFile["Size"])
		tmpOriginalSize, _ := strconv.Atoi(redisFile["OriginalSize"])
		ttl, _ := strconv.Atoi(redisFile["TTL"])

		if redisFile["ContentPath"] != "" {
			redisFile["Content"], err = minio.Get(c, app.Config, app.MinIO, redisFile["ContentPath"])
			if err != nil {
				app.Logger.Error(fmt.Sprintf("MinIO Error: %v", err))
				return false, false, models.File{}
			}
		}

		return true, true, models.File{
			Path:         redisFile["Path"],
			Content:      redisFile["Content"],
			Type:         redisFile["Type"],
			Encoding:     encoding,
			Size:         tmpSize,
			OriginalSize: tmpOriginalSize,
			TTL:          ttl,
		}
	}

	body, contentType, err := utils.FetchFile(app.Config, fmt.Sprintf("https://%s/%s", domain.Name, filePath))
	if err != nil {
		app.Logger.Error(fmt.Sprintf("FetchFile Error: %v", err), zap.String("url", fmt.Sprintf("https://%s/%s", domain.Name, filePath)))
		return false, false, models.File{}
	}

	originalSize := len(body)

	// We cannot rely only on the content-length header
	if originalSize > app.Config.FileSizeLimit*1024*1024 {
		return false, false, models.File{}
	}

	var content string
	if acceptsGzipAndIsCompressible {
		content = utils.Compress(&body)
	} else {
		content = string(body)
	}

	newFile := models.File{
		Path:         filePath,
		Type:         contentType,
		Encoding:     encoding,
		Size:         len(content),
		OriginalSize: originalSize,
	}

	if len(content) < app.Config.MemoryStorageLimit*1024*1024 {
		newFile.Content = content
	} else {
		objectName := minio.MakeObjectName(filePath)
		filePath, err := minio.Put(c, app.Config, app.MinIO, objectName, content, contentType)
		if err != nil {
			app.Logger.Error(fmt.Sprintf("MinIO Error: %v", err))
			return false, false, models.File{}
		}

		newFile.ContentPath = filePath
	}

	ttl := 14 * 24 * 3600
	if strings.HasPrefix(filePath, "photos") {
		ttl = 7 * 24 * 3600
	} else if strings.HasPrefix(filePath, "assets") {
		ttl = 3 * 24 * 3600
	} else if strings.HasPrefix(filePath, "fonts") {
		ttl = 30 * 24 * 3600
	}

	app.Redis.HSet(c, redisKey, map[string]interface{}{
		"Path":         newFile.Path,
		"Content":      newFile.Content,
		"ContentPath":  newFile.ContentPath,
		"Type":         newFile.Type,
		"Encoding":     strconv.Itoa(int(newFile.Encoding)),
		"Size":         newFile.Size,
		"OriginalSize": newFile.OriginalSize,
		"TTL":          ttl,
	})

	app.Redis.Expire(c, redisKey, time.Second*time.Duration(ttl))

	if newFile.ContentPath != "" {
		newFile.Content = content
	}

	newFile.TTL = ttl

	return true, false, newFile
}

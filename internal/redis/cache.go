package redis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	mio "github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/config"
	"github.com/muhammadmp97/TinyCDN/internal/minio"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
	"github.com/redis/go-redis/v9"
)

func GetFile(c context.Context, cfg *config.Config, rdb *redis.Client, mio *mio.Client, domain models.Domain, filePath string, headers http.Header) (found bool, hit bool, file models.File) {
	acceptsGzipAndIsCompressible := strings.Contains(headers.Get("Accept-Encoding"), "gzip") && utils.IsCompressible(filePath)
	encoding := models.EncodingNone
	if acceptsGzipAndIsCompressible {
		encoding = models.EncodingGZIP
	}

	redisKey := utils.MakeRedisKey(domain, filePath, acceptsGzipAndIsCompressible)
	redisFile, err := rdb.HGetAll(c, redisKey).Result()
	if err != nil {
		log.Printf("⚠️ Couldn't get the file from Redis: %v", err)
		return false, false, models.File{}
	}

	if len(redisFile) != 0 {
		tmpSize, _ := strconv.Atoi(redisFile["Size"])
		tmpOriginalSize, _ := strconv.Atoi(redisFile["OriginalSize"])

		if redisFile["ContentPath"] != "" {
			redisFile["Content"], err = minio.Get(c, cfg, mio, redisFile["ContentPath"])
			if err != nil {
				log.Printf("⚠️ MinIO Error: %v", err)
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
		}
	}

	ok, body, contentType := utils.FetchFile(fmt.Sprintf("https://%s/%s", domain.Name, filePath))
	if !ok {
		return false, false, models.File{}
	}

	originalSize := len(body)
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

	if len(content) < cfg.MemoryStorageLimit*1024*1024 {
		newFile.Content = content
	} else {
		objectName := minio.MakeObjectName(filePath)
		filePath, err := minio.Put(c, cfg, mio, objectName, content, contentType)
		if err != nil {
			log.Printf("⚠️ MinIO Error: %v", err)
		}

		newFile.ContentPath = filePath
	}

	rdb.HSet(c, redisKey, map[string]interface{}{
		"Path":         newFile.Path,
		"Content":      newFile.Content,
		"ContentPath":  newFile.ContentPath,
		"Type":         newFile.Type,
		"Encoding":     strconv.Itoa(int(newFile.Encoding)),
		"Size":         strconv.Itoa(newFile.Size),
		"OriginalSize": strconv.Itoa(newFile.OriginalSize),
	})

	ttl := 14
	if strings.HasPrefix(filePath, "photos") {
		ttl = 7 * 24 * 3600
	} else if strings.HasPrefix(filePath, "assets") {
		ttl = 3 * 24 * 3600
	} else if strings.HasPrefix(filePath, "fonts") {
		ttl = 30 * 24 * 3600
	}

	rdb.Expire(c, redisKey, time.Second*time.Duration(ttl))

	if newFile.ContentPath != "" {
		newFile.Content = content
	}

	return true, false, newFile
}

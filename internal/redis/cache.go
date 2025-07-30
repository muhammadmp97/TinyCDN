package redis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mio "github.com/minio/minio-go/v7"
	"github.com/muhammadmp97/TinyCDN/internal/minio"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
	"github.com/redis/go-redis/v9"
)

func GetFile(c context.Context, rdb *redis.Client, mio *mio.Client, domain models.Domain, filePath string, headers http.Header) (found bool, hit bool, file models.File) {
	acceptsGzipAndIsCompressible := strings.Contains(headers.Get("Accept-Encoding"), "gzip") && utils.IsCompressible(filePath)
	encoding := models.EncodingNone
	if acceptsGzipAndIsCompressible {
		encoding = models.EncodingGZIP
	}

	redisKey := utils.MakeRedisKey(domain, filePath, acceptsGzipAndIsCompressible)
	redisFile, err := rdb.HGetAll(Ctx, redisKey).Result()
	if err != nil {
		log.Printf("⚠️ Couldn't get the file from Redis: %v", err)
		return false, false, models.File{}
	}

	if len(redisFile) != 0 {
		tmpSize, _ := strconv.Atoi(redisFile["Size"])
		tmpOriginalSize, _ := strconv.Atoi(redisFile["OriginalSize"])

		if redisFile["ContentPath"] != "" {
			redisFile["Content"], err = minio.Get(c, mio, redisFile["ContentPath"])
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

	memoryStorageLimit, _ := strconv.Atoi(os.Getenv("MEMORY_STORAGE_LIMIT"))
	if len(content) < memoryStorageLimit {
		newFile.Content = content
	} else {
		objectName := minio.MakeObjectName(filePath)
		filePath, err := minio.Put(c, mio, objectName, content, contentType)
		if err != nil {
			log.Printf("⚠️ MinIO Error: %v", err)
		}

		newFile.ContentPath = filePath
	}

	rdb.HSet(Ctx, redisKey, map[string]interface{}{
		"Path":         newFile.Path,
		"Content":      newFile.Content,
		"ContentPath":  newFile.ContentPath,
		"Type":         newFile.Type,
		"Encoding":     strconv.Itoa(int(newFile.Encoding)),
		"Size":         strconv.Itoa(newFile.Size),
		"OriginalSize": strconv.Itoa(newFile.OriginalSize),
	})

	ttl, _ := strconv.Atoi(os.Getenv("FILE_CACHE_TTL"))
	rdb.Expire(Ctx, redisKey, time.Second*time.Duration(ttl))

	if newFile.ContentPath != "" {
		newFile.Content = content
	}

	return true, false, newFile
}

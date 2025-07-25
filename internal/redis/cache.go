package redis

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
	"github.com/redis/go-redis/v9"
)

func GetFile(rdb *redis.Client, domain models.Domain, filePath string, headers http.Header) (found bool, hit bool, file models.File) {
	acceptsGzipAndIsCompressible := strings.Contains(headers.Get("Accept-Encoding"), "gzip") && utils.IsCompressible(filePath)
	encoding := models.EncodingNone
	if acceptsGzipAndIsCompressible {
		encoding = models.EncodingGZIP
	}

	redisKey := utils.MakeRedisKey(domain, filePath, acceptsGzipAndIsCompressible)
	redisFile, err := rdb.HGetAll(Ctx, redisKey).Result()
	if err != nil {
		panic(err)
	}

	if len(redisFile) != 0 {
		tmpSize, _ := strconv.Atoi(redisFile["Size"])
		tmpOriginalSize, _ := strconv.Atoi(redisFile["OriginalSize"])
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
		Content:      content,
		Type:         contentType,
		Encoding:     encoding,
		Size:         len(content),
		OriginalSize: originalSize,
	}

	rdb.HSet(Ctx, redisKey, map[string]interface{}{
		"Path":         newFile.Path,
		"Content":      newFile.Content,
		"Type":         newFile.Type,
		"Encoding":     strconv.Itoa(int(newFile.Encoding)),
		"Size":         strconv.Itoa(newFile.Size),
		"OriginalSize": strconv.Itoa(newFile.OriginalSize),
	})

	ttl, _ := strconv.Atoi(os.Getenv("FILE_CACHE_TTL"))
	rdb.Expire(Ctx, redisKey, time.Second*time.Duration(ttl))

	return true, false, newFile
}

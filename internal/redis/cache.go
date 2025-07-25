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
	redisKey := fmt.Sprintf("%s/%s", domain.Name, filePath)
	acceptsGzip := strings.Contains(headers.Get("Accept-Encoding"), "gzip")
	isCompressible := utils.IsCompressible(filePath)
	encoding := models.EncodingNone
	if acceptsGzip && isCompressible {
		redisKey += ":gzip"
		encoding = models.EncodingGZIP
	}

	redisKey = fmt.Sprintf("tcdn:d:%d:f:%s", domain.Id, utils.XXHash(redisKey))
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

	fileUrl := fmt.Sprintf("https://%s/%s", domain.Name, filePath)
	ok, body, contentType := utils.FetchFile(fileUrl)
	if !ok {
		return false, false, models.File{}
	}

	originalSize := len(body)
	var content string
	if acceptsGzip && isCompressible {
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

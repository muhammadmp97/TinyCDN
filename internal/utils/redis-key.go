package utils

import (
	"fmt"

	"github.com/muhammadmp97/TinyCDN/internal/models"
)

func MakeRedisKey(domain models.Domain, filePath string, acceptsGzipAndIsCompressible bool) string {
	redisKey := fmt.Sprintf("%s/%s", domain.Name, filePath)

	if acceptsGzipAndIsCompressible {
		redisKey += ":gzip"
	}

	redisKey = fmt.Sprintf("tcdn:d:%d:f:%s", domain.Id, XXHash(redisKey))

	return redisKey
}

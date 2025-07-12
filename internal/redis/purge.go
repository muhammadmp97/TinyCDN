package redis

import (
	"fmt"

	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
	"github.com/redis/go-redis/v9"
)

func Purge(rdb *redis.Client, domain models.Domain, filePath string) (totalDeleted int64) {
	if filePath != "" { // Purge cache by single-file
		prefix := fmt.Sprintf("tcdn:d:%d:f:", domain.Id)
		redisKey1 := fmt.Sprintf("%s/%s", domain.Name, filePath)
		redisKey2 := fmt.Sprintf("%s/%s:gzip", domain.Name, filePath)
		totalDeleted, err := rdb.Del(Ctx, prefix+utils.XXHash(redisKey1), prefix+utils.XXHash(redisKey2)).Result()

		if err != nil {
			return 0
		}

		return totalDeleted
	} else { // Purge cache by domain
		prefix := fmt.Sprintf("tcdn:d:%d:f:", domain.Id)
		batchSize := 100
		var cursor uint64
		totalDeleted := 0

		for {
			keys, nextCursor, err := rdb.Scan(Ctx, cursor, prefix+"*", int64(batchSize)).Result()
			if err != nil {
				panic(err)
			}

			if len(keys) > 0 {
				if err := rdb.Unlink(Ctx, keys...).Err(); err != nil {
					panic(err)
				} else {
					totalDeleted += len(keys)
				}
			}

			if nextCursor == 0 {
				break
			}

			cursor = nextCursor
		}

		return int64(totalDeleted)
	}
}

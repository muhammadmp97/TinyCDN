package redis

import (
	"context"
	"fmt"

	"github.com/muhammadmp97/TinyCDN/internal/app"
	"github.com/muhammadmp97/TinyCDN/internal/models"
	"github.com/muhammadmp97/TinyCDN/internal/utils"
)

func Purge(c context.Context, app *app.App, domain models.Domain, filePath string) (totalDeleted int64) {
	if filePath != "" { // Purge cache by single-file
		prefix := fmt.Sprintf("tcdn:d:%d:f:", domain.Id)
		redisKey1 := fmt.Sprintf("%s/%s", domain.Name, filePath)
		redisKey2 := fmt.Sprintf("%s/%s:gzip", domain.Name, filePath)
		totalDeleted, err := app.Redis.Del(c, prefix+utils.XXHash(redisKey1), prefix+utils.XXHash(redisKey2)).Result()

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
			keys, nextCursor, err := app.Redis.Scan(c, cursor, prefix+"*", int64(batchSize)).Result()
			if err != nil {
				app.Logger.Error(fmt.Sprintf("rdb.Scan() failed: %v", err))
				return 0
			}

			if len(keys) > 0 {
				if err := app.Redis.Unlink(c, keys...).Err(); err != nil {
					app.Logger.Error(fmt.Sprintf("rdb.Unlink() failed: %v", err))
					return int64(totalDeleted)
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

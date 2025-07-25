package utils

import (
	"testing"

	"github.com/muhammadmp97/TinyCDN/internal/models"
)

func TestMakeRedisKey(t *testing.T) {
	domain := models.Domain{Id: 1, Name: "https://example.com", Token: "xxxyyyzzz"}

	t.Run("makes redis key for a file", func(t *testing.T) {
		key := MakeRedisKey(domain, "style.css", true)
		expectedKey := "tcdn:d:1:f:" + XXHash("https://example.com/style.css:gzip")

		if key != expectedKey {
			t.Error("MakeRedisKey() doesn't work properly!")
		}
	})
}

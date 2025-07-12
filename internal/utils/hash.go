package utils

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
)

func XXHash(str string) string {
	h := xxhash.Sum64String(str)
	return strconv.FormatUint(h, 16)
}

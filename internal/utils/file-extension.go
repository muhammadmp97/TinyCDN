package utils

import (
	"net/url"
	"path"
	"strings"
)

func GetExtension(filePath string) string {
	parsedUrl, err := url.Parse(filePath)
	if err != nil {
		return ""
	}

	return strings.ToLower(path.Ext(parsedUrl.Path))
}

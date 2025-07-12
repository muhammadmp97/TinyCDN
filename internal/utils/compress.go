package utils

import (
	"bytes"
	"compress/gzip"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func IsCompressible(filePath string) bool {
	parsedURL, err := url.Parse(filePath)
	if err != nil {
		return false
	}

	allowedExtensions := []string{
		".html",
		".css",
		".js",
		".json",
		".xml",
		".svg",
		".txt",
	}

	fileExt := strings.ToLower(filepath.Ext(path.Base(parsedURL.Path)))

	for _, allowedExtension := range allowedExtensions {
		if fileExt == allowedExtension {
			return true
		}
	}

	return false
}

func Compress(content *[]byte) string {
	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	w.Write(*content)
	w.Close()
	return buffer.String()
}

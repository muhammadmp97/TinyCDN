package utils

import (
	"bytes"
	"compress/gzip"
)

func IsCompressible(filePath string) bool {
	allowedExtensions := []string{
		".html",
		".css",
		".js",
		".json",
		".xml",
		".svg",
		".txt",
	}

	fileExt := GetExtension(filePath)
	if fileExt == "" {
		return false
	}

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

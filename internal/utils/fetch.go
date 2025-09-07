package utils

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/muhammadmp97/TinyCDN/internal/config"
)

func FetchFile(cfg *config.Config, fileUrl string) (body []byte, contentType string, err error) {
	client := &http.Client{Timeout: 15 * time.Second}

	headResp, err := client.Head(fileUrl)
	if err != nil {
		return nil, "", err
	}

	if contentLength := headResp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.Atoi(contentLength); err == nil && int(size) > cfg.FileSizeLimit*1024*1024 {
			return nil, "", fmt.Errorf("file too large")
		}
	}
	headResp.Body.Close()

	resp, err := client.Get(fileUrl)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 504 {
		return nil, "", err
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType = http.DetectContentType(body)

	return body, contentType, nil
}

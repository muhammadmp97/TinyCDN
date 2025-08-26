package utils

import (
	"io"
	"net/http"
	"time"
)

func FetchFile(fileUrl string) (body []byte, contentType string, err error) {
	client := &http.Client{Timeout: 15 * time.Second}
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

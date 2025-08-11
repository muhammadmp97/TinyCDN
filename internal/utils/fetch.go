package utils

import (
	"io"
	"net/http"
)

func FetchFile(fileUrl string) (ok bool, body []byte, contentType string) {
	resp, err := http.Get(fileUrl)

	if err != nil {
		return false, nil, ""
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 504 {
		return false, nil, ""
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, ""
	}

	contentType = http.DetectContentType(body)

	return true, body, contentType
}

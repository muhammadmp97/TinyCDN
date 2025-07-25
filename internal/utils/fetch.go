package utils

import (
	"io"
	"net/http"
	"strings"
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

	contentType = strings.Split(resp.Header.Get("Content-type"), ";")[0]

	return true, body, contentType
}

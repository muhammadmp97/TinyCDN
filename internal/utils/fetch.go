package utils

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/muhammadmp97/TinyCDN/internal/config"
	errs "github.com/muhammadmp97/TinyCDN/internal/errors"
)

func FetchFile(c context.Context, cfg *config.Config, fileUrl string) (body []byte, contentType string, err error) {
	client := &http.Client{}

	headReq, err := http.NewRequestWithContext(c, http.MethodHead, fileUrl, nil)
	if err != nil {
		return nil, "", err
	}
	headResp, err := client.Do(headReq)
	if err != nil {
		return nil, "", err
	}
	defer headResp.Body.Close()

	if contentLength := headResp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.Atoi(contentLength); err == nil && int(size) > cfg.FileSizeLimit*1024*1024 {
			return nil, "", errs.ErrFileSizeLimit
		}
	}

	getReq, err := http.NewRequestWithContext(c, http.MethodGet, fileUrl, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(getReq)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 504 {
		return nil, "", errs.ErrOriginFileNotFound
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// We cannot rely only on content-length header
	if len(body) > cfg.FileSizeLimit*1024*1024 {
		return nil, "", errs.ErrFileSizeLimit
	}

	contentType = http.DetectContentType(body)

	return body, contentType, nil
}

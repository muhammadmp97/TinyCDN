package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Encoding int8

const (
	EncodingNone Encoding = iota
	EncodingGZIP
)

type File struct {
	Path         string
	Content      string
	Type         string
	Encoding     Encoding
	Size         int
	OriginalSize int
}

type Domain struct {
	Name  string
	Files []File
}

var domains = []Domain{
	{
		Name:  "code.jquery.com",
		Files: []File{},
	},
}

func main() {
	router := gin.Default()

	router.GET("/g/:domain", func(c *gin.Context) {
		domainIndex := getDomainIndex(c.Param("domain"))
		if domainIndex == -1 {
			c.String(404, "Domain not found!")
			return
		}

		fileFound, hit, file := getFile(domainIndex, c.Query("file"), c.Request.Header)
		if !fileFound {
			c.String(404, "File not found!")
			return
		}

		if hit {
			c.Header("Cache-Status", "HIT")
		} else {
			c.Header("Cache-Status", "MISS")
		}

		if file.Encoding == EncodingGZIP {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
		}

		c.Header("Server", "TinyCDN")
		c.Header("Content-Length", strconv.Itoa(file.Size))
		c.Header("Content-Type", file.Type)
		c.String(200, file.Content)
	})

	router.Run()
}

func getDomainIndex(domainName string) int {
	for i, domain := range domains {
		if domain.Name == domainName {
			return i
		}
	}

	return -1
}

func getFile(domainIndex int, filePath string, headers http.Header) (bool, bool, File) {
	acceptsGzip := strings.Contains(headers.Get("Accept-Encoding"), "gzip")
	encoding := EncodingNone
	if acceptsGzip {
		encoding = EncodingGZIP
	}

	for _, file := range domains[domainIndex].Files {
		if file.Path == filePath && file.Encoding == encoding {
			return true, true, file
		}
	}

	fileUrl := fmt.Sprintf("https://%s/%s", domains[domainIndex].Name, filePath)
	resp, err := http.Get(fileUrl)

	if err != nil {
		return false, false, File{}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, false, File{}
	}

	contentType := strings.Split(resp.Header.Get("Content-type"), ";")[0]

	originalSize := len(body)
	var content string
	if acceptsGzip {
		content = compress(&body)
	} else {
		content = string(body)
	}

	newFile := File{
		Path:         filePath,
		Content:      content,
		Type:         contentType,
		Encoding:     encoding,
		Size:         len(content),
		OriginalSize: originalSize,
	}

	domains[domainIndex].Files = append(domains[domainIndex].Files, newFile)

	return true, false, newFile
}

func compress(content *[]byte) string {
	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	w.Write(*content)
	w.Close()
	return buffer.String()
}

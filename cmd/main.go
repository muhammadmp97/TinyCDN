package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type File struct {
	Path    string
	Content string
	Type    string
	Size    int
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

		fileFound, hit, file := getFile(domainIndex, c.Query("file"))
		if !fileFound {
			c.String(404, "File not found!")
			return
		}

		if hit {
			c.Header("Cache-Status", "HIT")
		} else {
			c.Header("Cache-Status", "MISS")
		}

		c.Header("Server", "TinyCDN")
		c.Header("Content-Length", fmt.Sprintf("%d", file.Size))
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

func getFile(domainIndex int, filePath string) (bool, bool, File) {
	for _, file := range domains[domainIndex].Files {
		if file.Path == filePath {
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
	contentLength := len(body)

	newFile := File{
		Path:    filePath,
		Content: string(body),
		Type:    contentType,
		Size:    contentLength,
	}

	domains[domainIndex].Files = append(domains[domainIndex].Files, newFile)

	return true, false, newFile
}

package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type File struct {
	Path    string
	Content string
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

		fileFound, file := getFile(domainIndex, c.Query("file"))
		if !fileFound {
			c.String(404, "File not found!")
			return
		}

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

func getFile(domainIndex int, filePath string) (bool, File) {
	for _, file := range domains[domainIndex].Files {
		if file.Path == filePath {
			return true, file
		}
	}

	fileUrl := fmt.Sprintf("https://%s/%s", domains[domainIndex].Name, filePath)
	resp, err := http.Get(fileUrl)

	if err != nil {
		return false, File{}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, File{}
	}

	newFile := File{Path: filePath, Content: string(body)}
	domains[domainIndex].Files = append(domains[domainIndex].Files, newFile)

	return true, newFile
}

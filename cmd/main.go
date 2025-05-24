package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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
	Name string
}

var domains = []Domain{
	{
		Name: "code.jquery.com",
	},
}

var ctx = context.Background()

func main() {
	router := gin.Default()

	err := godotenv.Load()
	if err != nil {
		panic("No .env file found!")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	router.GET("/g/:domain", func(c *gin.Context) {
		found, domain := getDomain(c.Param("domain"))
		if !found {
			c.String(404, "Domain not found!")
			return
		}

		found, hit, file := getFile(rdb, domain, c.Query("file"), c.Request.Header)
		if !found {
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

	router.POST("/purge/:domain", func(c *gin.Context) {
		found, domain := getDomain(c.Param("domain"))
		if !found {
			c.String(404, "Domain not found!")
			return
		}

		purge(rdb, domain, c.Query("file"))
	})

	router.Run()
}

func getDomain(domainName string) (found bool, domain Domain) {
	for _, domain := range domains {
		if domain.Name == domainName {
			return true, domain
		}
	}

	return false, Domain{}
}

func getFile(rdb *redis.Client, domain Domain, filePath string, headers http.Header) (found bool, hit bool, file File) {
	redisKey := fmt.Sprintf("%s/%s", domain.Name, filePath)
	acceptsGzip := strings.Contains(headers.Get("Accept-Encoding"), "gzip")
	encoding := EncodingNone
	if acceptsGzip {
		redisKey += ":gzip"
		encoding = EncodingGZIP
	}

	redisKey = xxHash(redisKey)
	redisFile, err := rdb.HGetAll(ctx, redisKey).Result()
	if err != nil {
		panic(err)
	}

	if len(redisFile) != 0 {
		tmpSize, _ := strconv.Atoi(redisFile["Size"])
		tmpOriginalSize, _ := strconv.Atoi(redisFile["OriginalSize"])
		return true, true, File{
			Path:         redisFile["Path"],
			Content:      redisFile["Content"],
			Type:         redisFile["Type"],
			Encoding:     encoding,
			Size:         tmpSize,
			OriginalSize: tmpOriginalSize,
		}
	}

	fileUrl := fmt.Sprintf("https://%s/%s", domain.Name, filePath)
	resp, err := http.Get(fileUrl)

	if err != nil {
		return false, false, File{}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 504 {
		return false, false, File{}
	}

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

	rdb.HSet(ctx, redisKey, map[string]interface{}{
		"Path":         newFile.Path,
		"Content":      newFile.Content,
		"Type":         newFile.Type,
		"Encoding":     strconv.Itoa(int(newFile.Encoding)),
		"Size":         strconv.Itoa(newFile.Size),
		"OriginalSize": strconv.Itoa(newFile.OriginalSize),
	})

	ttl, _ := strconv.Atoi(os.Getenv("FILE_CACHE_TTL"))
	rdb.Expire(ctx, redisKey, time.Hour*time.Duration(ttl))

	return true, false, newFile
}

func xxHash(str string) string {
	h := xxhash.Sum64String(str)
	return strconv.FormatUint(h, 16)
}

func compress(content *[]byte) string {
	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	w.Write(*content)
	w.Close()
	return buffer.String()
}

func purge(rdb *redis.Client, domain Domain, filePath string) {
	redisKey1 := fmt.Sprintf("%s/%s", domain.Name, filePath)
	redisKey2 := fmt.Sprintf("%s/%s:gzip", domain.Name, filePath)

	rdb.Del(ctx, xxHash(redisKey1), xxHash(redisKey2))
}

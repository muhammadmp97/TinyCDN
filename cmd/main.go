package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var (
	cacheHit  = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cache_hit_total"}, []string{"domain"})
	cacheMiss = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cache_miss_total"}, []string{"domain"})
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
	Id    int16
	Name  string
	Token string
}

var domains = []Domain{}

var ctx = context.Background()

func main() {
	prometheus.MustRegister(cacheHit)
	prometheus.MustRegister(cacheMiss)

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
	defer rdb.Close()

	loadDomains()

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
			cacheHit.WithLabelValues(domain.Name).Inc()
		} else {
			c.Header("Cache-Status", "MISS")
			cacheMiss.WithLabelValues(domain.Name).Inc()
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
			c.JSON(404, gin.H{"message": "Domain not found!"})
			return
		}

		if domain.Token != c.GetHeader("Authorization") {
			c.JSON(401, gin.H{"message": "Authorization failed!"})
			return
		}

		totalDeleted := purge(rdb, domain, c.Query("file"))

		c.JSON(200, gin.H{"total_deleted": totalDeleted})
	})

	router.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	router.Run()
}

func loadDomains() {
	file, err := os.Open(os.Getenv("DOMAINS_JSON_FILE_PATH"))
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&domains)
	if err != nil {
		panic(err.Error())
	}
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
	isCompressible := isCompressible(filePath)
	encoding := EncodingNone
	if acceptsGzip && isCompressible {
		redisKey += ":gzip"
		encoding = EncodingGZIP
	}

	redisKey = fmt.Sprintf("tcdn:d:%d:f:%s", domain.Id, xxHash(redisKey))
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
	if acceptsGzip && isCompressible {
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
	rdb.Expire(ctx, redisKey, time.Second*time.Duration(ttl))

	return true, false, newFile
}

func xxHash(str string) string {
	h := xxhash.Sum64String(str)
	return strconv.FormatUint(h, 16)
}

func isCompressible(filePath string) bool {
	parsedURL, err := url.Parse(filePath)
	if err != nil {
		return false
	}

	allowedExtensions := []string{
		".html",
		".css",
		".js",
		".json",
		".xml",
		".svg",
		".txt",
	}

	fileExt := strings.ToLower(filepath.Ext(path.Base(parsedURL.Path)))

	for _, allowedExtension := range allowedExtensions {
		if fileExt == allowedExtension {
			return true
		}
	}

	return false
}

func compress(content *[]byte) string {
	var buffer bytes.Buffer
	w := gzip.NewWriter(&buffer)
	w.Write(*content)
	w.Close()
	return buffer.String()
}

func purge(rdb *redis.Client, domain Domain, filePath string) (totalDeleted int64) {
	if filePath != "" { // Purge cache by single-file
		prefix := fmt.Sprintf("tcdn:d:%d:f:", domain.Id)
		redisKey1 := fmt.Sprintf("%s/%s", domain.Name, filePath)
		redisKey2 := fmt.Sprintf("%s/%s:gzip", domain.Name, filePath)
		totalDeleted, err := rdb.Del(ctx, prefix+xxHash(redisKey1), prefix+xxHash(redisKey2)).Result()

		if err != nil {
			return 0
		}

		return totalDeleted
	} else { // Purge cache by domain
		prefix := fmt.Sprintf("tcdn:d:%d:f:", domain.Id)
		batchSize := 100
		var cursor uint64
		totalDeleted := 0

		for {
			keys, nextCursor, err := rdb.Scan(ctx, cursor, prefix+"*", int64(batchSize)).Result()
			if err != nil {
				panic(err)
			}

			if len(keys) > 0 {
				if err := rdb.Unlink(ctx, keys...).Err(); err != nil {
					panic(err)
				} else {
					totalDeleted += len(keys)
				}
			}

			if nextCursor == 0 {
				break
			}

			cursor = nextCursor
		}

		return int64(totalDeleted)
	}
}

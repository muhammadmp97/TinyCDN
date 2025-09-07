package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	RedisAddress  string `mapstructure:"REDIS_ADDRESS"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	MinIOAddress    string `mapstructure:"MINIO_ADDRESS"`
	MinIOUser       string `mapstructure:"MINIO_USER"`
	MinIOPassword   string `mapstructure:"MINIO_PASSWORD"`
	MinIOBucketName string `mapstructure:"MINIO_BUCKET_NAME"`

	FileCacheTTL       int `mapstructure:"FILE_CACHE_TTL"`
	MemoryStorageLimit int `mapstructure:"MEMORY_STORAGE_LIMIT"`
	FileSizeLimit      int `mapstructure:"FILE_SIZE_LIMIT"`

	DomainsJsonFilePath string `mapstructure:"DOMAINS_JSON_FILE_PATH"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.SetDefault("FILE_CACHE_TTL", 3600) // TODO this is not used
	viper.SetDefault("FILE_SIZE_LIMIT", 100)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read .env: %w", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

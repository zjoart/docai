package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv           string
	Port             string
	DBURL            string
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioBucket      string
	OpenRouterAPIKey string
}

func Load() (*Config, error) {
	// Try loading .env from current or parent directories
	pathsToCheck := []string{".env", "../.env", "../../.env"}
	for _, path := range pathsToCheck {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			break
		}
	}

	return &Config{
		AppEnv:           getEnv("APP_ENV"),
		Port:             getEnv("PORT"),
		DBURL:            getEnv("DATABASE_URL"),
		MinioEndpoint:    getEnv("MINIO_ENDPOINT"),
		MinioAccessKey:   getEnv("MINIO_ACCESS_KEY"),
		MinioSecretKey:   getEnv("MINIO_SECRET_KEY"),
		MinioBucket:      getEnv("MINIO_BUCKET"),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY"),
	}, nil
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("%s is required", key))
}

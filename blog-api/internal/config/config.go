package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	JWTSecret    string
	TokenExpiry  time.Duration
	RefreshExpiry time.Duration
	DataDir      string
	PostsPerPage int
	Environment  string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET", "blog-super-secret-key-2024"),
		TokenExpiry:   time.Hour * 24,
		RefreshExpiry: time.Hour * 24 * 7,
		DataDir:       getEnv("DATA_DIR", "./data"),
		PostsPerPage:  getEnvAsInt("POSTS_PER_PAGE", 10),
		Environment:   getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

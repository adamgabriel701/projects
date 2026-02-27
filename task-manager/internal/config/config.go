package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	JWTSecret   string
	TokenExpiry time.Duration
	DataDir     string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		TokenExpiry: 24 * time.Hour,
		DataDir:     getEnv("DATA_DIR", "./data"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

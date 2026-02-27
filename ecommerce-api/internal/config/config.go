package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	JWTSecret      string
	TokenExpiry    time.Duration
	DataDir        string
	AdminEmail     string
	AdminPassword  string
	PaymentGateway string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "ecommerce-secret-key-2024"),
		TokenExpiry:    time.Hour * 24,
		DataDir:        getEnv("DATA_DIR", "./data"),
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@ecommerce.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin123"),
		PaymentGateway: getEnv("PAYMENT_GATEWAY", "mock"),
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

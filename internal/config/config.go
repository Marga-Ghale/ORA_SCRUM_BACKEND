// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	Environment   string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	JWTExpiry     int
	RefreshExpiry int
}

func Load() *Config {
	return &Config{
		Port:          getEnv("API_PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/ora_scrum?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiry:     getEnvInt("JWT_EXPIRY", 24),
		RefreshExpiry: getEnvInt("REFRESH_EXPIRY", 7),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

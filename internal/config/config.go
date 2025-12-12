package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret      string
	JWTExpiry      int // hours
	RefreshExpiry  int // days

	// Email (for notifications)
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("API_PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/ora_scrum?schema=public"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiry:     getEnvInt("JWT_EXPIRY", 24),
		RefreshExpiry: getEnvInt("REFRESH_EXPIRY", 7),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnvInt("SMTP_PORT", 587),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:      getEnv("SMTP_FROM", "noreply@ora-scrum.com"),
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

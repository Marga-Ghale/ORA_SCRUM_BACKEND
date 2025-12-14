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

	// Email configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string
	SMTPUseTLS   bool

	// Frontend URL for email links
	FrontendURL string
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

		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@ora-scrum.com"),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "ORA Scrum"),
		SMTPUseTLS:   getEnvBool("SMTP_USE_TLS", false),

		// Frontend URL for email links
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
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

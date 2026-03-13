package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CORSOrigins     []string
	CORSMethods     []string
	CORSHeaders     []string
	CORSCredentials bool
	CORSMaxAge      int
}

func Load() *Config {
	_ = godotenv.Load(".env")
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
		RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "168h")),
		CORSOrigins:     parseCSV(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		CORSMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		CORSCredentials: true,
		CORSMaxAge:      300,
	}
}

func getEnv(key string, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultVal == "" {
		panic("Missing required env var: " + key)
	}
	return defaultVal
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic("invalid duration: " + s)
	}
	return d
}

func parseCSV(s string) []string {
	var result []string
	for _, v := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

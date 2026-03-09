package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DBPath          string
	JWTSecret       string
	AccessTokenTTL  string
	RefreshTokenTTL string
}

func Load() *Config {
	_ = godotenv.Load(".env")
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  getEnv("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL: getEnv("REFRESH_TOKEN_TTL", "168h"),
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

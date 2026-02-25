package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Port        string
}

func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://biblios_user:biblios_pass@postgres:5432/biblios"),
		RedisURL:    getEnv("REDIS_URL", "redis://redis:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "changeme"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

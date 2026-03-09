package config

import (
	"os"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	Port              string
	GoogleBooksAPIKey string
	CoversDir         string // filesystem path where uploaded cover images are stored

}

func Load() Config {
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://biblios_user:biblios_pass@postgres:5432/biblios"),
		RedisURL:          getEnv("REDIS_URL", "redis://redis:6379"),
		JWTSecret:         getEnv("JWT_SECRET", "changeme"),
		Port:              getEnv("PORT", "8080"),
		GoogleBooksAPIKey: getEnv("GOOGLE_BOOKS_API_KEY", ""),
		CoversDir:         getEnv("COVERS_DIR", "./data/covers"),
	}
}

func getEnv(key string, fallback string) string {
	var val string = os.Getenv(key)
	if val != "" {
		return val
	}
	return fallback
}

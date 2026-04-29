package config

import "os"

type Config struct {
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	Port              string
	GoogleBooksAPIKey string
	CoversDir         string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
}

func Load() Config {
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://biblios_user:biblios_pass@postgres:5432/biblios"),
		RedisURL:          getEnv("REDIS_URL", "redis://redis:6379"),
		JWTSecret:         getEnv("JWT_SECRET", "changeme"),
		Port:              getEnv("PORT", "8080"),
		GoogleBooksAPIKey: getEnv("GOOGLE_BOOKS_API_KEY", ""),
		CoversDir:         getEnv("COVERS_DIR", "./data/covers"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@biblioslibrary.app"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

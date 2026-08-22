package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentTest        Environment = "test"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
)

const (
	minimumJWTSecretLength   = 32
	minimumDatabasePassword  = 16
	minimumRedisSecretLength = 16
)

type Config struct {
	Environment       Environment
	DatabaseURL       string
	RedisURL          string
	RedisPassword     string
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

func Load() (Config, error) {
	cfg := Config{
		Environment:       Environment(strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:          strings.TrimSpace(os.Getenv("REDIS_URL")),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		Port:              getEnv("PORT", "8080"),
		GoogleBooksAPIKey: getEnv("GOOGLE_BOOKS_API_KEY", ""),
		CoversDir:         getEnv("COVERS_DIR", "./data/covers"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@biblioslibrary.app"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var validationErrors []error

	switch c.Environment {
	case EnvironmentDevelopment, EnvironmentTest, EnvironmentStaging, EnvironmentProduction:
	case "":
		validationErrors = append(validationErrors, errors.New("APP_ENV is required"))
	default:
		validationErrors = append(validationErrors, errors.New("APP_ENV must be one of development, test, staging, or production"))
	}

	if err := validateDatabaseURL(c.DatabaseURL); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateRedis(c.RedisURL, c.RedisPassword); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateSecret("JWT_SECRET", c.JWTSecret, minimumJWTSecretLength); err != nil {
		validationErrors = append(validationErrors, err)
	}

	return errors.Join(validationErrors...)
}

func validateDatabaseURL(raw string) error {
	if raw == "" {
		return errors.New("DATABASE_URL is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("DATABASE_URL must use the postgres or postgresql scheme")
	}
	if parsed.User == nil {
		return errors.New("DATABASE_URL must include a database user and password")
	}
	password, present := parsed.User.Password()
	if parsed.User.Username() == "" || !present || password == "" {
		return errors.New("DATABASE_URL must include a database user and password")
	}
	if err := validateSecret("DATABASE_URL password", password, minimumDatabasePassword); err != nil {
		return err
	}
	return nil
}

func validateRedis(rawURL, password string) error {
	passwordError := validateSecret("REDIS_PASSWORD", password, minimumRedisSecretLength)
	if rawURL == "" {
		return errors.Join(errors.New("REDIS_URL is required"), passwordError)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("REDIS_URL must be a valid Redis URL")
	}
	if parsed.Scheme != "redis" && parsed.Scheme != "rediss" {
		return errors.New("REDIS_URL must use the redis or rediss scheme")
	}
	if parsed.User != nil {
		return errors.New("REDIS_URL must not embed credentials; use REDIS_PASSWORD")
	}
	return passwordError
}

func validateSecret(name, value string, minimumLength int) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if len(value) < minimumLength {
		return fmt.Errorf("%s must be at least %d characters", name, minimumLength)
	}
	if isInsecureSecret(value) {
		return fmt.Errorf("%s contains a known default or placeholder value", name)
	}
	return nil
}

func isInsecureSecret(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "biblios_pass", "changeme", "change-me", "password", "postgres", "redis":
		return true
	}
	return strings.Contains(normalized, "replace-with-") ||
		strings.Contains(normalized, "example-only") ||
		strings.Contains(normalized, "placeholder")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

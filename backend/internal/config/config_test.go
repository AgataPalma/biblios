package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{
		Environment:   EnvironmentDevelopment,
		DatabaseURL:   "postgres://biblios:correct-horse-battery-staple@postgres:5432/biblios",
		RedisURL:      "redis://redis:6379",
		RedisPassword: "local-redis-secret-123456789",
		JWTSecret:     "local-jwt-secret-with-at-least-32-characters",
	}
}

func TestConfigValidateAcceptsExplicitSecureConfiguration(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid configuration, got %v", err)
	}
}

func TestConfigValidateRejectsMissingEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = ""
	assertValidationError(t, cfg, "APP_ENV is required")
}

func TestConfigValidateRejectsUnknownEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = "local"
	assertValidationError(t, cfg, "APP_ENV must be one of")
}

func TestConfigValidateRejectsKnownDefaultJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWTSecret = "changeme"
	assertValidationError(t, cfg, "JWT_SECRET")
}

func TestConfigValidateRejectsShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWTSecret = "short"
	assertValidationError(t, cfg, "at least 32 characters")
}

func TestConfigValidateRejectsKnownDatabasePassword(t *testing.T) {
	cfg := validConfig()
	cfg.DatabaseURL = "postgres://biblios:biblios_pass@postgres:5432/biblios"
	assertValidationError(t, cfg, "DATABASE_URL password")
}

func TestConfigValidateRejectsShortDatabasePassword(t *testing.T) {
	cfg := validConfig()
	cfg.DatabaseURL = "postgres://biblios:too-short@postgres:5432/biblios"
	assertValidationError(t, cfg, "at least 16 characters")
}

func TestConfigValidateRequiresDatabaseCredentials(t *testing.T) {
	cfg := validConfig()
	cfg.DatabaseURL = "postgres://postgres:5432/biblios"
	assertValidationError(t, cfg, "database user and password")
}

func TestConfigValidateRejectsEmbeddedRedisCredentials(t *testing.T) {
	cfg := validConfig()
	cfg.RedisURL = "redis://:embedded-secret@redis:6379"
	assertValidationError(t, cfg, "must not embed credentials")
}

func TestConfigValidateRequiresRedisPassword(t *testing.T) {
	cfg := validConfig()
	cfg.RedisPassword = ""
	assertValidationError(t, cfg, "REDIS_PASSWORD is required")
}

func TestConfigValidateRejectsPlaceholderSecrets(t *testing.T) {
	cfg := validConfig()
	cfg.RedisPassword = "replace-with-random-secret"
	assertValidationError(t, cfg, "known default or placeholder value")
}

func TestLoadDoesNotFallbackToSecuritySensitiveDefaults(t *testing.T) {
	for _, key := range []string{"APP_ENV", "DATABASE_URL", "REDIS_URL", "REDIS_PASSWORD", "JWT_SECRET"} {
		t.Setenv(key, "")
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing security-sensitive configuration to fail")
	}
	for _, variable := range []string{"APP_ENV", "DATABASE_URL", "REDIS_URL", "REDIS_PASSWORD", "JWT_SECRET"} {
		if !strings.Contains(err.Error(), variable) {
			t.Errorf("expected error to identify %s without exposing a value; got %v", variable, err)
		}
	}
}

func assertValidationError(t *testing.T, cfg Config, expected string) {
	t.Helper()
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected validation error containing %q", expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected validation error containing %q, got %v", expected, err)
	}
}

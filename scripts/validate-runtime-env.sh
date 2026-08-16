#!/bin/sh

set -eu

mode="${1:-core}"

fail() {
  printf '%s\n' "runtime configuration rejected: $1" >&2
  exit 1
}

require_value() {
  name="$1"
  case "$name" in
    APP_ENV) value="${APP_ENV:-}" ;;
    POSTGRES_PASSWORD) value="${POSTGRES_PASSWORD:-}" ;;
    REDIS_PASSWORD) value="${REDIS_PASSWORD:-}" ;;
    JWT_SECRET) value="${JWT_SECRET:-}" ;;
    PGADMIN_EMAIL) value="${PGADMIN_EMAIL:-}" ;;
    PGADMIN_PASSWORD) value="${PGADMIN_PASSWORD:-}" ;;
    *) fail "unknown variable name" ;;
  esac
  [ -n "$value" ] || fail "$name is required"
}

is_insecure_secret() {
  normalized="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  case "$normalized" in
    biblios_pass|changeme|change-me|password|postgres|redis|*replace-with-*|*example-only*|*placeholder*)
      return 0
      ;;
  esac
  return 1
}

validate_secret() {
  name="$1"
  minimum_length="$2"
  require_value "$name"
  [ "${#value}" -ge "$minimum_length" ] || fail "$name must be at least $minimum_length characters"
  if is_insecure_secret "$value"; then
    fail "$name contains a known default or placeholder value"
  fi
}

case "$mode" in
  core)
    require_value APP_ENV
    case "$APP_ENV" in
      development|test|staging|production) ;;
      *) fail "APP_ENV must be one of development, test, staging, or production" ;;
    esac
    validate_secret POSTGRES_PASSWORD 16
    validate_secret REDIS_PASSWORD 16
    validate_secret JWT_SECRET 32
    ;;
  pgadmin)
    require_value PGADMIN_EMAIL
    validate_secret PGADMIN_PASSWORD 16
    ;;
  *)
    fail "unknown validation mode"
    ;;
esac

printf '%s\n' "runtime configuration accepted ($mode)"

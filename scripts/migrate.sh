#!/bin/bash
# scripts/migrate.sh - CI/CD Database Migration Runner (Goose)
# Usage:
#   ./scripts/migrate.sh up          - Apply all pending migrations
#   ./scripts/migrate.sh down        - Rollback last migration
#   ./scripts/migrate.sh status      - Show current migration status
#   ./scripts/migrate.sh version     - Show current version
#   ./scripts/migrate.sh force <V>   - Force set version (emergency)

set -e

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-sharah_db}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSLMODE="${DB_SSLMODE:-disable}"
GOOSE_TIMEOUT="${GOOSE_TIMEOUT:-2m}"

MIGRATIONS_DIR="${MIGRATIONS_DIR:-$(dirname "$0")/../db/main/migrations}"
ACTION="${1:-up}"
DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

if ! command -v goose &> /dev/null; then
    echo "Installing goose CLI..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
fi

echo "Running migration: ${ACTION}"
echo "Database: ${DB_NAME} at ${DB_HOST}:${DB_PORT}"
echo "Migrations: ${MIGRATIONS_DIR}"
echo "---"

cd "$MIGRATIONS_DIR"

case "$ACTION" in
    up)
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" up
        echo "All migrations applied successfully."
        ;;
    down)
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" down
        echo "Rolled back last migration."
        ;;
    status)
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" status
        ;;
    version)
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" version
        ;;
    force)
        VERSION="$2"
        if [ -z "$VERSION" ]; then
            echo "Error: force requires a version number"
            exit 1
        fi
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" up-to "$VERSION"
        echo "Forced version to $VERSION"
        ;;
    reset)
        goose -timeout "$GOOSE_TIMEOUT" postgres "$DBSTRING" reset
        echo "All migrations rolled back."
        ;;
    *)
        echo "Unknown action: $ACTION"
        echo "Usage: $0 {up|down|status|version|force|reset}"
        exit 1
        ;;
esac

# Makefile for Shahrah Database Migrations
# Uses pressly/goose CLI for production-ready migrations

# Configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= sharah_db
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_SSLMODE ?= disable
MIGRATIONS_DIR := db/main/migrations
DBSTRING := "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

# Install goose CLI
.PHONY: migrate-install
migrate-install:
	@goose -version > /dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest

# Apply all pending migrations
.PHONY: migrate-up
migrate-up: migrate-install
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) up

# Rollback last migration
.PHONY: migrate-down
migrate-down: migrate-install
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) down

# Rollback all migrations
.PHONY: migrate-reset
migrate-reset: migrate-install
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) reset

# Show current migration version and status
.PHONY: migrate-status
migrate-status: migrate-install
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) status

# Create a new migration file
# Usage: make migrate-create name=my_new_feature
.PHONY: migrate-create
migrate-create: migrate-install
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=my_new_feature"; exit 1; fi
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) create $(name) sql

# Docker-based migration (run migrations in docker-compose environment)
.PHONY: docker-migrate-up
docker-migrate-up:
	docker compose -f infra/compose/local.yml run --rm migrate up

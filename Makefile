.PHONY: help setup setup-full setup-local build run test-unit test-modules test-contracts test-tunnel-server test-tunnel-client test-silk-road test-integration test-integration-db test fmt vet lint check docker-up docker-up-mailpit docker-up-pgadmin docker-down docker-migrate docker-logs docker-clean \
	migrate-install migrate-up migrate-down migrate-reset migrate-status migrate-create docker-migrate-up docker-model-discovery-migrate

CONTROL_PLANE_DIR := services/control-plane
COMPOSE_FILE := infra/compose/local.yml
POSTGRES_VERSION ?= 18
export POSTGRES_VERSION
DOCKER_COMPOSE := docker compose -f $(COMPOSE_FILE)
GO := go
GOLANGCI_LINT := golangci-lint

ifeq ($(OS),Windows_NT)
GO_TEST_FLAGS := -shuffle=on
else
GO_TEST_FLAGS := -race -shuffle=on
endif

# Migration configuration (goose)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= sharah_db
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_SSLMODE ?= disable
MIGRATIONS_DIR := db/main/migrations
DBSTRING := "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "Write-Output 'Usage: make [target]'; Write-Output ''; Write-Output 'Targets:'; $$targets = Get-Content '$(firstword $(MAKEFILE_LIST))' | Where-Object { $$_ -match '^[a-zA-Z_-]+:.*?## .*' } | ForEach-Object { if ($$_ -match '^([a-zA-Z_-]+):.*?## (.*)$$') { [PSCustomObject]@{ Name = $$matches[1]; Desc = $$matches[2] } } } | Sort-Object Name; foreach ($$t in $$targets) { '{0,-20} {1}' -f $$t.Name, $$t.Desc }"
else
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
endif

setup: setup-local ## Default local setup without remote-sensitive tidy

setup-local: ## Download control-plane dependencies (local-safe, no tidy)
	$(GO) -C $(CONTROL_PLANE_DIR) mod download
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration mod download

setup-full: ## Download and tidy control-plane dependencies (may require VCS access)
	$(GO) -C $(CONTROL_PLANE_DIR) mod tidy
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration mod tidy

build: ## Build the control-plane service
	$(GO) -C $(CONTROL_PLANE_DIR) build -o ../../bin/control-plane ./cmd/server

run: ## Run the control-plane service locally
	$(GO) -C $(CONTROL_PLANE_DIR) run ./cmd/server

test-unit: ## Run all control-plane unit and package tests
	$(GO) -C $(CONTROL_PLANE_DIR) test -v $(GO_TEST_FLAGS) ./cmd/... ./internal/... ./pkg/... ./tests/unit/...

test-contracts: ## Run contracts module tests
	$(GO) -C packages/contracts test -v $(GO_TEST_FLAGS) ./...

test-tunnel-server: ## Run tunnel-server module tests
	$(GO) -C services/tunnel-server test -v $(GO_TEST_FLAGS) ./...

test-tunnel-client: ## Run tunnel-client module tests
	$(GO) -C services/tunnel-client test -v $(GO_TEST_FLAGS) ./...

test-silk-road: ## Run silk-road module tests
	$(GO) -C services/silk-road test -v $(GO_TEST_FLAGS) ./...

test-modules: test-contracts test-tunnel-server test-tunnel-client test-silk-road ## Run non-control-plane Go module tests

test-integration: ## Run control-plane integration tests
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration test -v $(GO_TEST_FLAGS) ./...

test-integration-db: ## Run Postgres-backed control-plane integration tests (requires DATABASE_URL)
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration test -v $(GO_TEST_FLAGS) ./... -run "PostgresBacked|StorageSchemaContract|ConcurrentWalletDebitUsesRowLock"

test: test-unit test-modules ## Run all non-infrastructure Go tests

fmt: ## Format all Go modules
	$(GO) -C packages/contracts fmt ./...
	$(GO) -C services/tunnel-server fmt ./...
	$(GO) -C services/tunnel-client fmt ./...
	$(GO) -C services/silk-road fmt ./...
	$(GO) -C $(CONTROL_PLANE_DIR) fmt ./...
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration fmt ./...

vet: ## Run go vet for all Go modules
	$(GO) -C packages/contracts vet ./...
	$(GO) -C services/tunnel-server vet ./...
	$(GO) -C services/tunnel-client vet ./...
	$(GO) -C services/silk-road vet ./...
	$(GO) -C $(CONTROL_PLANE_DIR) vet ./...
	$(GO) -C $(CONTROL_PLANE_DIR)/tests/integration vet ./...

lint: ## Run golangci-lint for control-plane
	cd $(CONTROL_PLANE_DIR) && $(GOLANGCI_LINT) run ./...
	cd $(CONTROL_PLANE_DIR)/tests/integration && $(GOLANGCI_LINT) run ./...

check: fmt vet lint test ## Run the default checks

docker-up: docker-migrate-up docker-model-discovery-migrate ## Start the local backend stack (including Mailpit and pgAdmin)
	$(DOCKER_COMPOSE) up -d postgres redis clickhouse mailpit pgadmin silk-road control-plane model-discovery support-admin


docker-up-mailpit: ## Start only Mailpit for local email testing
	$(DOCKER_COMPOSE) up -d mailpit


docker-up-pgadmin: ## Start only pgAdmin for Postgres administration
	$(DOCKER_COMPOSE) up -d pgadmin
docker-down: ## Stop local dependencies
	$(DOCKER_COMPOSE) down

docker-migrate: docker-migrate-up ## Run the main database migrations

docker-logs: ## Follow local compose logs
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Remove local compose resources
	$(DOCKER_COMPOSE) down -v --remove-orphans

docker-model-discovery-migrate: ## Run Prisma migrations for model-discovery service
	$(DOCKER_COMPOSE) run --rm model-discovery-migrate

# Install goose CLI
migrate-install: ## Install goose CLI if missing
	@goose -version > NUL 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest

# Apply all pending migrations
migrate-up: migrate-install ## Apply all pending goose migrations
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) up

# Rollback last migration
migrate-down: migrate-install ## Roll back last goose migration
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) down

# Rollback all migrations
migrate-reset: migrate-install ## Roll back all goose migrations
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) reset

# Show current migration version and status
migrate-status: migrate-install ## Show goose migration status
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) status

# Create a new migration file
# Usage: make migrate-create name=my_new_feature
migrate-create: migrate-install ## Create a new goose SQL migration (name=...)
	@$(if $(name),,$(error Usage: make migrate-create name=my_new_feature))
	@cd $(MIGRATIONS_DIR) && goose postgres $(DBSTRING) create $(name) sql

# Docker-based migration (run migrations in docker-compose environment)
docker-migrate-up: ## Run migrations in docker compose migrate service
	docker compose -f infra/compose/local.yml run --rm migrate

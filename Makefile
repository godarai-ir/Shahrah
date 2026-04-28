.PHONY: help setup setup-full setup-local build run test-unit test-integration test-integration-db test fmt vet lint check docker-up docker-down docker-migrate docker-logs docker-clean

CONTROL_PLANE_DIR := services/control-plane
COMPOSE_FILE := infra/compose/local.yml
POSTGRES_VERSION ?= 18
export POSTGRES_VERSION
DOCKER_COMPOSE := docker compose -f $(COMPOSE_FILE)
GO := go
GOLANGCI_LINT := golangci-lint

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: setup-local ## Default local setup without remote-sensitive tidy

setup-local: ## Download control-plane dependencies (local-safe, no tidy)
	$(GO) -C $(CONTROL_PLANE_DIR) mod download

setup-full: ## Download and tidy control-plane dependencies (may require VCS access)
	$(GO) -C $(CONTROL_PLANE_DIR) mod tidy

build: ## Build the control-plane service
	$(GO) -C $(CONTROL_PLANE_DIR) build -o ../../bin/control-plane ./cmd/server

run: ## Run the control-plane service locally
	$(GO) -C $(CONTROL_PLANE_DIR) run ./cmd/server

test-unit: ## Run control-plane unit tests
	$(GO) -C $(CONTROL_PLANE_DIR) test -v -race ./tests/unit/...

test-integration: ## Run control-plane integration tests
	$(GO) -C $(CONTROL_PLANE_DIR) test -v -race ./tests/integration/...

test-integration-db: ## Run Postgres-backed control-plane integration tests (requires DATABASE_URL)
	$(GO) -C $(CONTROL_PLANE_DIR) test -v -race ./tests/integration/... -run "PostgresBacked|StorageSchemaContract|ConcurrentWalletDebitUsesRowLock"

test: test-unit ## Run default tests

fmt: ## Format control-plane Go code
	$(GO) -C $(CONTROL_PLANE_DIR) fmt ./...

vet: ## Run go vet for control-plane
	$(GO) -C $(CONTROL_PLANE_DIR) vet ./...

lint: ## Run golangci-lint for control-plane
	cd $(CONTROL_PLANE_DIR) && $(GOLANGCI_LINT) run ./...

check: fmt vet lint test ## Run the default checks

docker-up: ## Start the local backend stack
	$(DOCKER_COMPOSE) up -d postgres redis silk-road tunnel-server tunnel-client control-plane

docker-down: ## Stop local dependencies
	$(DOCKER_COMPOSE) down

docker-migrate: ## Run the main database bootstrap migrations
	$(DOCKER_COMPOSE) exec postgres sh -lc "psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql && psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/002_platform_foundations.sql"

docker-logs: ## Follow local compose logs
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Remove local compose resources
	$(DOCKER_COMPOSE) down -v --remove-orphans

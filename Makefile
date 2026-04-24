# Makefile for go-llm-proxy
# Usage: make [target]

.PHONY: all setup build run test-unit test-integration test lint fmt clean vet docker-up docker-down docker-migrate help

# =============================================================================
# Variables
# =============================================================================

# Build variables
BINARY_NAME     := go-llm-proxy
BUILD_DIR       := bin
VERSION        ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME      := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO              := go
GOFMT           := gofmt

# Build flags
LDFLAGS         := -ldflags "-s -w \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GitCommit=$(GIT_COMMIT)"

# Test variables
COVERAGE_DIR    := coverage
TEST_ARGS       := -v -race
TEST_PACKAGES   := ./...

# Docker variables
DOCKER_COMPOSE  := docker-compose

# Lint variables
GOLANGCI_LINT   := golangci-lint

# OS detection
ifeq ($(OS),Windows_NT)
	PLATFORM := windows
	ARCH     := amd64
else
	UNAME_S := $(shell uname -s)
	UNAME_M := $(shell uname -m)
	PLATFORM := $(shell echo $(UNAME_S) | tr '[:upper:]' '[:lower:]')
	ARCH     := $(UNAME_M)
endif

# =============================================================================
# Default target
# =============================================================================

all: test build ## Run tests and build binary

# =============================================================================
# Help
# =============================================================================

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Build
# =============================================================================

setup: ## Download and tidy Go module dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/

run: ## Run the server locally
	$(GO) run $(LDFLAGS) ./cmd/server/

build-linux-amd64: ## Build binary for Linux amd64
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server/

build-darwin-amd64: ## Build binary for macOS amd64
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server/

build-darwin-arm64: ## Build binary for macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server/

build-windows-amd64: ## Build binary for Windows amd64
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server/

build-all: build-linux-amd64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 ## Build binaries for all platforms

# =============================================================================
# Test
# =============================================================================

test-unit: ## Run unit tests
	$(GO) test $(TEST_ARGS) -coverprofile=$(COVERAGE_DIR)/unit.coverage ./tests/unit/...

test-integration: ## Run integration tests (requires Docker)
	$(GO) test $(TEST_ARGS) -coverprofile=$(COVERAGE_DIR)/integration.coverage ./tests/integration/...

test: test-unit ## Run all tests

test-coverage: ## Run tests and generate HTML coverage report
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test $(TEST_ARGS) -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

test-bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem ./...

# =============================================================================
# Code Quality
# =============================================================================

lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

fmt: ## Format Go source files
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GO) vet ./...

check: fmt vet lint test ## Run all code quality checks

# =============================================================================
# Database / Docker
# =============================================================================

docker-up: ## Start development environment (PostgreSQL + Redis)
	$(DOCKER_COMPOSE) up -d postgres redis

docker-down: ## Stop development environment
	$(DOCKER_COMPOSE) down

docker-migrate: ## Run database migrations
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Clean Docker resources
	$(DOCKER_COMPOSE) down -v --remove-orphans

# =============================================================================
# Cleanup
# =============================================================================

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	$(GO) clean -cache -testcache -modcache

# =============================================================================
# Release
# =============================================================================

release: clean test build-all ## Build all release binaries
	@echo "Release $(VERSION) built successfully in $(BUILD_DIR)/"

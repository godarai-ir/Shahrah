# Contributing to Go LLM Proxy

Thank you for your interest in contributing to Go LLM Proxy! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows the standard open source code of conduct. Be respectful and inclusive in all interactions.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (or use PowerShell commands directly on Windows)

### Setup

```bash
# Clone the repository
git clone https://github.com/BerriAI/go-llm-proxy.git
cd go-llm-proxy

# Install dependencies
make setup

# Start development environment
make docker-up

# Run migrations
make docker-migrate
```

## Development Workflow

### Running Tests

```bash
# Run unit tests
make test-unit

# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make test-bench
```

### Code Style

- Follow idiomatic Go conventions
- Use `gofmt` or `gofumpt` for formatting
- Run `make fmt` to format code
- Run `make lint` before submitting PRs

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet
```

### Building

```bash
# Build binary
make build

# Run server
make run

# Build all platforms
make build-all
```

## Pull Request Guidelines

1. **Branch naming**: Use descriptive branch names (e.g., `feature/add-anthropic-provider`, `fix/router-cooldown-bug`)

2. **Commit messages**: Follow conventional commit format:
   ```
   type(scope): description

   [optional body]

   [optional footer]
   ```
   Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `ci`, `chore`

3. **Tests**: All new features must include tests. Bug fixes should include a regression test.

4. **Documentation**: Update README.md and relevant documentation when adding new features.

5. **Go module**: Run `go mod tidy` after adding new dependencies.

## Adding a New Provider

1. Create a new directory under `internal/providers/<provider_name>/`
2. Implement the `providers.Provider` interface:
   - `Name() string`
   - `Complete(ctx context.Context, req types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)`
   - `StreamComplete(ctx context.Context, req types.ChatCompletionRequest) (<-chan types.StreamChunk, <-chan error, error)`
3. Register the provider in `internal/proxy/handler.go`'s `createProvider` function
4. Add unit tests for the provider

## Adding a New Routing Strategy

1. Create a new file under `internal/router/` (e.g., `least_busy.go`)
2. Implement the `RoutingStrategy` interface:
   - `Name() string`
   - `SelectDeployment(deployments []*Deployment, model string) (*Deployment, error)`
3. Register the strategy in `internal/router/router.go`'s `NewRouter` function

## Adding Tests

### Unit Tests

- Place tests in `tests/unit/` mirroring the `internal/` structure
- Use table-driven tests where appropriate
- Test file naming: `<component>_test.go`

Example table-driven test:

```go
func TestHashKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
	}{
		{
			name:    "valid API key produces 64 char hash",
			input:   "sk-test-key",
			wantLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashKey(tt.input)
			assert.Len(t, hash, tt.wantLen)
		})
	}
}
```

### Integration Tests

- Place integration tests in `tests/integration/`
- Use Docker Compose for external services
- Tests should be self-contained and clean up after themselves

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or error messages

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

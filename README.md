# Shahrah

A high-performance, horizontally scalable LLM proxy written in Go. This proxy provides a unified OpenAI-compatible interface for 100+ LLM providers with intelligent routing, load balancing, and API key management.

## Features

- **Unified OpenAI Interface**: All providers return responses in OpenAI format
- **Intelligent Routing**: Multiple routing strategies (simple-shuffle, least-busy, lowest-cost, lowest-latency, usage-based)
- **Load Balancing**: Distribute requests across multiple deployments of the same model
- **Automatic Failover**: Fallback to secondary deployments on failure
- **API Key Authentication**: Bearer token, `api-key`, and `x-api-key` header support
- **Rate Limiting**: Per-key RPM/TPM limits (Redis-backed for multi-instance)
- **Budget Tracking**: Per-key and per-team budget enforcement
- **Health Checks**: Automatic detection and cooldown of failed deployments
- **Streaming Support**: SSE streaming for real-time responses

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (optional, for development)
- PostgreSQL 15+ (optional, for production)
- Redis 7+ (optional, for distributed caching)

### Setup

#### Linux/macOS (with Make)

```bash
# Clone the repository
git clone https://github.com/Delpak/shahrah.git
cd shahrah

# Install dependencies
make setup

# Start development environment (PostgreSQL + Redis)
make docker-up

# Run database migrations
make docker-migrate

# Run unit tests
make test-unit

# Start the server
make run
```

#### Windows (PowerShell - no Make required)

```powershell
cd shahrah

# Install dependencies
go mod download
go mod tidy

# Start development environment (PostgreSQL + Redis)
docker-compose up -d postgres redis

# Wait for PostgreSQL to be ready, then run migrations
docker-compose exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql

# Run unit tests
go test -v -race ./tests/unit/...

# Start the server
go run ./cmd/server/
```

### Using Docker Compose

```bash
# Build and start all services
docker-compose up --build

# The proxy will be available at http://localhost:4000
```

### Configuration

Copy and edit the configuration file:

```bash
cp configs/proxy_config.yaml configs/proxy_config.local.yaml
```

The configuration uses `${ENV_VAR}` syntax for environment variable references. See [`configs/proxy_config.yaml`](configs/proxy_config.yaml) for the full format.

### Environment Variables

| Variable | Description | Required |
|---|---|---|
| `OPENAI_API_KEY` | OpenAI API key | Yes (for OpenAI models) |
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `REDIS_HOST` | Redis server hostname | No (falls back to memory-only) |
| `REDIS_PORT` | Redis server port | No |
| `REDIS_PASSWORD` | Redis password | No |
| `MASTER_KEY` | Admin API key for proxy | Yes |

## Usage

### Chat Completions

```bash
curl http://localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer sk-admin-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming

```bash
curl http://localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer sk-admin-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

### Health Check

```bash
curl http://localhost:4000/health
```

## Project Structure

```
shahrah/
├── cmd/server/           # Application entry point
├── internal/
│   ├── auth/             # Authentication middleware & key store
│   ├── cache/            # Dual cache (memory + Redis)
│   ├── config/           # Configuration loader
│   ├── providers/        # LLM provider implementations
│   │   └── openai/       # OpenAI-compatible client
│   ├── proxy/            # Request handler & proxy logic
│   ├── router/           # Router & routing strategies
│   └── server/           # Fiber server setup
├── pkg/types/            # Shared type definitions
├── tests/
│   ├── unit/             # Unit tests
│   └── integration/      # Integration tests
├── migrations/           # Database migrations
├── configs/              # Configuration files
├── Makefile              # Build & dev commands
├── docker-compose.yml    # Docker Compose setup
└── Dockerfile            # Docker build file
```

## Routing Strategies

| Strategy | Description | Use Case |
|---|---|---|
| `simple-shuffle` | Random selection (supports weighted) | Default, even distribution |
| `least-busy` | Fewest active connections | High concurrency scenarios |
| `lowest-cost` | Cheapest deployment | Budget optimization |
| `lowest-latency` | Fastest response time | Low-latency requirements |
| `usage-based-routing` | Most remaining TPM/RPM | Capacity-based routing |

## Development

### Running Tests

```bash
# Run unit tests
make test-unit
# Or on Windows:
go test -v -race ./tests/unit/...

# Run integration tests (requires Docker)
make test-integration
# Or on Windows:
go test -v -race ./tests/integration/...

# Run all tests with coverage
make test-coverage
# Or on Windows:
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build release binaries
make release
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License. See [LICENSE](LICENSE) for details.


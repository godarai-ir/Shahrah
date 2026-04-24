# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Go LLM Proxy
- OpenAI-compatible chat completions endpoint (`/v1/chat/completions`)
- Streaming support via Server-Sent Events (SSE)
- Intelligent routing with simple-shuffle strategy
- Weighted load balancing across multiple deployments
- Automatic failover and cooldown for failed deployments
- Dual cache implementation (Ristretto memory + Redis)
- API key authentication middleware (Bearer, api-key, x-api-key headers)
- Master key support for admin access
- Configuration via YAML file with environment variable resolution
- Health and readiness endpoints
- Docker Compose for local development (PostgreSQL + Redis)
- Database migration scripts (schema based on LiteLLM)
- Unit test suite for auth and router components
- Cross-compilation support in Makefile
- GitHub Actions CI/CD pipeline
- golangci-lint configuration

### Changed
- Module path updated to `github.com/BerriAI/go-llm-proxy`
- Go version updated to 1.22

### Fixed
- Duplicate `HashKey` function removed from router package
- Import paths corrected for standalone repository

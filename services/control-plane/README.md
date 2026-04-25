# Control Plane

This service contains the current Go implementation for Shahrah's API gateway and control-plane logic.

## Structure

- `cmd/server`: entrypoint
- `internal`: service-private implementation
- `pkg/types`: shared public types for this module
- `configs`: local config templates
- `tests`: unit and integration tests

## Commands

```bash
go test ./...
go run ./cmd/server
go build ./cmd/server
```

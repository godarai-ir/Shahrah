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

## Required Configuration

- `DATABASE_URL`: required, startup fails fast when missing or unreachable
- `MASTER_KEY`: required for `/api/v1/*` management APIs
- `TUNNEL_CLIENT_URL`: control-plane hop into the tunnel client
- `REDIS_HOST` and `REDIS_PASSWORD`: optional, memory cache is used when Redis is not configured

## Local Verification

Apply migrations in this order:

1. `db/main/migrations/001_initial.sql`
2. `db/main/migrations/002_platform_foundations.sql`

Example flow after the stack is running on `http://localhost:4000`:

```bash
curl -s http://localhost:4000/api/v1/keys \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"key_alias":"tenant-key","models":["gpt-4o-mini"]}'

curl -s http://localhost:4000/api/v1/models \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model_name":"gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"http://provider.example/v1","api_key":"sk-provider"}'

curl -s http://localhost:4000/api/v1/wallet/topup \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"amount_irr":1500000}'

curl -s http://localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer $TENANT_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}'
```

Successful inference should create a wallet debit plus request and spend log records for the issuing tenant.

For a real Postgres-backed verification run:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sharah_db?sslmode=disable make test-integration-db
```

That target runs only the database-backed onboarding test and skips when `DATABASE_URL` is unset.

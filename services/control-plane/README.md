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

## Model Discovery Contract

`control-plane` consumes `model-discovery` as an upstream API for provider-catalog flows.

- Live Swagger UI: `http://localhost:3002/docs`
- Live OpenAPI JSON: `http://localhost:3002/docs/openapi.json`
- Shared contract notes: [docs/model-discovery-api-contract.md](/D:/source/Shahrah/docs/model-discovery-api-contract.md)
- Contract test: [services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go](/D:/source/Shahrah/services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go)

When changing `model-discovery` usage in `control-plane`:

1. Regenerate the spec with `pnpm --dir apps/model-discovery openapi:export`.
2. Update `control-plane` client logic.
3. Extend or adjust the OpenAPI contract assertions for any new upstream endpoint or critical parameter constraint.

Do not add undocumented fallback behavior that hides upstream contract drift.

## Required Configuration

- `DATABASE_URL`: required, startup fails fast when missing or unreachable
- `MASTER_KEY`: required for `/api/v1/*` management APIs
- `SHAHRAH_DISCOVERY_TOKEN`: optional dedicated model-discovery service credential; never falls back to `MASTER_KEY`. Leave unset while control-plane only consumes public catalog endpoints.
- `TUNNEL_CLIENT_URL`: control-plane hop into the tunnel client
- `REDIS_HOST` and `REDIS_PASSWORD`: optional, memory cache is used when Redis is not configured
- `TOKEN_USAGE_AGGREGATION_ENABLED`: optional (`true` by default), queues token spend and key last-used aggregates through Redis instead of updating `verification_tokens` per request
- `TOKEN_USAGE_REDIS_FAIL_OPEN`: optional (`true` by default), keeps gateway traffic moving when Redis aggregation is unavailable; `spend_logs` remains the billing source of truth
- `TOKEN_USAGE_REDIS_KEY_PREFIX`: optional, default `usage`
- `TOKEN_USAGE_FLUSH_INTERVAL_SECONDS`: optional, default `5`
- `TOKEN_USAGE_FLUSH_BATCH_SIZE`: optional, default `500`
- `TOKEN_USAGE_DELTA_TTL_SECONDS`: optional, default `86400`
- `TOKEN_LAST_USED_FLUSH_INTERVAL_SECONDS`: optional, default `60`
- `TOKEN_LAST_USED_TTL_SECONDS`: optional, default `86400`
- `ZARINPAL_MERCHANT_ID`: required when `provider=zarinpal`
- `ZARINPAL_BASE_URL`: optional, default `https://api.zarinpal.com`
- `IDPAY_API_KEY`: required when `provider=idpay`
- `IDPAY_BASE_URL`: optional, default `https://api.idpay.ir/v1.1`
- `IDPAY_SANDBOX`: optional (`true|false`)
- `CLICKHOUSE_DSN`: required, used for request traces, request logs, and model metrics storage/read APIs
- `LOG_RETENTION_CLEANUP_ENABLED`: optional (`true` by default), enables entitlement-based cleanup for `request_logs_ch`
- `LOG_RETENTION_CLEANUP_INTERVAL_MS`: optional, default `3600000`
- `LOG_RETENTION_CLEANUP_TIMEOUT_MS`: optional, default `30000`
- `OPENAI_API_KEY`: optional, used by the default OpenAI entries in `configs/proxy_config.yaml`
- `ANTHROPIC_API_KEY`: optional, used by the default Anthropic entries in `configs/proxy_config.yaml`
- `SHAHRAH_SMART_CACHE_EMBEDDING_BASE_URL`: ArvanCloud BGE-M3 OpenAI-compatible base URL ending in `/v1`; treat the complete opaque URL as a secret
- `SHAHRAH_SMART_CACHE_EMBEDDING_API_KEY`: bearer credential for the Iran-hosted BGE-M3 endpoint

Platform workloads use database-backed service credentials instead of `MASTER_KEY`. Provision `support-admin` with `platform:read`, `platform:catalog:write`, and `platform:tenant:write` using `go run ./services/control-plane/cmd/platform-service-account`.

## Local Verification

Apply migrations in this order:

1. `db/main/migrations/001_initial.sql`
2. `db/main/migrations/002_platform_foundations.sql`
3. `db/main/migrations/010_personal_access_tokens.sql`
4. `db/main/migrations/012_drop_postgres_trace_tables.sql`

Example flow after the stack is running on `http://localhost:4000`:

```bash
curl -s http://localhost:4000/api/v1/personal-access-tokens \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "x-godarai-user-id: <tenant-user-id>" \
  -H "Content-Type: application/json" \
  -d '{"name":"tenant-pat"}'

curl -s http://localhost:4000/api/v1/models \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model_name":"gpt-4o-mini","provider":"openai","model":"gpt-4o-mini","api_base":"http://provider.example/v1","api_key":"sk-provider"}'

curl -s http://localhost:4000/api/v1/wallet/topup/intents \
  -H "Authorization: Bearer $MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{"amount_usd":100,"provider":"mock"}'

curl -i "http://localhost:4000/api/v1/wallet/topup/callback/mock?authority=mock_123&status=ok"

curl -s "http://localhost:4000/api/v1/wallet/transactions?limit=20" \
  -H "Authorization: Bearer $MASTER_KEY"

curl -s http://localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer $TENANT_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}'
```

Successful inference should create a wallet debit plus request and spend log records for the issuing tenant.

## Gateway Response Caching

Response caching is disabled by default and can be enabled with `response_cache.enabled` in `configs/proxy_config.yaml`. Clients opt in per non-streaming chat request:

```bash
curl -s http://localhost:4000/v1/chat/completions \
  -H "Authorization: Bearer $TENANT_KEY" \
  -H "Content-Type: application/json" \
  -H 'x-godarai-cache-config: {"type":"simple","ttl":600,"namespace":"production"}' \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}'
```

Use `type: smart` with an optional `similarity_threshold` from `0.80` through `1.0` for semantic matching. Smart caching first checks the exact request and then embeds only the final user text with the Iran-hosted ArvanCloud `Bge-m3` service. All prior messages and request parameters remain exact-match constraints. Streaming, multimodal, and tool requests bypass smart caching.

Response headers are:

- `x-godarai-cache-status`: `hit`, `miss`, `bypass`, or `error`
- `x-godarai-cached-trace-id`: source trace on a hit
- `x-godarai-cache-similarity-score`: smart semantic hits only

Entries are isolated by tenant, virtual key/account, namespace, logical model, virtual-model version, and embedding generation. Cache or embedding failures fail open to normal inference. Cache hits do not call the provider or debit the tenant wallet. The complete ArvanCloud base URL and API key must come from runtime secrets and must never be logged or committed.

## OpenAI-Compatible Moderations

Shahrah exposes moderation-capable models through:

```bash
curl -s http://localhost:4000/v1/moderations \
  -H "Authorization: Bearer $TENANT_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai:default:omni-moderation-latest","input":"I want to check this text."}'
```

Example response:

```json
{
  "id": "modr_example",
  "model": "omni-moderation-latest",
  "results": [
    {
      "flagged": false,
      "categories": {
        "sexual": false,
        "hate": false,
        "harassment": false,
        "self-harm": false,
        "violence": false
      },
      "category_scores": {
        "sexual": 0.0001,
        "hate": 0.0002,
        "harassment": 0.0001,
        "self-harm": 0,
        "violence": 0.0003
      }
    }
  ]
}
```

A model is compatible with `/v1/moderations` only when the model-discovery `modes` array includes `"moderation"` and the tenant provider account/key exposes that model with moderation capability. Chat-only, completion-only, or embedding-only models are rejected on this endpoint. Moderation traces record metadata such as model, input count, input lengths, and short hashes; raw moderated text is not written to trace payloads.

## Personal Access Tokens (PAT)

Canonical PAT endpoints:

- `GET /api/v1/personal-access-tokens`
- `POST /api/v1/personal-access-tokens`
- `GET /api/v1/personal-access-tokens/exists?name=...`
- `DELETE /api/v1/personal-access-tokens/:id`
- `DELETE /api/v1/personal-access-tokens/revoke/all`

Notes:

- PATs are tenant-scoped and user-owned.
- PAT model access is derived from Provider Account and Virtual Model access mappings.
- Raw token is returned only on creation.
- Revocation is soft (`revoked_at`) and revoked tokens are rejected by auth lookup.

Rollback notes:

- If rollback is required, disable PAT endpoints first, then drop PAT columns/indexes from `verification_tokens` in a new rollback migration (do not edit applied migration files).

## Wallet Top-up Flow (Implemented)

1. `POST /api/v1/wallet/topup/intents` creates a pending top-up intent from USD amount.
2. Control-plane converts USD to IRR using current exchange rate, then applies tax/fee settings.
3. Gateway create-payment is called (`mock`, `zarinpal`, `idpay`) and returns `payment_url` + `authority`.
4. User is redirected to provider payment page.
5. Provider callback hits `GET|POST /api/v1/wallet/topup/callback/:provider`.
6. Control-plane verifies payment server-to-server with provider API.
7. On success, wallet is credited and a `DEPOSIT` transaction is inserted atomically.
8. Callback redirects to tenant app URL with status query params.

Idempotency: callback processing is idempotent by unique `authority` + row-level transactional completion.

For a real Postgres-backed verification run:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sharah_db?sslmode=disable make test-integration-db
```

That target runs only the database-backed onboarding test and skips when `DATABASE_URL` is unset.

# Model Discovery Service

Production-focused Model Catalog backend (no LLM gateway/execution).

## Stack
- NestJS + TypeScript
- PostgreSQL + Prisma
- Typesense
- Redis + BullMQ
- Swagger/OpenAPI

## Typesense compatibility
- Server target: `typesense/typesense:29.1`
- JS client package: `typesense@3.0.6`
- `typesense-js` v3 is backward-compatible with Typesense server versions older than v30, including v29.1.

## Typesense read-path rollout
- `GET /api/v1/models` now prefers Typesense for search, filter, sort, pagination, and count-heavy listing flows.
- `GET /api/v1/facets` now prefers Typesense facet counts and uses Postgres only for small lookup tables and fallback.
- Postgres remains the source of truth for:
  - writes and admin upserts
  - detail-rich reads such as `GET /api/v1/models/:canonicalId`
  - compare/detail relations that do not benefit enough from index-only reads
- If Typesense is unavailable, list and facet endpoints fall back to Postgres. Readiness can still fail in strict environments.

### Required env vars
- `TYPESENSE_HOST`
- `TYPESENSE_PORT`
- `TYPESENSE_PROTOCOL`
- `TYPESENSE_API_KEY`
- `TYPESENSE_COLLECTION_MODELS`
- `TYPESENSE_READS_ENABLED`
- `TYPESENSE_READY_STRICT`

### Environment expectations
- `develop`: keep `TYPESENSE_READS_ENABLED=true` and `TYPESENSE_READY_STRICT=true` so local behavior matches higher environments.
- `stage`: enable the same flags and validate full reindex, incremental reindex, and fallback behavior before promotion.
- `production`: provision persistent storage, keep Typesense internal-only, enable strict readiness, and make sure the GitOps repo carries the same env values and secret wiring.

## Run locally
1. `cp .env.example .env`
2. `pnpm --filter model-discovery prisma:generate`
3. `pnpm --filter model-discovery prisma:migrate:dev`
4. Run your dedicated seed/import script (default sample seed is disabled)
5. `pnpm --filter model-discovery start:dev`

Swagger: `http://localhost:3002/docs`
OpenAPI JSON: `http://localhost:3002/docs/openapi.json`
OpenAPI YAML: `http://localhost:3002/docs/openapi.yaml`
API base: `http://localhost:3002/api/v1`

## Protected endpoint authentication

Admin endpoints accept only an opaque service-account bearer key. Human credentials and support-admin browser sessions are intentionally not valid here.

Create or rotate the `support-admin` credential after applying Prisma migrations:

```powershell
$env:DATABASE_URL = "postgresql://.../model_discovery?schema=public"
pnpm --filter model-discovery service-account:create -- --name=support-admin --scopes=catalog:read,catalog:write,provider:write,import:execute,indexing:execute,display-settings:write
```

The command prints the credential once. Store it as `MODEL_DISCOVERY_SERVICE_TOKEN` in the `support-admin` secret store. Rotation is additive: deploy the new credential first, then revoke the previous key ID:

```powershell
pnpm --filter model-discovery service-account:revoke -- --key-id=<key-id>
```

Export the checked-in spec artifact with:

- `pnpm --dir apps/model-discovery openapi:export`

Exported files:

- `apps/model-discovery/openapi/model-discovery.openapi.json`
- `apps/model-discovery/openapi/model-discovery.openapi.yaml`

## Important endpoints
- `GET /api/v1/models`
- `GET /api/v1/models/:canonicalId`
- `GET /api/v1/models/compare?ids=...`
- `GET /api/v1/providers`
- `GET /api/v1/providers/:slug`
- `GET /api/v1/facets`
- `GET /api/v1/categories`
- `GET /api/v1/capabilities`
- `GET /api/v1/supported-parameters`
- `GET /api/v1/modalities`
- `GET /api/v1/regions`
- `GET /api/v1/settings/display`

## Admin auth (dev)
- `POST /api/v1/auth/login`
- body: `{ "username": "admin", "password": "admin123" }`
- use returned bearer token for admin endpoints.

## Admin endpoints
- `POST /api/v1/admin/models/upsert`
- `GET /api/v1/admin/models/audit-log`
- `POST /api/v1/admin/indexing/rebuild`
- `POST /api/v1/admin/indexing/model`
- `POST /api/v1/admin/import/catalog-json`
- `POST /api/v1/providers/admin/upsert`
- `PUT /api/v1/settings/display`

## Display settings and exchange rate
- `GET /api/v1/settings/display` returns display preferences plus cached `USD->IRR` exchange rate.
- `PUT /api/v1/settings/display` updates settings (`currencyUnit`, `dateCalendar`, `priceScale`) and requires admin bearer token.
- `EXCHANGE_RATE_API_URL` controls the exchange-rate source label. Current implementation is mocked but cache-ready for a real provider.

## Seed/import files
- Sample import JSON: `seed/model-catalog.sample.json`
- Default seed script: `prisma/seed.ts` (disabled by design)
- YAML importer script: `scripts/import-models-from-yaml.ts`
- Versioned provider catalog: `providers/`

### Import from provider YAML folders
- Provider YAML files are maintained directly under
  `apps/model-discovery/providers` and are included in the migration image.
- Import the curated OpenAI catalog with OpenAI-specific allowlist validation:
  - `pnpm --filter model-discovery import:provider:openai`
- Parse and validate OpenAI without writing to the database:
  - `pnpm --filter model-discovery import:provider:openai -- --dry-run`
- Reconcile the OpenAI database catalog, deleting rows absent from the curated
  allowlist:
  - `pnpm --filter model-discovery import:provider:openai -- --clear-missing`
- Import the curated Anthropic catalog with Anthropic-specific allowlist,
  pricing, and runtime-profile validation:
  - `pnpm --filter model-discovery import:provider:anthropic`
- Parse and validate Anthropic without writing to the database:
  - `pnpm --filter model-discovery import:provider:anthropic -- --dry-run`
- Reconcile the Anthropic database catalog, deleting rows absent from the
  curated allowlist:
  - `pnpm --filter model-discovery import:provider:anthropic -- --clear-missing`
- Import all checked-in providers:
  - `pnpm --filter model-discovery import:providers:yaml`
- Import specific providers:
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai,anthropic`
- Dry-run (parse + summarize only):
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai --dry-run`
- Optional sync cleanup (delete DB models missing from YAML for selected providers):
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai --clear-missing`
- Full catalog reconciliation (also delete providers whose entire directory is
  missing):
  - `pnpm --filter model-discovery import:providers:yaml -- --clear-missing --clear-missing-providers`
  - `--clear-missing-providers` is intentionally incompatible with
    `--providers`; it must compare the complete checked-in catalog with the
    database.
- Custom providers root:
  - `pnpm --filter model-discovery import:providers:yaml -- --providers-root <path>`

No sample seed data is created by default.

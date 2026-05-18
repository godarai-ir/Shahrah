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

## Run locally
1. `cp .env.example .env`
2. `pnpm --filter model-discovery prisma:generate`
3. `pnpm --filter model-discovery prisma:migrate:dev`
4. Run your dedicated seed/import script (default sample seed is disabled)
5. `pnpm --filter model-discovery start:dev`

Swagger: `http://localhost:3002/docs`
API base: `http://localhost:3002/api/v1`

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

## Seed/import files
- Sample import JSON: `seed/model-catalog.sample.json`
- Default seed script: `prisma/seed.ts` (disabled by design)
- YAML importer script: `scripts/import-models-from-yaml.ts`

### Import from provider YAML folders
- Import all providers under `D:\source\models\providers`:
  - `pnpm --filter model-discovery import:providers:yaml`
- Import specific providers:
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai,anthropic`
- Dry-run (parse + summarize only):
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai --dry-run`
- Optional sync cleanup (delete DB models missing from YAML for selected providers):
  - `pnpm --filter model-discovery import:providers:yaml -- --providers openai --clear-missing`
- Custom providers root:
  - `pnpm --filter model-discovery import:providers:yaml -- --providers-root D:\source\models\providers`

No sample seed data is created by default.

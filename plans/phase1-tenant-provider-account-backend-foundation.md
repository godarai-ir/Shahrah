# Phase 1: Tenant Provider-Account Backend Foundation

> Status: Completed  
> Date: 2026-04-27  
> Scope: Shared DB schema + control-plane APIs + tenant-admin proxy routes

## Objective

Implement the backend foundation required by the tenant model onboarding wizard:

1. Provider catalog
2. Provider account create/list/get
3. Model selection persistence
4. Access control persistence
5. Credential test endpoint (stub-safe behavior)

## Deliverables

- SQL migration:
  - `db/main/migrations/003_tenant_provider_accounts.sql`
- Shared DB schema update:
  - `packages/database/schema.prisma`
- Control-plane implementation:
  - `services/control-plane/internal/admin/provider_catalog.go`
  - `services/control-plane/internal/admin/handlers.go`
  - `services/control-plane/internal/storage/tenant_provider_account_repo.go`
  - `services/control-plane/internal/server/server.go`
- Tenant-admin proxy API routes:
  - `apps/tenant-admin/app/api/provider-catalog/providers/route.ts`
  - `apps/tenant-admin/app/api/provider-catalog/providers/[provider]/models/route.ts`
  - `apps/tenant-admin/app/api/tenant-provider-accounts/**`

## API Surface Added

- `GET /api/v1/provider-catalog/providers`
- `GET /api/v1/provider-catalog/providers/:provider/models`
- `GET /api/v1/access-principals?kind=user|team&q=&limit=`
- `POST /api/v1/tenant-provider-accounts`
- `GET /api/v1/tenant-provider-accounts`
- `GET /api/v1/tenant-provider-accounts/:accountID`
- `PUT /api/v1/tenant-provider-accounts/:accountID/models`
- `PUT /api/v1/tenant-provider-accounts/:accountID/access`
- `POST /api/v1/tenant-provider-accounts/:accountID/test-credentials`

## Validation

- `go test ./services/control-plane/...` passed.
- `go test ./packages/contracts/...` passed.
- `pnpm -C apps/tenant-admin exec tsc --noEmit` passed.
- `node scripts/sync-prisma.js` executed successfully with local `DATABASE_URL` (migrations `003` and `004` applied).
- Prisma clients regenerated for:
  - `packages/database`
  - `apps/tenant-admin`
  - `apps/support-admin`

## Product Decisions Applied

- Access picker source is platform-level `users` and `teams` tables.
- `manager` role allows only `user` and `team` principals (enforced in validation and DB check constraint).
- Provider/model catalog is fetched dynamically from discovery service:
  - `SHAHRAH_MODEL_DISCOVERY_URL`
  - optional `SHAHRAH_DISCOVERY_TOKEN`

## Deferred to Phase 2

- Tenant-admin multi-step wizard UI replacement.
- Real provider credential validation against upstream APIs.
- Runtime enforcement of account-level access rules in request path.
- User/team searchable picker integration for access assignment.

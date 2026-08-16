# Phase 0: Tenant Provider-Account Contract Lock

> Status: Completed  
> Date: 2026-04-27  
> Scope: Contract + boundaries only (no runtime behavior change)

## Objective

Lock the product and API contract for the new tenant-admin flow:

1. Provider selection
2. Provider-model selection + model-level settings
3. Access control (manager/user)

This phase intentionally avoids DB/runtime cutover and focuses on stable contracts for Phase 1 migrations and Phase 2 API implementation.

## Inputs Reviewed

- `apps/tenant-admin` existing model/key dashboard and auth flow
- `services/control-plane/internal/admin/handlers.go` current management API surface
- `db/main/migrations/*.sql` + `packages/database/schema.prisma` shared DB shape

## Phase 0 Deliverables

- Shared Go contracts added in:
  - `packages/contracts/platform/tenant_provider_account.go`
- Locked vocabulary for:
  - auth modes
  - model types
  - pricing modes
  - principal types
  - access roles
- Request/response DTOs for upcoming endpoints:
  - provider catalog listing
  - provider model listing
  - create provider account
  - upsert provider account models
  - update provider account access control
  - test credentials

## Contract Decisions (Locked)

1. New root entity is `TenantProviderAccount`.
2. Model selection is stored as account-scoped `TenantProviderAccountModel[]`.
3. Pricing per model supports `public` and `private` modes.
4. Access control has two explicit buckets:
   - `manager`
   - `user`
5. Principals are typed:
   - `user`
   - `team`
   - `everyone`
6. Provider auth mode is explicit, not inferred from provider/model names.
7. Credential testing is an explicit API call and remains optional in the wizard.

## Assumptions to Unblock Phase 1/2

1. `application-plane` in discussions maps to the current `control-plane` service in this repository.
2. Step-1 provider list should come from control-plane catalog APIs (not hardcoded UI values).
3. Account creation can proceed without a BYOK credential if virtual-key usage is selected.
4. Access control defined in step-3 is authoritative for both:
   - tenant-admin management actions
   - runtime model access filtering (implemented in later phase)
5. Provider model metadata (types/pricing capabilities) is catalog-driven and versionable.

## Include Now / Defer / Exclude

### Include Now

- Stable contract for provider/model/access wizard
- BYOK optional credential structure
- Private pricing override structure

### Defer

- Auto-sync of pricing maps from external sources
- Advanced access groups across MCP/agents
- Key rotation and enterprise-only policy controls

### Exclude in this track

- Full endpoint parity across org/team/user management APIs outside this contract

## Exit Criteria

1. Phase 1 can add migrations without rethinking API vocabulary.
2. Phase 2 can implement endpoints directly from shared contracts.
3. Tenant-admin wizard implementation can use typed DTOs without contract churn.

## Execution Result

- Added shared contracts at `packages/contracts/platform/tenant_provider_account.go`.
- Verified formatting with `gofmt`.
- Verified compile with `go test ./packages/contracts/...` (passed).

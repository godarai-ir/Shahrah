# Execution Prompt: Implement Cost Center in Shahrah

Use this prompt with a coding agent that has full access to the GodarAI repository.

---

You are implementing a **Cost Center** system in GodarAI.

## Context

- Repo root: `d:\source\Shahrah`
- Core service: `services/control-plane`
- Existing behavior:
  - key-level controls in `verification_tokens` (`max_budget`, `rpm_limit`, `tpm_limit`)
  - tenant wallet checks in billing service
  - spend logging in `spend_logs`
- Existing architecture/docs:
  - `plans/cost-center-implementation-plan.md`
  - `plans/02_multi_tenant_architecture.md`
  - `plans/01_iranian_payment_system.md`

## Goal

Add first-class cost-center policy enforcement and attribution without breaking current key/wallet enforcement.

## Hard Requirements

1. Preserve backward compatibility:
   - Do not remove or weaken existing key `max_budget/rpm/tpm` checks.
   - Do not remove tenant wallet balance checks/debit logic.
2. Add policy-based cost center rules:
   - conditions + group_by + priority + hard/soft enforcement.
3. Attribute request usage to one or more cost-center buckets.
4. Support metadata dimensions (`metadata.<key>`).
5. Provide admin CRUD and usage query APIs.
6. Add tests for policy matching and enforcement.

## Implementation Scope

### A) Data Model and Migrations

Create migrations in `db/main/migrations` for:

- `cost_center_policies`
  - id, tenant_id, name, enabled, priority
  - conditions_json, group_by_json
  - metric_type (`cost_usd|tokens|requests`)
  - limit_value, window_unit
  - enforcement_mode (`hard|soft`)
  - alert_thresholds_json
  - created_at, updated_at
- `cost_center_usage`
  - tenant_id, policy_id, bucket_key, window_start, window_end
  - used_cost_usd, used_tokens, used_requests
  - created_at, updated_at
  - unique composite index on `(tenant_id, policy_id, bucket_key, window_start)`
- `cost_center_alert_events`
  - tenant_id, policy_id, bucket_key, threshold, used_value, limit_value, emitted_at

### B) Contracts / Types

Add typed structures under `packages/contracts/platform` for:

- `CostCenterPolicy`
- `CostCenterCondition`
- `CostCenterGroupByKey`
- `CostCenterUsageRow`
- API request/response DTOs for CRUD/list/usage endpoints

### C) Service Module

Add `services/control-plane/internal/costcenter`:

- `resolver.go`
  - resolve dimensions from auth + tenant + request metadata + route/model/provider
- `policy_store.go`
  - load active tenant policies, cache with TTL, invalidate on mutation
- `matcher.go`
  - deterministic policy matching (AND over conditions)
- `bucket.go`
  - deterministic bucket key generation from group_by
- `window.go`
  - compute window start/end for daily/weekly/monthly/rolling
- `enforcer.go`
  - evaluate matched policies, return allow/block + reason
- `tracker.go`
  - update usage counters for all matched policies
- `alerts.go`
  - detect threshold crossings and persist alert events

### D) Proxy Integration

Integrate in both:

- `services/control-plane/internal/proxy/handler.go`
- `services/control-plane/internal/proxy/completions_handler.go`

Flow:

1. Keep existing auth + key-level checks + wallet checks.
2. Before upstream call:
   - estimate request usage
   - run cost-center precheck (hard policies may block)
3. After response:
   - finalize actual usage
   - track usage for all matched policies
   - emit soft-limit alerts if needed

### E) Storage Layer

Add repository operations in `services/control-plane/internal/storage` for:

- create/list/update/delete cost-center policies
- upsert usage counters
- list usage with filters
- list alert events

### F) Admin API Layer

Add handlers/routes in control-plane admin module:

- `POST /api/v1/cost-centers/policies`
- `GET /api/v1/cost-centers/policies`
- `PATCH /api/v1/cost-centers/policies/:id`
- `DELETE /api/v1/cost-centers/policies/:id`
- `GET /api/v1/cost-centers/usage`
- `GET /api/v1/cost-centers/usage/:policyId/buckets`
- `GET /api/v1/cost-centers/alerts`

Respect tenant scoping and existing RBAC style.

### G) Tenant Admin UI

In `apps/tenant-admin`:

- add policy list and form pages/components
- add usage dashboard view
- add alerts view

Reuse patterns near:

- `apps/tenant-admin/app/components/team-management.tsx`

### H) Feature Flags

Add config flags:

- `cost_center_enabled` (default false)
- `cost_center_enforcement_mode` (`off|dry_run|enforced`)

Behavior:

- `off`: no matching/tracking
- `dry_run`: evaluate + track + alert, no block
- `enforced`: hard limits block

### I) Tests

Add:

- unit tests for matcher, bucket key derivation, windowing
- unit tests for hard/soft enforcement
- integration tests for proxy path with:
  - team-based rule
  - metadata-based rule
  - overlapping policies with priority
- regression tests confirming old key/wallet behavior unchanged

## Coding Constraints

- Keep changes minimal and modular.
- Follow existing naming and package structure.
- Use tenant_id in all relevant queries.
- Avoid introducing breaking changes to existing APIs.
- Add concise comments only where logic is non-obvious.

## Done Definition

1. Migrations apply cleanly.
2. New APIs work with tenant scoping.
3. Proxy enforces cost-center policies under feature flag.
4. Usage and alerts persist and can be queried.
5. Existing budget/rate-limit/wallet tests still pass.
6. New tests pass.

## Output Format Required From You

When implementation is complete, return:

1. Summary of files changed
2. Key design decisions
3. API examples (create policy + usage query)
4. Test commands executed and results
5. Any follow-up tasks explicitly marked as TODO


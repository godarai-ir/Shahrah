# Shahrah Cost Center Implementation Plan

## 1) Objective

Implement a production-grade **Cost Center** capability in Shahrah that:

- attributes usage and spend to configurable business dimensions
- enforces budget/rate limits at those dimensions
- preserves current tenant isolation and wallet-based billing behavior
- supports gradual rollout without breaking existing key-level controls

## 2) Current State (Codebase Reality)

Based on current control-plane code:

- Strong enforcement exists at:
  - **Key level**: `max_budget`, `rpm_limit`, `tpm_limit` from `verification_tokens`
  - **Tenant wallet level**: IRR balance check + debit
- Spend attribution is logged with:
  - `tenant_id`, `organization_id`, `team_id`, `token_hash`, `model`, `cost_usd`, `cost_irr`
- IAM entities already include:
  - `OrgUnit`, `Team`, `TeamMembership`, `VirtualUser`, `BudgetID`
- Gap:
  - No generalized policy/rule engine for cost centers (workspace/team/project/metadata buckets)

## 3) Target Capability

### 3.1 Cost Center Definition

A **Cost Center** is a dynamic spend bucket resolved per request by:

- principal dimensions: `tenant_id`, `org_unit_id`, `team_id`, `virtual_user_id`, `key_id`
- request dimensions: `model`, `provider`, `endpoint_type`
- metadata dimensions: `metadata.<key>` (example: `project_id`, `environment`)

### 3.2 Policy Model

Introduce `cost_center_policies` with:

- `name`, `enabled`, `priority`
- `conditions` (AND logic):
  - keys like `team_id`, `org_unit_id`, `key_id`, `model`, `provider`, `metadata.project_id`
- `group_by` (bucket key creation):
  - one or more keys like `team_id`, `metadata.project_id`, `model`
- `metric_type`:
  - `cost_usd` | `tokens` | `requests`
- `limit_value`, `window_unit`:
  - `daily`, `weekly`, `monthly`, `rolling_1m`, `rolling_1h`
- `enforcement_mode`:
  - `hard` (block) | `soft` (warn only)
- `alert_thresholds`:
  - percent checkpoints (e.g. 75, 90, 100)

### 3.3 Enforcement Semantics

- Multiple policies can match one request.
- First matching policy by priority decides allow/block for `hard` mode.
- Usage is tracked for **all matched policies**.
- Existing key and wallet checks remain active; cost-center checks are additive.

## 4) Architecture Changes

### 4.1 Data Layer (Postgres)

Add migrations in `db/main/migrations`:

1. `cost_center_policies` table
2. `cost_center_usage` table
3. `cost_center_alert_events` table
4. indexes on `(tenant_id, policy_id, bucket_key, window_start)`

### 4.2 Cache Layer (Redis)

Use Redis for fast counters:

- key pattern: `cc:{tenant}:{policy_id}:{bucket}:{window}`
- atomic increment per request
- periodic flush / reconciliation to Postgres

### 4.3 Control-Plane Services

Add module: `services/control-plane/internal/costcenter`

- `resolver.go`: builds request context dimensions
- `policy_store.go`: loads active policies per tenant
- `enforcer.go`: pre-check (estimated cost/tokens/requests)
- `tracker.go`: post-response finalize with actual usage
- `scheduler.go`: window reset / rollover tasks
- `alerts.go`: threshold crossing events

### 4.4 Proxy Integration

Integrate into:

- `services/control-plane/internal/proxy/handler.go`
- `services/control-plane/internal/proxy/completions_handler.go`

Flow:

1. existing auth + routing + key/wallet checks remain
2. estimate usage
3. call cost-center precheck
4. forward request
5. finalize actual usage
6. update cost-center counters and logs

### 4.5 Storage Integration

Extend spend logging to persist resolved cost-center bucket data:

- add optional fields in spend logs payload/metadata for:
  - `cost_center_policy_ids`
  - `cost_center_buckets`

## 5) API Surface

Add admin endpoints under tenant-admin authorization:

- `POST /api/v1/cost-centers/policies`
- `GET /api/v1/cost-centers/policies`
- `PATCH /api/v1/cost-centers/policies/:id`
- `DELETE /api/v1/cost-centers/policies/:id`
- `GET /api/v1/cost-centers/usage`
- `GET /api/v1/cost-centers/usage/:policyId/buckets`
- `GET /api/v1/cost-centers/alerts`

## 6) UI Surface (Tenant Admin)

Add pages/components in tenant-admin app:

- policy list + create/edit form
- usage dashboard by policy and bucket
- threshold alerts feed

Start from existing team management context:

- `apps/tenant-admin/app/components/team-management.tsx`

## 7) Rollout Strategy

### Phase 0: Guardrails

- feature flag: `cost_center_enforcement_enabled`
- dry-run mode: evaluate + track without blocking

### Phase 1: Read-only Attribution

- resolve dimensions + persist usage buckets
- no budget blocking

### Phase 2: Soft Limits

- threshold warnings + alert events

### Phase 3: Hard Limits

- block requests when policy hard limits exceed
- keep fallback to existing wallet/key controls

## 8) Testing Strategy

### Unit

- policy matching (conditions + priority)
- bucket key derivation
- window calculation and reset
- hard/soft enforcement behavior

### Integration

- end-to-end request through proxy with:
  - metadata-based bucketing
  - team-based bucketing
  - overlapping policies
- verify:
  - allow/block behavior
  - spend logs contain expected bucket annotations
  - counters increment accurately

### Regression

- existing key `max_budget`, `rpm`, `tpm` behavior unchanged
- wallet debit flow unchanged

## 9) Risks and Mitigations

1. Counter drift between Redis and Postgres
- mitigation: periodic reconciliation job + idempotent upserts

2. Performance overhead in request path
- mitigation: in-memory cached policy snapshots + O(n) policy eval with low n per tenant

3. False positives from estimated usage
- mitigation: precheck on estimate, finalize on actual usage, configurable safety margin

4. Migration complexity with legacy budget fields
- mitigation: keep legacy columns; introduce cost-center engine as additive capability

## 10) Acceptance Criteria

1. Tenant admin can define policies on team/org/model/metadata dimensions.
2. Requests are attributed to one or more buckets with deterministic keys.
3. Hard-limit policies block requests with clear error type and policy identifier.
4. Soft-limit policies emit threshold events without blocking.
5. Existing key-level budget/rate-limit and wallet checks remain backward compatible.
6. Usage and alerts are queryable via API and visible in tenant-admin UI.


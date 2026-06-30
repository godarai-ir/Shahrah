# Implementation Plan: Portkey-Inspired Request Logs

## Objective

Add a first-class `Request Logs` surface to Shahrah as a request-centric complement to the existing `Traces` feature. The MVP should give tenant operators a chronological, filterable, privacy-aware view of gateway requests with enough detail to answer day-1 operational questions quickly: what was sent, what model/provider handled it, how long it took, what it cost, whether retry or fallback happened, and how to jump into the related trace when deeper debugging is needed.

This plan assumes:

- Shahrah is not yet in production
- backward compatibility for legacy `request_logs` is not required
- migration of request logs to ClickHouse can be done in one shot

## Inputs Reviewed

- PortKey docs:
  - `.portkey-docs/markdown/product/docs__product__observability__logs.md`
  - `.portkey-docs/markdown/product/docs__product__observability__logs-export.md`
  - `.portkey-docs/markdown/product/docs__product__administration__configure-logs-access-permissions-in-workspace.md`
  - `.portkey-docs/markdown/product/docs__product__administration__configuring-request-logging.md`
  - `.portkey-docs/markdown/product/docs__product__observability__traces.md`
- Shahrah plans:
  - `plans/README.md`
  - `plans/request-traces-executable-prompt.md`
  - `plans/observability-platform-implementation-plan.md`
- Shahrah commercial and entitlement context:
  - `docs/pricing-table.md`

## Decision Summary

- Build `Logs` as a separate product surface from `Traces`, but keep a strong `log -> trace` deep link.
- Move `Request Logs` to ClickHouse in one step; do not dual-write request-log payloads to PostgreSQL.
- Keep PostgreSQL as the source of truth for billing, wallet, auth, and audit-like transactional records.
- Make ClickHouse the single query source for tenant-facing request logs APIs and UI.
- Ship logs permissions as two levels from the start:
  - `logs:read_metadata`
  - `logs:read_content`
- Treat privacy mode as part of the product, not as a later hardening detail.
- Keep export jobs, replay, and manual feedback out of MVP.

## PortKey Capability Map

| PortKey capability | User/job value | MVP decision | Shahrah adaptation | Rationale |
|---|---:|---|---|---|
| Chronological request logs list | high | include now | Separate `Logs` list in tenant-admin with time, provider, model, status, tokens, cost, and latency | Core operator workflow |
| Detailed request/response inspection | high | include now | Detail page or drawer with permission- and privacy-aware payload rendering | Required for troubleshooting |
| Status visibility for cache/retry/fallback/load balancing | high | include now | First-class routing outcome fields and badges | Distinguishes logs from traces |
| Shareable deep link per log | medium | include now | Stable per-log route inside tenant-admin | Cheap and useful |
| Tracing link and grouped lifecycle context | high | include now | Link each log to trace detail and show trace summary | Reuses Shahrah's trace investment |
| Metrics-only / do-not-track mode | high | include now | Tenant logging policy plus optional request override header later | Strong privacy and compliance value |
| Workspace override for logging mode | medium | defer | Start tenant-wide, revisit org-unit or role override later | Simpler MVP |
| Logs access split: metadata vs content | high | include now | Separate permissions at launch | Needed if logs content is available in MVP |
| Logs export jobs | medium | defer | Add async export jobs after logs UX is stable | Useful but not required for launch |
| Replay request from log | medium | defer | Consider only after prompt and safe sandbox flows mature | High surface area and safety risk |
| Manual feedback on logs | low | defer | Wait for prompt analytics/evals | Not needed for initial operator workflows |

## Architecture

Shahrah should treat `Request Logs` as a product-layer query model optimized for observability, not as a normalized transactional table.

### Proposed split

- Gateway/proxy layer:
  - emits a normalized request outcome event for each completed request
  - emits trace spans independently
  - never blocks the request path on ClickHouse query availability
- Persistence layer:
  - ClickHouse stores request-log rows optimized for list/detail/filter workloads
  - PostgreSQL remains the source of truth for `spend_logs`, wallet changes, auth, and audit-like business records
- Observability layer:
  - request-log writer
  - trace writer
  - metrics writer
  - shared privacy and redaction policy
- Admin/control surface:
  - `Logs` is request-centric
  - `Traces` is span-centric
  - `Metrics` remains aggregate-centric

### Why ClickHouse is the correct primary store for request logs

Request logs are append-heavy, retention-heavy, and query-heavy:

- time-window filtering
- provider/model/route/status filtering
- top-N analysis
- support searches over recent history
- plan-based retention windows
- high-cardinality metadata and routing outcomes

These are ClickHouse workloads, not PostgreSQL workloads.

### What stays in PostgreSQL

- `spend_logs`
- wallet ledger and payment rows
- PAT/VAT and auth state
- subscriptions and commercial state
- audit-style operator mutations

### One-shot migration decision

Because Shahrah is not in production, request logs should move in one step:

- stop treating legacy PostgreSQL `request_logs` as part of the target architecture
- write new request logs only to ClickHouse
- build logs APIs only on ClickHouse
- do not spend time on dual-write, backfill, or backward compatibility for `request_logs`

## Product Scope

### Include Now

- Separate `Logs` item in tenant-admin sidebar
- `GET /api/v2/logs`
- `GET /api/v2/logs/:logID`
- Deep link to related trace when available
- Compact request list with:
  - timestamp
  - provider
  - model
  - virtual model
  - route
  - status code
  - latency
  - total tokens
  - cost
  - retry count
  - fallback status
  - cache status
- Detail view with:
  - request overview
  - routing outcomes
  - usage and cost
  - request payload when allowed
  - response payload when allowed
  - metadata
  - trace summary and trace link
- Logging policy modes:
  - `none`
  - `metadata-only`
  - `redacted-content`
  - `full-content`
- Dedicated permissions:
  - `logs:read_metadata`
  - `logs:read_content`
- Plan-based retention:
  - developer `3d`
  - team `30d`
  - growth `30d`
  - enterprise custom

### Defer

- Export jobs and download center
- Prompt replay from a log
- Manual feedback attached to logs
- Cross-tenant support-admin logs console
- Org-unit or team-specific logging policy overrides

### Exclude

- Full Portkey parity
- Full-text payload search in MVP
- Sync writes to analytical storage in request path
- Unlimited raw payload retention by default

## UX Principles

### Information architecture

- `Logs` answers: "what happened for this request?"
- `Traces` answers: "how did the request flow internally?"
- `Metrics` answers: "what is the aggregate pattern?"

The UI must keep these distinctions obvious to avoid operator confusion.

### Logs list behavior

- Default sort: newest first
- Filters visible above the list, not hidden behind advanced mode
- Fast filters:
  - time range
  - provider
  - model
  - route
  - status
  - error type
  - request id
  - trace id
- Status badges should summarize retry/fallback/cache/load-balance outcomes quickly

### Logs detail behavior

- Open from list without losing filter state
- Show summary first, raw payload second
- Always show why payload is hidden:
  - permission denied
  - logging mode metadata-only
  - payload truncated
  - do-not-track
- Make trace deep link obvious but secondary

### Empty and failure states

- No logs in retention window
- Logging disabled for tenant
- ClickHouse unavailable
- You can view metadata but not content

## API Contract

### `GET /api/v2/logs`

Query params:

- `from`
- `to`
- `provider`
- `model`
- `virtual_model`
- `route`
- `status_code`
- `error_type`
- `request_id`
- `trace_id`
- `cache_status`
- `fallback_status`
- `sort_by`
- `sort_order`
- `limit`
- `offset`

Response shape:

```json
{
  "data": [
    {
      "log_id": "log_...",
      "request_id": "req_...",
      "trace_id": "trace_...",
      "tenant_id": "tenant_...",
      "provider": "openai",
      "model": "gpt-4o",
      "virtual_model": "default-chat",
      "route": "/chat/completions",
      "status_code": 200,
      "error_type": "",
      "started_at": "2026-06-30T10:00:00Z",
      "duration_ms": 842,
      "input_tokens": 120,
      "output_tokens": 410,
      "total_tokens": 530,
      "cost_usd": 0.012,
      "cost_irr": 10560,
      "cache_status": "miss",
      "retry_count": 1,
      "fallback_status": "inactive",
      "load_balance_status": "active",
      "payload_mode": "metadata-only"
    }
  ],
  "total": 1234,
  "limit": 25,
  "offset": 0,
  "next_offset": 25
}
```

### `GET /api/v2/logs/:logID`

Returns:

- request overview
- routing outcome summary
- usage and cost summary
- content visibility state
- request payload section
- response payload section
- metadata
- trace summary

### Compatibility notes

- Public gateway API remains OpenAI-compatible
- Logs policy is an internal Shahrah control-plane behavior
- `trace_id` correlation remains stable across logs and traces

## Data Model

### ClickHouse primary table: `request_logs_ch`

- `log_id String`
- `tenant_id String`
- `request_id String`
- `trace_id String`
- `provider String`
- `model String`
- `virtual_model String`
- `route String`
- `status_code Int32`
- `error_type String`
- `started_at DateTime64(3, 'UTC')`
- `duration_ms Int32`
- `input_tokens UInt32`
- `output_tokens UInt32`
- `total_tokens UInt32`
- `cost_usd Decimal(18, 6)`
- `cost_irr Decimal(18, 2)`
- `cache_status LowCardinality(String)`
- `retry_count UInt16`
- `fallback_status LowCardinality(String)`
- `load_balance_status LowCardinality(String)`
- `actor_type LowCardinality(String)`
- `actor_id String`
- `team_id String`
- `org_unit_id String`
- `payload_mode LowCardinality(String)`
- `request_payload_json String`
- `response_payload_json String`
- `request_payload_truncated UInt8`
- `response_payload_truncated UInt8`
- `metadata_json String`

Suggested layout:

- partition by `toDate(started_at)`
- order by `(tenant_id, started_at, log_id)`

### Payload storage rules

- `metadata-only`:
  - no request/response content stored
  - only metadata and derived usage/routing fields stored
- `redacted-content`:
  - redacted and size-capped request/response JSON stored inline
- `full-content`:
  - full request/response JSON stored inline up to hard cap
  - if payload exceeds cap, store truncated inline content plus truncation markers
- `none`:
  - no payload content and no payload metadata beyond minimal operational counters

### Payload cap

- define explicit `LOG_PAYLOAD_MAX_BYTES`
- default MVP cap should be conservative, e.g. `32 KiB` per side
- large multimodal payloads should be summarized, not fully copied

### Source-of-truth rule for logs UI

Tenant-facing logs APIs should use ClickHouse as their single query source for:

- request summary
- request content visibility
- usage counters displayed in logs
- routing outcomes

PostgreSQL `spend_logs` remains the financial source of truth, but the logs UI should not require runtime joins to PostgreSQL for basic list/detail rendering.

## RBAC and Privacy

### Permissions

Add these permissions in MVP:

- `logs:read_metadata`
- `logs:read_content`

Rules:

- metadata list requires `logs:read_metadata`
- request/response payload rendering requires `logs:read_content`
- if a user lacks content permission, the UI still shows the metadata shell and the reason content is hidden

### Policy model

Do not reuse trace privacy mechanically. Logs need their own productized behavior even if implementation shares helpers.

Recommended tenant-level request logging policy:

- `none`
- `metadata-only`
- `redacted-content`
- `full-content`

Later, Shahrah can map these to lower-level trace capture behavior where useful, but logs policy should be defined in logs terms.

## Work Breakdown

### Phase 1: Contract, RBAC, and Policy

- [ ] Define request logs contracts in `packages/contracts/platform`
- [ ] Add `logs:read_metadata` and `logs:read_content` to permission catalog and system roles
- [ ] Add tenant request logging policy endpoints and storage
- [ ] Define canonical routing outcome vocabularies for cache, retry, fallback, and load balancing

### Phase 2: ClickHouse Write Path

- [ ] Add `request_logs_ch` schema creation in `services/control-plane/internal/storage`
- [ ] Replace PostgreSQL `LogRequest` path with ClickHouse-backed request-log writer
- [ ] Keep request-path behavior fail-open if analytical writes fail
- [ ] Write summary usage and routing fields directly into request log rows so logs UI does not need PostgreSQL joins

### Phase 3: Control-Plane Read APIs

- [ ] Add `GET /api/v2/logs`
- [ ] Add `GET /api/v2/logs/:logID`
- [ ] Add facets endpoint for provider, model, route, status, and error type
- [ ] Enforce tenant scoping plus metadata/content permissions
- [ ] Return explicit content-visibility state in detail responses

### Phase 4: Tenant Admin UX

- [ ] Add `Logs` item to sidebar next to `Traces` and `Metrics`
- [ ] Build request-centric list page with compact filters and status badges
- [ ] Build detail page or drawer optimized for summary first, payload second
- [ ] Add states for disabled logging, hidden content, retention expiry, and backend unavailability

### Phase 5: Retention and Hardening

- [ ] Enforce plan-based retention in ClickHouse cleanup jobs
- [ ] Add request payload truncation and redaction tests
- [ ] Benchmark ingestion overhead and recent-window query latency
- [ ] Validate that ClickHouse outages never break gateway traffic

## Testing Strategy

- Unit:
  - request outcome normalization
  - payload redaction and truncation
  - permission-based content visibility
  - retention policy decisions
  - routing outcome derivation
- Integration:
  - tenant isolation on logs APIs
  - metadata-only vs content-permitted behavior
  - ClickHouse fail-open write behavior
  - trace deep link correctness
- Performance:
  - request-log writer overhead
  - recent-window list query latency
  - high-cardinality metadata behavior

## Rollout Plan

1. add contracts, permissions, and tenant logging policy
2. add ClickHouse request logs table and write path
3. stop using PostgreSQL `request_logs` for the target architecture
4. ship internal logs APIs
5. ship tenant-admin logs UI
6. validate retention, privacy, and performance
7. remove obsolete request-log assumptions from legacy code and docs

## Risks

### Product risks

- Operators may still confuse `Logs` and `Traces` unless labels and empty states are explicit
- Exposing content without split permissions would create avoidable privacy risk

### Technical risks

- Inline payload storage can bloat ClickHouse if caps are too loose
- Routing status fields may be inconsistent until proxy/routing emits canonical signals
- If logs UI depends on PostgreSQL joins, it will become slower and more fragile than needed

### Operational risks

- Retention promises must match actual cleanup behavior
- ClickHouse outage handling must remain fail-open for gateway traffic

## Open Questions

- Should Shahrah also support a request-level override header in MVP, similar in spirit to Portkey `debug:false`, or should tenant policy be enough for the first release?
- For `full-content`, is inline capped storage enough for Shahrah MVP, or do we want encrypted blob references in the first version?
- Should support-admin get a cross-tenant logs explorer immediately after tenant-admin, or only after the tenant workflow stabilizes?

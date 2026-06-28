# Model Metrics Implementation Plan (ClickHouse)

## Objective
Implement a production-ready `Model Metrics` subsystem for GodarAI using ClickHouse, aligned with GodarAI multi-tenant architecture and the target dashboard (RPS, failures, latency percentiles, token usage, cost).

## Inputs Reviewed
- `D:\source\Shahrah\.truefoundry-docs\markdown\docs__ai-gateway__analytics.md`
- `D:\source\Shahrah\.truefoundry-docs\markdown\docs__ai-gateway__fetch-raw-metrics.md`
- `D:\source\Shahrah\.truefoundry-docs\markdown\docs__ai-gateway__cost-tracking.md`
- `D:\source\Shahrah\.truefoundry-docs\markdown\docs__ai-gateway__request-logging.md`
- `D:\source\Shahrah\plans\README.md`
- `D:\source\Shahrah\plans\09_shahrah_features.md`
- `D:\source\Shahrah\plans\05_admin_panel.md`
- `D:\source\Shahrah\plans\10_performance_availability.md`

## Scope Decisions (MVP)
### Include now
- Model metrics dashboard primitives:
  - Total input tokens
  - Total output tokens
  - Total request count
  - Total token cost (USD)
  - Requests per second
  - Request failure rate
  - Request failures breakdown
  - Request failure rate by error type
  - Request latency p50/p75/p90/p99
  - TTFT p50/p75/p90/p99
  - ITL p50/p75/p90/p99
  - TPOT p50/p75/p90/p99
  - Cost of inference over time
  - Input/output tokens over time
- Pivot dimensions:
  - `model`, `virtual_model`, `user`, `virtual_account`, `team`, `metadata.<key>`
- Query API:
  - `distribution` and `timeseries`
  - filters + group-by + interval
- Tenant-safe access control in metrics query API.

### Defer
- Advanced CSV export UX and report builder.
- Full parity tabs for MCP/Guardrail/Routing/Cache in same milestone.

### Exclude (MVP)
- Full feature parity clone with TrueFoundry analytics.
- Separate stream-processing stack (Flink/Kafka Streams) for v1.

## Architecture
- Producer:
  - `services/control-plane` emits per-request metric events asynchronously.
- Storage:
  - ClickHouse raw events table + materialized aggregate table.
- Query layer:
  - Control-plane metrics query service translates API filters to ClickHouse SQL.
- UI:
  - Admin metrics page consumes query API only (no direct DB calls).

## Metric Event Contract (Canonical)
All timestamps are UTC.

```json
{
  "tenant_id": "string",
  "request_id": "string",
  "ts": "RFC3339 timestamp",
  "model_name": "string",
  "virtual_model_name": "string|null",
  "provider_name": "string",
  "provider_model_name": "string",
  "request_type": "chat|responses|embeddings|...",
  "user_id": "string|null",
  "user_email": "string|null",
  "virtual_account_id": "string|null",
  "virtual_account_name": "string|null",
  "team_ids": ["string"],
  "team_names": ["string"],
  "status_code": 200,
  "error_type": "none|provider_timeout|rate_limit|budget_limit|...",
  "latency_ms": 0,
  "ttft_ms": 0,
  "itl_ms": 0,
  "tpot_ms": 0,
  "input_tokens": 0,
  "output_tokens": 0,
  "cost_usd": 0,
  "metadata": {
    "environment": "prod"
  }
}
```

## KPI Rules
- `request_count = count(*)`
- `failed_count = countIf(status_code >= 400)`
- `failure_rate = failed_count / request_count`
- `rps = request_count / bucket_seconds`
- Percentiles use `quantileTDigest`.
- `ttft_ms/itl_ms/tpot_ms` may be `NULL` for non-streaming responses.

## ClickHouse DDL (Initial)

```sql
CREATE TABLE IF NOT EXISTS gateway_model_events
(
  ts DateTime64(3, 'UTC'),
  tenant_id String,
  request_id String,

  model_name LowCardinality(String),
  virtual_model_name Nullable(LowCardinality(String)),
  provider_name LowCardinality(String),
  provider_model_name LowCardinality(String),
  request_type LowCardinality(String),

  user_id Nullable(String),
  user_email Nullable(String),
  virtual_account_id Nullable(String),
  virtual_account_name Nullable(String),

  team_ids Array(String),
  team_names Array(String),

  status_code UInt16,
  error_type LowCardinality(String),

  latency_ms UInt32,
  ttft_ms Nullable(UInt32),
  itl_ms Nullable(UInt32),
  tpot_ms Nullable(UInt32),

  input_tokens UInt32,
  output_tokens UInt32,
  cost_usd Decimal(18, 8),

  metadata_json String
)
ENGINE = MergeTree
PARTITION BY toDate(ts)
ORDER BY (tenant_id, ts, model_name, virtual_model_name)
TTL ts + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
```

```sql
CREATE TABLE IF NOT EXISTS gateway_model_metrics_1m
(
  bucket_ts DateTime('UTC'),
  tenant_id String,

  view_model_name LowCardinality(String),
  view_virtual_model_name Nullable(LowCardinality(String)),
  view_user_email Nullable(String),
  view_virtual_account_name Nullable(String),
  view_team_name Nullable(String),
  view_error_type LowCardinality(String),

  request_count UInt64,
  failed_count UInt64,

  input_tokens_sum UInt64,
  output_tokens_sum UInt64,
  cost_usd_sum Decimal(20, 8),

  latency_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  latency_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  latency_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  latency_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  ttft_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  ttft_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  ttft_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  ttft_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  itl_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  itl_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  itl_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  itl_p99 AggregateFunction(quantileTDigest(0.99), UInt32),

  tpot_p50 AggregateFunction(quantileTDigest(0.50), UInt32),
  tpot_p75 AggregateFunction(quantileTDigest(0.75), UInt32),
  tpot_p90 AggregateFunction(quantileTDigest(0.90), UInt32),
  tpot_p99 AggregateFunction(quantileTDigest(0.99), UInt32)
)
ENGINE = AggregatingMergeTree
PARTITION BY toDate(bucket_ts)
ORDER BY (tenant_id, bucket_ts, view_model_name, view_virtual_model_name)
TTL bucket_ts + INTERVAL 540 DAY
SETTINGS index_granularity = 8192;
```

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_gateway_model_metrics_1m
TO gateway_model_metrics_1m
AS
SELECT
  toStartOfMinute(ts) AS bucket_ts,
  tenant_id,

  model_name AS view_model_name,
  virtual_model_name AS view_virtual_model_name,
  user_email AS view_user_email,
  virtual_account_name AS view_virtual_account_name,
  NULL AS view_team_name,
  error_type AS view_error_type,

  countState() AS request_count,
  countIfState(status_code >= 400) AS failed_count,

  sumState(toUInt64(input_tokens)) AS input_tokens_sum,
  sumState(toUInt64(output_tokens)) AS output_tokens_sum,
  sumState(cost_usd) AS cost_usd_sum,

  quantileTDigestState(0.50)(latency_ms) AS latency_p50,
  quantileTDigestState(0.75)(latency_ms) AS latency_p75,
  quantileTDigestState(0.90)(latency_ms) AS latency_p90,
  quantileTDigestState(0.99)(latency_ms) AS latency_p99,

  quantileTDigestStateIf(0.50)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p50,
  quantileTDigestStateIf(0.75)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p75,
  quantileTDigestStateIf(0.90)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p90,
  quantileTDigestStateIf(0.99)(ttft_ms, ttft_ms IS NOT NULL) AS ttft_p99,

  quantileTDigestStateIf(0.50)(itl_ms, itl_ms IS NOT NULL) AS itl_p50,
  quantileTDigestStateIf(0.75)(itl_ms, itl_ms IS NOT NULL) AS itl_p75,
  quantileTDigestStateIf(0.90)(itl_ms, itl_ms IS NOT NULL) AS itl_p90,
  quantileTDigestStateIf(0.99)(itl_ms, itl_ms IS NOT NULL) AS itl_p99,

  quantileTDigestStateIf(0.50)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p50,
  quantileTDigestStateIf(0.75)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p75,
  quantileTDigestStateIf(0.90)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p90,
  quantileTDigestStateIf(0.99)(tpot_ms, tpot_ms IS NOT NULL) AS tpot_p99
FROM gateway_model_events
GROUP BY
  bucket_ts,
  tenant_id,
  view_model_name,
  view_virtual_model_name,
  view_user_email,
  view_virtual_account_name,
  view_team_name,
  view_error_type;
```

Note:
- Team-based grouping can be handled in v1 by raw-table query + `arrayJoin(team_names)`; optimize into dedicated MV in v1.1.

## Metrics Query API Contract (Initial)

### Endpoint
- `POST /api/v1/metrics/query`

### Request schema

```json
{
  "startTs": "2026-05-01T00:00:00.000Z",
  "endTs": "2026-05-01T03:00:00.000Z",
  "datasource": "modelMetrics",
  "type": "distribution",
  "aggregations": [
    { "type": "count", "column": "modelName" },
    { "type": "sum", "column": "inputTokens" },
    { "type": "sum", "column": "outputTokens" },
    { "type": "sum", "column": "costInUSD" },
    { "type": "p99", "column": "latencyMs" }
  ],
  "groupBy": ["modelName"],
  "filters": [
    { "fieldName": "virtualModelName", "operator": "IS_NULL", "value": true }
  ],
  "intervalInSeconds": 300
}
```

### Allowed aggregations
- `count`
- `sum`
- `p50`
- `p75`
- `p90`
- `p99`

### Allowed columns
- `costInUSD`
- `inputTokens`
- `outputTokens`
- `latencyMs`
- `timeToFirstTokenMs`
- `interTokenLatencyMs`
- `timePerOutputTokenLatencyMs`

### GroupBy fields
- `modelName`
- `virtualModelName`
- `userEmail`
- `virtualAccount`
- `team`
- `errorCode`
- `requestType`
- `providerModelName`
- `metadata.<key>`

### Filter operators
- String: `EQUAL`, `IN`, `NOT_IN`, `STRING_CONTAINS`, `STRING_STARTS_WITH`, `STRING_ENDS_WITH`, `IS_NULL`
- Number: `EQUAL`, `GREATER_THAN`, `LESS_THAN`, `GREATER_THAN_EQUAL`, `LESS_THAN_EQUAL`, `BETWEEN`, `IN`, `NOT_IN`
- Array: `ARRAY_HAS_ANY`, `ARRAY_HAS_NONE`

### Access rules
- Tenant admin: full tenant scope
- User: own records + team scope
- Virtual account token: own VA scope (plus team if configured)

## Execution Phases

## Phase 0 - Contract and Instrumentation Design
- [ ] Finalize canonical event fields and nullability.
- [ ] Decide stable `error_type` taxonomy.
- [ ] Define metadata key whitelist for group/filter.
- [ ] Add feature flag: `metrics.model.enabled`.

## Phase 1 - Ingestion Pipeline and ClickHouse Setup
- [ ] Add async emitter in `services/control-plane` request lifecycle.
- [ ] Implement batching (size/time threshold).
- [ ] Add retry with backoff and dead-letter strategy.
- [ ] Provision ClickHouse tables + MV in infra manifests.
- [ ] Add retention and storage policies.

## Phase 2 - Query Service and API
- [ ] Implement request validation and schema parsing.
- [ ] Build filter-to-SQL translator with strict allow-list.
- [ ] Implement distribution query path.
- [ ] Implement timeseries query path.
- [ ] Enforce tenant/user/VA access constraints.
- [ ] Add endpoint tests (happy path + abuse cases).

## Phase 3 - Admin UI (Model Metrics)
- [ ] Add top counters cards.
- [ ] Add time range selector and global filters.
- [ ] Add view-by selector tabs.
- [ ] Add charts: RPS, failures, latency, TTFT, ITL, TPOT, cost/tokens.
- [ ] Add percentile selector state handling.
- [ ] Add empty state and partial error handling.

## Phase 4 - Cost Correctness
- [ ] Implement model pricing source (public/private rates).
- [ ] Compute `cost_usd` deterministically at request close.
- [ ] Validate costs against wallet/budget records.
- [ ] Add audit log fields for pricing inputs used.

## Phase 5 - Reliability, Performance, and Rollout
- [ ] Load test ingest at target QPS.
- [ ] Load test query latency for 3h/24h/7d windows.
- [ ] Add SLO metrics and alerts.
- [ ] Run tenant-internal dogfood rollout.
- [ ] Roll out to selected tenants.
- [ ] GA rollout + docs.

## Testing Strategy
- Unit tests:
  - filter parser
  - SQL builder
  - percentile and aggregation mapping
  - access-scope resolver
- Integration tests:
  - ClickHouse query correctness against seeded fixtures
  - API response shape consistency
- End-to-end tests:
  - generate synthetic traffic and verify dashboard numbers
- Performance tests:
  - ingest throughput
  - query p95/p99 under concurrency

## SLO and Acceptance Criteria
- Ingest data loss rate: `< 0.1%`
- Query API availability: `>= 99.9%`
- Dashboard load p95 (3h range): `< 2.5s`
- Query p95:
  - 3h: `< 800ms`
  - 24h: `< 1.5s`
  - 7d: `< 2.0s`

## Open Questions
- Should team group-by be exact historical membership or current membership snapshot?
- For metadata group-by, what max key cardinality is acceptable per tenant?
- Retention policy per tenant tier (90/180/365 days)?
- Should raw metric export be enabled in MVP or post-MVP?

## Recommended Implementation Paths (Code Ownership)
- `services/control-plane/internal/observability/*`:
  - event build + emitter + retries
- `services/control-plane/internal/metricsquery/*`:
  - request schema + planner + SQL generation + execution
- `apps/*admin*/metrics/*`:
  - dashboard UI cards/charts/filters
- `infra/*clickhouse*`:
  - table/mv migrations and resource sizing

## Rollback Plan
- Keep metrics feature behind flag.
- Disable API routes via config if ClickHouse unavailable.
- Continue serving gateway traffic without metrics writes.

## Deliverables Checklist
- [ ] PRD-approved metrics contract
- [ ] ClickHouse schema migration files
- [ ] Metrics query API + tests
- [ ] Admin UI page + charts
- [ ] Runbook (ops, alerting, retention, backfill)
- [ ] Production rollout report

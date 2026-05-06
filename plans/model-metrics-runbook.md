# Model Metrics Runbook (MVP)

## Health checks
- `GET /ready` on control-plane
- Verify ClickHouse is reachable from control-plane (`CLICKHOUSE_DSN` configured)
- Verify `METRICS_MODEL_ENABLED=true`

## Feature flag rollback
1. Set `METRICS_MODEL_ENABLED=false`.
2. Restart control-plane.
3. Confirm inference traffic remains healthy and `/api/v1/metrics/query` may return no data.

## DLQ operations
- Failed metric events are appended to `./tmp/model-metrics-dlq.jsonl`.
- If DLQ grows, inspect ClickHouse connectivity, schema drift, and query logs.

## Common incidents
- `metrics store not configured`:
  - Check `CLICKHOUSE_DSN`.
  - Check network access and ClickHouse service state.
- high write failure rate:
  - Check disk and table TTL pressure.
  - Verify table schemas match migrations under `infra/clickhouse/migrations`.

## Backfill strategy
- Rebuild from `spend_logs` in PostgreSQL for missing windows.
- Recompute event fields:
  - `tenant_id`, `request_id`, `model_name`
  - `input_tokens`, `output_tokens`, `cost_usd`, `response_time_ms -> latency_ms`

## SLO checklist
- Ingest loss `< 0.1%`
- Query API availability `>= 99.9%`
- Dashboard p95 `< 2.5s` (3h window)

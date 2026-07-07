# ClickHouse Observability Migrations

## Apply Order
1. `001_gateway_model_events.sql`
2. `002_gateway_model_metrics_1m.sql`
3. `003_mv_gateway_model_metrics_1m.sql`
4. `004_observability_runtime_schema.sql`

## Rollback Notes
- Safest rollback order:
  1. `DROP TABLE IF EXISTS request_logs_ch`
  2. `DROP TABLE IF EXISTS request_spans_ch`
  3. `DROP TABLE IF EXISTS request_traces_ch`
  4. `DROP VIEW IF EXISTS mv_gateway_model_metrics_1m`
  5. `DROP TABLE IF EXISTS gateway_model_metrics_1m`
  6. `DROP TABLE IF EXISTS gateway_model_events`
- Use rollback only in non-production unless coordinated with analytics consumers.

## Assumptions
- DB timezone is UTC.
- TTL policies follow MVP plan:
  - Raw events: 90 days
  - 1-minute aggregates: 540 days
  - Request traces/spans: row-level `retention_days`; `0` means unlimited, positive values expire through ClickHouse TTL.
- Team aggregation in MVP query path can use raw table + `arrayJoin(team_names)`.
- Control-plane startup validates these tables and columns but does not create or alter ClickHouse schema.

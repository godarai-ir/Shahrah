# ClickHouse Model Metrics Migrations

## Apply Order
1. `001_gateway_model_events.sql`
2. `002_gateway_model_metrics_1m.sql`
3. `003_mv_gateway_model_metrics_1m.sql`

## Rollback Notes
- Safest rollback order:
  1. `DROP VIEW IF EXISTS mv_gateway_model_metrics_1m`
  2. `DROP TABLE IF EXISTS gateway_model_metrics_1m`
  3. `DROP TABLE IF EXISTS gateway_model_events`
- Use rollback only in non-production unless coordinated with analytics consumers.

## Assumptions
- DB timezone is UTC.
- TTL policies follow MVP plan:
  - Raw events: 90 days
  - 1-minute aggregates: 540 days
- Team aggregation in MVP query path can use raw table + `arrayJoin(team_names)`.

-- 012_drop_postgres_trace_tables

-- +goose Up

-- Migration: 012_drop_postgres_trace_tables
-- Description: Remove deprecated PostgreSQL trace tables after moving trace storage to ClickHouse.

DROP TABLE IF EXISTS request_spans;
DROP TABLE IF EXISTS request_traces;

-- +goose Down

-- Migration: 012_drop_postgres_trace_tables (DOWN)
-- Description: NOOP - Cannot recreate dropped trace tables without original schema.

-- This migration intentionally does nothing as the original trace table schema
-- is not available for recreation. The trace data has been migrated to ClickHouse.

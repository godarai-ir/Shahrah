-- Migration: 012_drop_postgres_trace_tables.sql
-- Description: Remove deprecated PostgreSQL trace tables after moving trace storage to ClickHouse.

DROP TABLE IF EXISTS request_spans;
DROP TABLE IF EXISTS request_traces;

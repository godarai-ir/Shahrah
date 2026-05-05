-- Migration: 011_request_traces.sql
-- Description: Persistent request trace and span storage for tenant-scoped observability.

CREATE TABLE IF NOT EXISTS request_traces (
    trace_id TEXT PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    request_id TEXT,
    model TEXT,
    route TEXT NOT NULL,
    status_code INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    duration_ms INT NOT NULL,
    error_code TEXT,
    sampled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS request_spans (
    span_id TEXT PRIMARY KEY,
    trace_id TEXT NOT NULL REFERENCES request_traces(trace_id) ON DELETE CASCADE,
    parent_span_id TEXT,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    duration_ms INT NOT NULL,
    status TEXT NOT NULL,
    attributes_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_traces_tenant_started
    ON request_traces(tenant_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_traces_status_started
    ON request_traces(status_code, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_traces_model_started
    ON request_traces(model, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_traces_route_started
    ON request_traces(route, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_spans_trace_started
    ON request_spans(trace_id, started_at ASC);


# Executable Prompt: Request Traces for Shahrah

Use this prompt with a coding agent to implement a production-grade `Request Traces` feature inGodarAI.

## Prompt

You are a senior Go backend engineer working in `d:\source\Shahrah`.
Implement a high-performance `Request Traces` feature for Shahrah's control plane and tenant admin.

### Hard Constraints

- Prioritize platform stability and throughput over trace completeness.
- End-to-end tracing overhead must stay under 5% p95 latency increase.
- Never block request-response path on trace persistence/export.
- If downstream trace systems fail, core gateway traffic must continue unaffected.

### Existing Context You Must Reuse

- Service: `services/control-plane` (Go, chi/net/http).
- Existing trace header behavior already exists in proxy/tunnel paths (`x-shahrah-trace-id`, `X-Trace-Id`, `traceparent` pass-through).
- Tenant admin app: `apps/tenant-admin`.
- Database migrations live under `db/main/migrations`.

### Product Scope (MVP include/defer/exclude)

- Include now:
  - Request-level trace capture for gateway requests.
  - Span graph for: auth -> rate-limit -> model-route -> provider-call -> response.
  - Adjustable head-based sampling + deterministic tenant/key hash sampling.
  - Async export pipeline (OTLP + local durable fallback queue).
  - Query APIs for trace list/detail in control plane.
  - Tenant-scoped UI in tenant-admin for trace search + detail waterfall.
- Defer:
  - Full-text body search across prompts/responses.
  - Cross-region distributed trace joins outside GodarAI boundary.
  - Advanced AI eval feedback loops.
- Exclude:
  - Synchronous trace writes in request path.
  - Always-on 100% sampling in production.

### Target Architecture

Implement this architecture:

1. Instrumentation Layer (in-process)
- Use OpenTelemetry Go SDK.
- Create spans in gateway request lifecycle with minimal attributes.
- Default to attribute allowlist; large payload attributes disabled by default.

2. Sampling Layer
- Head sampler with rules:
  - errors: 100%
  - latency above threshold (e.g. >1500ms): 100%
  - default traffic: 1-5% deterministic hash sample (`tenant_id + request_id`).
- Runtime-configurable per tenant and global fallback.

3. Async Trace Pipeline
- Request path only enqueues lightweight trace events to in-memory lock-efficient ring buffer.
- Dedicated worker batch-processes and exports spans.
- Export path:
  - Primary: OTLP collector (gRPC).
  - Secondary: local append-only WAL queue when exporter unavailable.
  - Replay worker drains WAL when exporter recovers.
- Backpressure policy:
  - bounded queue with drop-oldest for non-error spans.
  - never drop error/timeout spans unless memory safety threshold is exceeded.

4. Storage and Query
- Keep hot query index in ClickHouse (preferred) for low-latency analytical queries.
- If ClickHouse is not available in current stack, implement PostgreSQL MVP schema with strict retention and partitioning-ready design.
- TTL retention defaults:
  - high-cardinality spans: 7 days
  - error traces: 30 days
- Query APIs must be tenant-scoped and cursor-paginated.

5. UI/Operator Surface
- Tenant admin page: trace list filters (time range, status, model, route, trace id).
- Trace detail page: span timeline + key attributes + provider error context.
- Add deep link support from request logs to trace detail.

### Implementation Tasks (execute in order)

#### Phase 1: Foundation

1. Add tracing package in control-plane:
- `services/control-plane/internal/observability/tracing/`
  - `config.go`
  - `sampler.go`
  - `middleware.go`
  - `exporter.go`
  - `queue.go`
2. Wire middleware into HTTP server bootstrap.
3. Standardize trace context headers:
- read/propagate `traceparent`, `tracestate`, `x-shahrah-trace-id`.
- always set response header `x-shahrah-trace-id`.

#### Phase 2: Span Model + Pipeline

4. Implement canonical span names and attributes:
- `gateway.request`
- `auth.validate`
- `quota.check`
- `router.select_model`
- `provider.call`
- `response.write`
5. Add async bounded queue with metrics:
- queue depth
- dropped spans count by reason
- export batch size
- exporter latency/failure counters
6. Implement OTLP exporter config via env:
- `TRACE_OTLP_ENDPOINT`
- `TRACE_EXPORT_TIMEOUT_MS`
- `TRACE_BATCH_SIZE`
- `TRACE_QUEUE_CAPACITY`
- `TRACE_DEFAULT_SAMPLE_RATE`
- `TRACE_SLOW_REQUEST_MS`

#### Phase 3: Persistence + APIs

7. Add migration(s) for trace query store metadata (if Postgres MVP path):
- `request_traces`
- `request_spans`
- indexes for `(tenant_id, started_at desc)`, `(trace_id)`, `(status_code, started_at)`
8. Build read APIs:
- `GET /api/v1/traces`
- `GET /api/v1/traces/:trace_id`
- `GET /api/v1/traces/:trace_id/spans`
9. Enforce RBAC/tenant scoping in all trace queries.

#### Phase 4: Tenant Admin UI

10. Add traces pages in tenant-admin:
- list page with server-side filtering and pagination
- detail page with waterfall timeline
11. Add loading/error states optimized for large result sets.

#### Phase 5: Hardening

12. Add fail-open behavior tests (exporter down, queue full).
13. Add load tests validating <5% p95 regression at target throughput.
14. Add retention cleanup job and operational runbook docs.

### Data Model (Postgres MVP fallback)

Use this schema shape if ClickHouse is not integrated yet:

- `request_traces`
  - `trace_id` (pk, text)
  - `tenant_id` (text, indexed)
  - `request_id` (text)
  - `model` (text)
  - `route` (text)
  - `status_code` (int)
  - `started_at` (timestamptz, indexed)
  - `duration_ms` (int)
  - `error_code` (text, nullable)
  - `sampled` (bool)
- `request_spans`
  - `span_id` (pk, text)
  - `trace_id` (fk/index)
  - `parent_span_id` (text, nullable)
  - `name` (text)
  - `kind` (text)
  - `started_at` (timestamptz)
  - `duration_ms` (int)
  - `status` (text)
  - `attributes_json` (jsonb)

Add partial indexes for frequent filters and keep JSONB usage limited to non-filtered attributes.

### Performance Rules

- Do not perform DB/network trace writes in request goroutine.
- Avoid reflection-heavy attribute construction.
- Pre-allocate span attribute slices for hot paths.
- Cap per-request span count to prevent cardinality explosions.
- Sanitize and truncate large fields (prompt/response bodies off by default).

### Reliability Rules

- Exporter failure must only affect tracing subsystem.
- Trace queue must expose health metrics and drop policy.
- On sustained overload, auto-reduce sampling rate before dropping critical spans.

### Example Pseudocode (implement equivalent in Go)

```go
func TraceMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "gateway.request")
    start := time.Now()

    rw := newStatusCapturingWriter(w)
    defer func() {
      dur := time.Since(start)
      span.SetAttributes(
        attrTenantID(ctx),
        attrRoute(r.URL.Path),
        attrStatus(rw.StatusCode()),
        attribute.Int64("duration_ms", dur.Milliseconds()),
      )
      if rw.StatusCode() >= 500 || dur > slowThreshold {
        forceSample(span)
      }
      span.End()
    }()

    next.ServeHTTP(rw, r.WithContext(ctx))
  })
}
```

### Required Tests

- Unit:
  - sampler decisions
  - queue drop policy
  - header propagation
  - span attribute sanitizer
- Integration:
  - trace APIs tenant isolation
  - exporter outage fail-open
- Performance:
  - benchmark middleware overhead
  - load test at production-like concurrency

### Deliverables

1. Code changes in control-plane + tenant-admin.
2. DB migration(s) for trace metadata/querying.
3. Docs:
- feature README under `services/control-plane/docs/request-traces.md`
- operational runbook with sampling/backpressure tuning.
4. Final report including:
- measured latency overhead
- drop rates under stress
- recommended default sampling profile

### Acceptance Criteria

- p95 latency regression <5% at target load.
- No request failures caused by tracing subsystem outages.
- Engineers can query trace list/detail by tenant in <1.5s p95 for recent data window.
- Trace IDs are consistently propagated across gateway and tunnel boundaries.


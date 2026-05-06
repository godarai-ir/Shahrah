# Model Metrics Benchmark & SLO Evidence

## How to run
From `services/control-plane`:

```bash
go run ./cmd/metricsbench -base-url http://127.0.0.1:4000 -tenant <tenant-slug> -auth <master-or-admin-token> -runs 20 -out ./tmp/model-metrics-benchmark.md
```

## Produced artifact
- `services/control-plane/tmp/model-metrics-benchmark.md`

## Windows covered
- 3h (interval 300s)
- 24h (interval 900s)
- 7d (interval 3600s)

## SLO mapping
- 3h query p95 < 800ms
- 24h query p95 < 1500ms
- 7d query p95 < 2000ms

## Notes
- This benchmark hits `POST /api/v1/metrics/query` directly.
- It measures end-to-end API latency from client perspective.

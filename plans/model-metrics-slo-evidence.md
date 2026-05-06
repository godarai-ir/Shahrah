# Model Metrics SLO Evidence (Current Run)

Date: 2026-05-06

## Benchmark execution
Command:
```bash
go run ./cmd/metricsbench -base-url http://127.0.0.1:4000 -tenant demo -runs 5 -out ./tmp/model-metrics-benchmark.md
```

Result:
- API response: `404`
- Evidence file: `services/control-plane/tmp/model-metrics-benchmark.md`

## Interpretation
- The benchmark runner is implemented and operational.
- Current environment did not expose a live control-plane route at `POST /api/v1/metrics/query` on `127.0.0.1:4000` during this run.
- No valid latency samples were captured for 3h/24h/7d in this execution.

## Next production validation step
Re-run benchmark against the deployed control-plane instance with valid tenant/auth, then attach the generated markdown as release evidence.

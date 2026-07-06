# Shahrah Load Testing and Go Profiling

This folder contains repeatable load-test scenarios for the control-plane gateway and helper tooling for collecting Go runtime profiles while tests are running.

## 1. Install tools

Required:

- k6: https://grafana.com/docs/k6/latest/set-up/install-k6/
- Go toolchain, for `go tool pprof` and `go tool trace`
- Docker Compose, when testing the local stack

Optional but useful:

- Grafana/Prometheus for long-running metric dashboards
- `jq` for inspecting JSON responses

Windows quick install examples:

```powershell
winget source update
winget install --id GrafanaLabs.k6 -e --source winget
```

If winget still cannot find k6, use one of these alternatives:

```powershell
choco install k6
scoop install k6
winget install GoLang.Go
```

## 2. Enable profiling

The control-plane exposes `/debug/pprof/*` only when explicitly enabled.

For cost-safe local load testing, start the stack with the load-test override. This override routes static model traffic to Silk Road's mock provider and enables pprof for the control-plane:

```powershell
docker compose -f infra/compose/local.yml -f infra/compose/loadtest.override.yml up --build control-plane tunnel-client tunnel-server silk-road
```

The mock provider is disabled in normal compose/prod paths. It is only registered when `SILK_ROAD_ENABLE_MOCK_PROVIDER=true`, and the load-test override also mounts `loadtests/config/proxy_config.yaml` into the control-plane so static model routes point at provider `mock`.

You can tune mock behavior before starting compose:

```powershell
$env:SILK_ROAD_MOCK_LATENCY_MS="75"
$env:SILK_ROAD_MOCK_ERROR_RATE="0.01"
$env:SILK_ROAD_MOCK_ERROR_STATUS="429"
```

For a directly running service:

```powershell
Set-Location services/control-plane
$env:PROFILING_ENABLE_PPROF="true"
$env:PROFILING_MUTEX_PROFILE_FRACTION="5"
$env:PROFILING_BLOCK_PROFILE_RATE="10000"
go run ./cmd/server
```

Use profiling only in local, staging, or a tightly firewalled performance environment. The pprof endpoints expose process internals.

## 3. Run scenarios

Set the common variables:

```powershell
$env:BASE_URL="http://localhost:4000"
$env:API_KEY="sk-dev-master-key-12345"
$env:TENANT="tenant1"
$env:MODEL="gpt-4o-mini"
```

`TENANT` must be an existing tenant slug with enough wallet balance for protected `/v1/*` gateway calls. Prefer a real PAT or virtual-user token for staging tests; the local master key is only convenient for bootstrap/local verification. When the load-test override is active, requests use the mock provider instead of paid upstream providers.

If repeated local runs have exhausted the monthly gateway counter, reset the load-test tenant before smoke/ramp tests and clear the local Redis usage-counter cache:

```powershell
docker exec compose-postgres-1 psql -U postgres -d sharah_db -c "update tenant_usage_counters c set used = 0, updated_at = now() from tenants t where t.tenant_id = c.tenant_id and t.slug = 'tenant1' and c.feature_key = 'gateway.requests.monthly' and now() between c.window_start and c.window_end;"
docker exec compose-redis-1 sh -lc "redis-cli -a s4M6gOMp77ss --scan --pattern 'usage:tenant:*:feature:gateway.requests.monthly:*' | xargs -r redis-cli -a s4M6gOMp77ss DEL"
```

Smoke test:

```powershell
k6 run -e SCENARIO=smoke loadtests/k6/gateway-scenarios.js
```

Gradual ramp:

```powershell
k6 run -e SCENARIO=ramp -e TARGET_VUS=75 -e HOLD=15m loadtests/k6/gateway-scenarios.js
```

Arrival-rate mixed traffic:

```powershell
k6 run -e SCENARIO=mixed -e TARGET_RATE=100 -e HOLD=20m loadtests/k6/gateway-scenarios.js
```

Spike:

```powershell
k6 run -e SCENARIO=spike -e BASELINE_VUS=25 -e SPIKE_VUS=250 loadtests/k6/gateway-scenarios.js
```

Soak:

```powershell
k6 run -e SCENARIO=soak -e VUS=50 -e DURATION=4h loadtests/k6/gateway-scenarios.js
```

Response-cache repeated prompt pressure:

```powershell
k6 run -e SCENARIO=cache -e RATE=40 -e DURATION=30m loadtests/k6/gateway-scenarios.js
```

## 4. Profile during the test

Start k6 in one terminal, then collect profiles from another:

```powershell
.\loadtests\scripts\collect-go-profiles.ps1 `
  -BaseUrl "http://localhost:4000" `
  -OutDir "loadtests/results/profiles" `
  -IntervalSeconds 60 `
  -CpuSeconds 30 `
  -Samples 10
```

Inspect collected profiles:

```powershell
go tool pprof -http=:0 loadtests/results/profiles/<run>/001-cpu.pb.gz
go tool pprof -http=:0 loadtests/results/profiles/<run>/001-heap.pb.gz
go tool pprof -http=:0 loadtests/results/profiles/<run>/001-mutex.pb.gz
go tool pprof -http=:0 loadtests/results/profiles/<run>/001-block.pb.gz
go tool trace loadtests/results/profiles/<run>/001-trace.out
```

## 5. What to watch

CPU profile:

- Hot handlers, JSON encode/decode, provider transport, DB calls, cache paths.
- Compare CPU under `smoke`, `mixed`, and `spike`; look for new hot paths that only appear under concurrency.

Heap profile:

- Allocation-heavy request paths, retained buffers, response cache growth, trace/log payload retention.
- Use pprof `alloc_space` for churn and `inuse_space` for retained memory.

Goroutine count:

- Prometheus `/metrics`: `go_goroutines`.
- Text snapshots: `*-goroutine.txt`.
- During a steady soak, count should stabilize. A rising line usually means leaked workers, blocked upstream calls, or stuck DB/cache operations.

GC pause time:

- Prometheus `/metrics`: check Go GC duration/runtime metrics such as `go_gc_duration_seconds` and `go_memstats_pause_total_ns`, depending on collector output.
- Correlate p95/p99 k6 latency spikes with GC pause increases and heap growth.

Mutex contention:

- `go tool pprof -http=:0 <run>/001-mutex.pb.gz`.
- Focus on contended locks in shared caches, trace queues, billing/metrics queues, and tenant/auth lookup paths.

Blocking profile:

- `go tool pprof -http=:0 <run>/001-block.pb.gz`.
- Watch channel sends/receives, queue backpressure, DB pool waits, and transport waits.

## 6. Suggested test matrix

Run the matrix against local first with the mock provider, then staging with production-like database and either mock latency or carefully capped real-provider traffic:

| Scenario | Purpose | First target |
| --- | --- | --- |
| `smoke` | Validate credentials, tenant routing, and baseline success | 2 VUs for 1m |
| `ramp` | Find saturation curve | 25, 50, 100 VUs |
| `mixed` | Production-like endpoint mix | 50-150 req/s |
| `spike` | Burst resilience and queue behavior | 5x normal load |
| `soak` | Leaks, GC stability, goroutine stability | 2-4 hours |
| `cache` | Repeated prompt/cache contention | 30-100 req/s |

Keep one variable changing at a time: load level, provider latency, cache enabled, tracing sample rate, metrics enabled, or DB pool size.

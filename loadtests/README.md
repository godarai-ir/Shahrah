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

The mock provider is disabled in normal compose/prod paths. It is only registered when `SILK_ROAD_ENABLE_MOCK_PROVIDER=true`, and each request must also match both `SILK_ROAD_MOCK_ALLOWED_TENANT_IDS` and `SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS`. Missing allowlists fail closed with HTTP 403. The load-test override allows all local tenants but only the static deployment ID `loadtest-mock-1`; production must use an exact tenant ID instead of `*`.

The mock provider is always treated as `external_billed` by the control-plane, so mock traffic does not reserve or debit the tenant wallet. It still emits normal request, trace, token, latency, and provider metadata for gateway load testing.

### Production provisioning sequence

Use a dedicated production tenant. While signed in to that tenant's admin UI,
create the mock deployment from the browser console so no admin token is copied
outside the existing session:

```javascript
const result = await fetch("/api/models", {
  method: "POST",
  credentials: "include",
  headers: { "content-type": "application/json" },
  body: JSON.stringify({
    model_name: "loadtest/mock",
    provider: "mock",
    model: "mock/chat",
    rpm: 100000,
    tpm: 10000000,
    weight: 1,
    timeout: 60,
  }),
}).then(async (response) => ({
  status: response.status,
  body: await response.json(),
}));
console.log(result);
```

The tenant-admin BFF only permits missing upstream credentials for the exact
internal provider slug `mock`; external providers still require both fields.
Keep the returned tenant UUID and deployment UUID; they are the exact values
committed to the production GitOps allowlists. Create the VAT with an empty model
allowlist in this dedicated tenant. Legacy mock deployments are intentionally
not added to the public provider-account catalogue; an empty VAT allowlist in a
tenant without provider-account policy enforcement uses the tenant's available
models.

Do not enable the production mock flag until the fail-closed Silk Road image is
deployed. After GitOps contains the exact allowlists, run a single smoke request
with `EXPECT_MOCK_PROVIDER=true` and require
`X-Shahrah-Mock-Provider: true` before starting any ramp, spike, or soak test.

For production, set every safety-sensitive value explicitly in the same terminal.
Keep the VAT local to that terminal and do not paste it into logs, Git, or chat:

```powershell
$env:PRODUCTION_LOADTEST="true"
$env:BASE_URL="https://<tenant-slug>.godarai.ir"
$env:API_KEY="<dedicated-load-test-vat>"
$env:TENANT="<tenant-slug>"
$env:MODEL="mock:loadtest:chat"
$env:EMBEDDING_MODEL="mock:loadtest:chat"
$env:EXPECT_MOCK_PROVIDER="true"

k6 run -e SCENARIO=mock-smoke loadtests/k6/gateway-scenarios.js
```

The `mock-smoke` scenario sends exactly one chat request. Production URLs and
`PRODUCTION_LOADTEST=true` enable a fail-fast guard: k6 refuses to start unless
the dedicated VAT, tenant slug, exact mock model, HTTPS base URL, and mock-marker
expectation are all present. A missing `X-Shahrah-Mock-Provider: true` fails the
test through the `gateway_success_rate` threshold.

### Troubleshoot a stale production load-test script

Production must use the canonical model FQN `mock:loadtest:chat`. If the safety
check instead reports either of these messages, k6 is running an older copy of
`gateway-scenarios.js`:

```text
MODEL must equal loadtest/mock
EMBEDDING_MODEL must be unset or equal loadtest/mock
```

Do not change the production model to `loadtest/mock`; that alias is only valid
for the local compose load-test deployment. Verify that the script contains the
production guard before continuing:

```powershell
Select-String `
  -Path loadtests/k6/gateway-scenarios.js `
  -Pattern "productionMockModel = 'mock:loadtest:chat'"
```

Update the checkout before running k6. If the current workspace contains
uncommitted work, preserve it and create a clean worktree instead of pulling,
resetting, or overwriting those changes:

```powershell
git fetch origin
git worktree add ../shahrah-loadtest origin/main
Set-Location ../shahrah-loadtest
```

Export `PRODUCTION_LOADTEST`, `BASE_URL`, `API_KEY`, `TENANT`, `MODEL`,
`EMBEDDING_MODEL`, and `EXPECT_MOCK_PROVIDER` again in that same PowerShell
session. Then run `mock-smoke` from the clean checkout before starting a ramp.

Only after `mock-smoke` passes should you run the regular scenarios in the same
terminal. Clear the VAT from the terminal when the test window ends:

```powershell
k6 run -e SCENARIO=smoke loadtests/k6/gateway-scenarios.js
k6 run -e SCENARIO=ramp -e TARGET_VUS=25 -e HOLD=5m loadtests/k6/gateway-scenarios.js

Remove-Item Env:API_KEY
```

### Run protected production tests from GitHub Actions

The `Production k6 Load Test` workflow provides a manually dispatched CI runner
for `mock-smoke`, `smoke`, `ramp`, `soak`, and `spike`. It runs only from
`main`, serializes production load tests with a workflow-level concurrency
group, uses the exact production mock guard values, and uploads the warning log
and JSON summary as GitHub Actions artifacts.

Create a GitHub Environment named `production-loadtest`, restrict it to the
custom deployment branch policy `main`, and add the dedicated load-test VAT as
the environment secret `LOADTEST_VAT`. Require an environment reviewer when the
repository's GitHub plan supports that protection rule. Do not use a provider
credential, master key, or ordinary tenant token for this secret.

Run the workflow from **Actions > Production k6 Load Test > Run workflow** and
explicitly check the production confirmation input. Start with `mock-smoke`.
Only approve a higher-load scenario after that run passes and confirms the
mock-provider marker. The workflow exposes bounded choices for VUs and duration
so an accidental dispatch cannot request an unreviewed load level.

The GitHub-hosted runner's source network differs from a developer workstation.
Treat CI timeout or connection failures as load-generator network evidence; a
request that never reaches the production ingress cannot create a tenant trace.
Download the workflow artifact to inspect `k6.log` and `summary.json`.

You can tune mock behavior before starting compose:

```powershell
$env:SILK_ROAD_MOCK_LATENCY_MS="75"
$env:SILK_ROAD_MOCK_ERROR_RATE="0.01"
$env:SILK_ROAD_MOCK_ERROR_STATUS="429"
$env:SILK_ROAD_MOCK_ALLOWED_TENANT_IDS="*"
$env:SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS="loadtest-mock-1"
$env:EXPECT_MOCK_PROVIDER="true"
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

The cache scenario now sends a valid JSON `x-godarai-cache-config` header and defaults to `CACHE_TYPE=simple` for deterministic exact-match hits. Override `CACHE_TYPE=smart` only when the smart-cache embedder is configured and reachable.

### Rate-limit rule contract test

Use the dedicated scenario for an exact tenant rule such as "10 requests per minute". Configure the rule as enabled, hard limit (`soft_limit=false`), `dimension=requests`, `max=10`, a one-minute window, and preferably `retry_after_header=true`. A rule with no conditions applies tenant-wide; tenant isolation is already part of the runtime counter key.

```powershell
$env:BASE_URL="http://localhost:4000"
$env:API_KEY="sk-dev-master-key-12345"
$env:TENANT="tenant1"
$env:MODEL="gpt-4o-mini"
$env:RULE_LIMIT="10"
$env:WINDOW_SECONDS="60"

k6 run --summary-export loadtests/results/rate-limit-tenant1.json loadtests/k6/rate-limit-rule.js
```

The test first sends one preflight request and waits for its reported reset time, so the measured burst starts with a clean window. It then sends 20 concurrent requests by default and requires exactly 10 successful responses and 10 rate-limit responses. After the 60-second window plus a small grace period, it sends three recovery requests and requires all three to succeed.

Use a dedicated load-test tenant (or a rule scoped to a dedicated virtual account) when possible. Any unrelated request matching the same tenant-wide rule correctly consumes one of the ten slots and makes the exact-count threshold fail; that is useful evidence, but it can make a shared staging environment nondeterministic.

Important environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `RULE_LIMIT` | `10` | Exact configured request limit |
| `WINDOW_SECONDS` | `60` | Exact configured window length |
| `BURST_REQUESTS` | `RULE_LIMIT * 2` | Requests sent in the initial burst; must exceed the limit |
| `BURST_VUS` | up to `20` | Concurrent workers for the burst |
| `RECOVERY_REQUESTS` | up to `3` | Requests that must pass after reset |
| `GRACE_SECONDS` | `2` | Clock/network tolerance around reset |
| `ALIGN_WINDOW` | `true` | Preflight and wait for a clean window; disable only after explicitly clearing counters |
| `EXPECT_RETRY_AFTER` | `true` | Require a valid `Retry-After` on rejected requests |

The run fails if the exact 2xx/429 counts differ, a response is neither 2xx nor 429, the OpenAI-compatible error type is not `rate_limit_error`, the rate-limit headers are inconsistent, or traffic does not recover after reset. Expected 429 responses are deliberately treated as successful HTTP observations by k6; correctness is enforced through the custom counters and thresholds instead of the generic `http_req_failed` metric.

### Budget-limit rule concurrency test

Budget enforcement is cost-based rather than request-count-based. Configure a fresh hard budget rule with `block_on_budget_exceed=true`, choose the required scope/conditions, and set `BUDGET_LIMIT_IRR` to the exact `limit_to` value. For a quick local run with the mock provider, a tenant or virtual-account rule around `400 IRR` normally allows several calls before becoming exhausted.

Current runtime enforcement compares `limit_to` with `TotalCostIRR`. The tenant-admin copy currently labels this field as dollars; until that product contract is reconciled, use the numeric IRR value actually enforced by the gateway when running this test.

```powershell
$env:BASE_URL="http://localhost:4000"
$env:API_KEY="<dedicated PAT or virtual-account token>"
$env:TENANT="budget-loadtest-tenant"
$env:MODEL="gpt-4o-mini"
$env:BUDGET_RULE_ID="loadtest-daily-budget-v1"
$env:BUDGET_LIMIT_IRR="400"

k6 run --summary-export loadtests/results/budget-limit.json loadtests/k6/budget-limit-rule.js
```

The burst sends 30 concurrent requests by default. Because the gateway atomically reserves estimated cost and later commits actual provider usage, the exact number of 2xx responses depends on model pricing and returned usage. The test therefore enforces stronger cost invariants instead of a fragile exact response count:

- At least one request must succeed and expose a valid `X-Godarai-Cost-IRR`.
- At least one request must be rejected with HTTP `402` and `error.type=budget_exceeded`.
- The sum of successful response costs must not exceed `BUDGET_LIMIT_IRR` (plus rounding tolerance).
- After all concurrent requests settle, teardown sends verification calls and requires all of them to remain blocked.
- HTTP 429 is treated as interfering Rate Limit traffic and fails the run.

Important environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `BUDGET_LIMIT_IRR` | `400` | Must equal the configured rule's `limit_to` |
| `BUDGET_RULE_ID` | `budget-loadtest` | Metric tag and unique prompt namespace for the tested rule |
| `BURST_REQUESTS` | `30` | Concurrent pressure used to exhaust the budget |
| `BURST_VUS` | up to `30` | Concurrent workers |
| `SATURATION_REQUESTS` | `3` | Verification calls after the burst has fully settled |
| `MIN_ALLOWED` | `1` | Minimum expected successful calls in a clean budget window |
| `MIN_REJECTED` | `1` | Minimum expected budget rejections during the burst |
| `MAX_TOKENS` | `8` | Controls estimated output cost per request |
| `COST_TOLERANCE_IRR` | `1` | Tolerance for summing two-decimal response cost headers |
| `METADATA_JSON` | unset | Optional JSON metadata used by metadata-scoped rules |

Use a new rule ID or explicitly clear only that rule's local Redis budget window before each repeat. Budget windows are aligned to UTC day/week/month boundaries, so a generic load test should not wait for reset. Test rollover separately near the real UTC boundary, or with an integration test that controls the clock. Also use a dedicated tenant or virtual account and avoid overlapping Rate Limit rules so 429 responses do not obscure budget behavior.

Keep the test tenant's wallet balance materially higher than the rule limit and disable other matching hard budget rules. Otherwise a valid 402 may come from wallet exhaustion or a different rule, which the public error envelope intentionally does not distinguish.

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

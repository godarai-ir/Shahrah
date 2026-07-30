# Anthropic Claude 5 production-readiness plan

**Scope:** Sonnet 5, Haiku 4.5, Opus 5, Fable 5  
**Current gate:** `partially supported`  
**Rule:** no phase may be marked complete without its named evidence artifact.

## Phase 0 — merge-local correctness

Owner: gateway + catalog maintainers

- [x] Add exact canonical IDs and explicit runtime-profile version 2.
- [x] Remove Claude 5 assistant-prefill claims.
- [x] Encode thinking modes and Opus 5 disabled-effort ceiling.
- [x] Encode model-specific server-tool families and beta headers.
- [x] Encode 1M/200K input and 128K/64K output limits.
- [x] Add Fable retention/ZDR metadata.
- [x] Add effective-dated Sonnet 5 launch and standard pricing.
- [x] Add Batch/cache/data-residency/fast/tool pricing dimensions.
- [x] Make imported runtime-profile versions immutable.
- [x] Surface the upstream served model in compatible responses.
- [x] Rerun all changed Go tests with constrained parallelism.
- [x] Run model-discovery build and provider YAML dry-run.

Exit evidence:

- CI links for contracts, control-plane, Silk Road, and model-discovery.
- Catalog snapshot showing four version-2 profiles and non-overlapping
  Sonnet 5 pricing periods.
- `git diff --check` clean.

Rollback:

- Disable the four model deployments.
- Retain catalog/profile history; do not mutate an already imported profile.

## Phase 1 — credentialed provider conformance

Owner: provider integrations

Prerequisites:

- Staging GodarAI virtual key.
- Anthropic API-key provider account.
- WIF provider account and complete identity configuration.
- Separate Anthropic Admin API key and optional workspace ID.
- Approved spend ceiling.

Run from `services/control-plane`:

```powershell
$env:GODARAI_BASE_URL = "https://<staging>"
$env:GODARAI_VIRTUAL_KEY = "<secret>"
$env:GODARAI_ANTHROPIC_PROVIDER_ACCOUNT = "<account>"
$env:GODARAI_ANTHROPIC_AUTH_LABEL = "api_key"
$env:ANTHROPIC_ADMIN_KEY = "<admin-secret>"
$env:ANTHROPIC_ADMIN_WORKSPACE_ID = "<workspace>"

go run ./cmd/anthropic-readiness `
  -run-resources `
  -test-300k `
  -report anthropic-api-key-readiness.json
```

Repeat for WIF with a separate provider account and
`GODARAI_ANTHROPIC_AUTH_LABEL=wif`.

Mandatory assertions:

- [ ] Models API returns every requested identity and published limits.
- [ ] Sync and streaming response model equals the actual served model.
- [ ] Count Tokens, structured output, thinking, tool calls, and stream
  termination pass on every model.
- [ ] Negative matrix returns expected 4xx for sampling, prefill, thinking,
  web-fetch, speed, and 300K incompatibilities.
- [ ] Fable refusal is surfaced as HTTP 200 with `stop_reason=refusal`.
- [ ] Fable fallback reports the fallback model and correct billing model.
- [ ] API-key and WIF results are behaviorally equivalent.
- [ ] Admin Rate Limits snapshot has zero blocked checks.

Exit evidence:

- Two sanitized readiness reports with `failed=0` and `blocked=0`.
- Request-ID SHA-256 digests retained for correlation; no raw
  prompts/secrets/organization metadata/opaque IDs.
- Signed review of identity, auth, and negative matrices.

Rollback:

- Disable affected provider accounts or model deployments independently.

## Phase 2 — billing reconciliation

Owner: billing + FinOps

- [ ] Reconcile standard input/output/cache usage for all four models.
- [ ] Reconcile Sonnet 5 before and after 2026-09-01 in a clock-controlled
  staging replay.
- [ ] Reconcile Batch's 50% rates, including cache dimensions.
- [ ] Reconcile US-only 1.1 multiplier.
- [ ] Reconcile Opus 5 fast mode and fast-cache/data-residency stacking.
- [ ] Verify fallback charges use the served model.
- [ ] Verify failed web search is unbilled and successful search is $0.01.
- [ ] Verify code-execution free/billable evidence and fail-closed behavior.
- [ ] Verify cancelled/expired/failed Batch settlement.
- [ ] Compare GodarAI ledger totals to Anthropic Cost/Usage API and invoice
  rounding over at least one billing day.

Exit evidence:

- Redacted reconciliation workbook/report with per-dimension deltas.
- Zero unexplained material variance.
- Billing owner sign-off.

Rollback:

- Disable the affected priced capability (fast, geo, Batch, or tool) without
  disabling ordinary Messages traffic.

## Phase 3 — resilience, limits, and load

Owner: SRE

- [ ] Run the 500 MB bounded-memory upload test.
- [ ] Run long-context tests near every model limit and verify local/upstream
  failure classification.
- [ ] Exercise 429 acceleration limits, retry-after handling, jitter, and
  retry budgets.
- [ ] Exercise 529 overload, connection resets, mid-stream errors, and client
  cancellation.
- [ ] Confirm no automatic retry duplicates non-idempotent resources.
- [ ] Confirm Batch polling is bounded and cancellable.
- [ ] Confirm workspace limits inherit/override organization limits as
  reported by the Admin API.
- [ ] Confirm rate-limit buckets are model-group specific.
- [ ] Run sustained concurrency and backpressure tests at the approved
  staging ceiling.

Exit evidence:

- Load report with CPU, memory, goroutine, queue, and latency percentiles.
- Fault-injection report with duplicate-resource count equal to zero.
- SRE sign-off.

Rollback:

- Per-model concurrency/rate reduction and feature kill switches.

## Phase 4 — security, privacy, and observability

Owner: security + privacy + observability

- [ ] Verify API keys and WIF assertions never enter logs/traces/reports.
- [ ] Verify prompt/response and file redaction policies.
- [ ] Enforce Fable's minimum 30-day retention and prevent ZDR placement.
- [ ] Verify other models' ZDR workspace behavior where contracted.
- [ ] Verify US-only routing and billing evidence.
- [ ] Run prompt/tool abuse cases, path traversal, untrusted web results,
  malicious files, and MCP/tool permission checks.
- [ ] Dashboard requested model, served model, profile/catalog revision,
  auth mode, speed, service tier, geo, cache, Batch, tool use, status class,
  retries, and billing reconciliation state.
- [ ] Alert on identity mismatch, stale pricing/profile snapshot, unpriced
  usage, blocked reconciliation, refusal/fallback surge, 429/529 surge, and
  profile/version drift.
- [ ] Deploy a weekly first-party source and Models API diff job that records
  response hashes, capability/limit/pricing changes, and opens a reviewed
  catalog/profile update rather than mutating a published version.
- [ ] Deploy a daily deprecation-window check with 180/90/30-day alerts and a
  replacement/rollback owner for each enabled model.

Exit evidence:

- Security/privacy review.
- Dashboard and alert links with synthetic firing proof.
- Data-retention configuration export.

Rollback:

- Revoke provider credentials, disable tools/files, and isolate affected
  tenants/models.

## Phase 5 — canary and production release

Owner: release manager

- [ ] Human review of code, catalog, evidence, runbooks, and rollback.
- [ ] Deploy to staging and pass all prior gates.
- [ ] Canary API-key traffic for allowlisted tenants.
- [ ] Canary WIF traffic separately.
- [ ] Expand one model at a time: Haiku, Sonnet, Opus, then Fable.
- [ ] Hold each step for the agreed error/billing/latency observation window.
- [ ] Verify no P0/P1 open issues and no blocked readiness checks.
- [ ] Record final release approval.

Production exit criteria:

- `failed=0`, `blocked=0` for both auth modes.
- Billing reconciliation signed.
- Actual Admin API limits captured and configured.
- Load/security/privacy/observability gates signed.
- Rollback drill completed.

Only after every production exit criterion is met may the status be changed
from `partially supported` to `production-ready`.

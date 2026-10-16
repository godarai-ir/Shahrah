# Phase 3 Design: Credit Enforcement, Budget Rules, Fleet Cutover

Status: **3a and 3c implemented**; 3d–3e pending. 3b (downgrade) was dropped (P5).
Source: `plans/credit-pricing-execution-prompt.md` §6 Phase 3 and Phase 4.
Builds on Phase 2 (`plans/credit-pricing/phase-2-design.md`, migrations 066–069).

## Product-owner decisions (2026-10-16)

These replace parts of the prompt. They were given in answer to the §8 stop-and-ask list.

| # | Decision | Effect on this phase |
| --- | --- | --- |
| P1 | **No outbound notification channel**, neither email nor in-panel. | FR-7 alerts are not delivered to anyone. `on_exceed = alert` allows the request and records a deduplicated structured log line plus a Prometheus counter. The 30-day price notices from Phase 2 keep relying on the public changelog only. Phase 4 item 1 (direct message with a calculator) is dropped. |
| P2 | **Every tenant is converted; nobody opts in or out.** | The cutover tool has no tenant-approval step. Support-admin runs it, per tenant or for the whole fleet. §8.3 (tenants that never approve) no longer applies. |
| P3 | **No pilot.** | Gate 3's "3 pilot tenants for one billing period" is waived. It is replaced by a staging run plus a fleet dry run with no errors (see "Gate 3 as amended"). Phases 3 and 4 merge: the fleet cutover is part of this phase. |
| P4 | **Every tenant moves to the same-slug credit plan at cutover** (option C). The owner confirmed that no current subscription is sensitive, so the simplest mechanics were chosen (C2). | The current subscription is canceled and a new `active` subscription starts at once on the published credit version of the same plan slug and interval. The plan grant worker grants the first slice right away. That first period is free and *is* the compensation (Phase 4 item 3), so no separate compensation lot is granted. The new price is paid from the next renewal through the existing payment flow. |
| P6 | **Side usage is priced per call** (2026-10-17): **1 credit per guardrail judge call, 0 per semantic-cache embedding.** Neither call reports a provider cost, so the `500 × USD` formula cannot price them. | Constants `credits.GuardrailJudgeCallMC` and `credits.SemanticCacheEmbeddingMC`, published in `/v1/rate-card` under `side_usage` and part of its ETag. With a price of 0, semantic-cache embeddings write no spend log and no ledger entry. |
| P5 | **No `downgrade`** (2026-10-17). Budget rules offer only `block` and `alert`. | FR-8 (route to a cheaper model), the substitution headers, and the daily downgrade quota are not built; PR 3b is dropped. FR-10's tenant default `on_exceed` is dropped too: with only `alert` left as an alternative, a pool default of `alert` would mean unlimited debt, so an exhausted balance always blocks unless overage is on. |

## Current state (verified on `feat/credits-phase-2-to-main`)

- `BudgetService.ReserveBudget` (`internal/billing/budget_service.go`) reserves the IRR estimate in a monthly tenant window `budget:tenant:{id}:window:{YYYYMM}` and in one window per matching rule. Only rules with `block_on_budget_exceed = true` are enforced. The others are ignored.
- The tenant window is seeded with `allowance = wallet.balance_irr + settled_irr(window)`. Seeding happens on first use and again after every wallet change (`syncCurrentBudgetWindowAllowance`). `WalletSettlementProcessor` turns committed window spend into wallet debits, and skips credit tenants since Phase 2.
- **BYOK (`external_billed`) skips reservation entirely** (`request_pipeline.go` `Reserve`/`Settle`), so BYOK requests count against no budget today.
- Credits per request are already computed (`credits.FromBreakdown`, Phase 1). The proxy stamps them on spend logs (`CreditShadowSpendLogger`), and the ledger settles them asynchronously (Phase 2).
- **Credit tenants are not enforced in credits.** Their IRR window is seeded from a wallet that is no longer debited, so it never runs out. This is why no tenant may be switched before Phase 3 lands.
- Side usage: `guardrail` and `semantic_cache` exist as product labels only. Judge and embedding calls produce no spend logs.
- `err.Error()` still leaks into client responses at `anthropic_native_handler.go:285`, `gemini_native_handler.go:138`, and `tokenbazaar_sync.go:84`.

## Design

### 1. Billing unit on the hot path

- `platform.Tenant` gains `BillingUnit string` (`irr` | `credit`). The runtime profile compiler fills it, so `BudgetService` reads it from the cached profile (`runtimeProfiles.GetSnapshot`) with no extra DB query.
- Switching the unit recompiles the tenant profile, using the existing "tenant changed" event. A request that raced the switch uses the old unit. That request is still settled correctly, because ledger settlement reads `tenants.billing_unit` at settle time.

### 2. Credit tenant pool window

| | IRR tenant (unchanged) | Credit tenant |
| --- | --- | --- |
| Window key | `budget:tenant:{id}:window:{YYYYMM}` | `credit:tenant:{id}:window:{YYYYMM}` |
| Amount unit | IRR (float) | millicredits (whole numbers in the float field) |
| Allowance | `balance_irr + settled_irr(window)` | `available_mc − debt_mc + overage_headroom_mc + settled_mc(window)` |
| Settled by | `WalletSettlementProcessor` (wallet debit) | nothing: the ledger settles from spend logs. The window is enforcement only. |

- The `credit:` prefix keeps credit windows out of `WalletSettlementProcessor`, which only handles `budget:tenant:`. IRR and credit amounts can never mix (invariant 7).
- The Redis Lua scripts are unit-agnostic, so they are reused as they are. Their `*_irr` hash field names stay; a comment documents that the unit is fixed by the key prefix. Renaming them would touch every window and gains nothing.
- `settled_mc(window)` comes from a new table `credit_budget_window_settlements (tenant_id, window_id, settled_mc)`. The ledger settlement transaction increments it, for the window of each spend log's `created_at`. Then `remaining = balance − reserved − (committed − settled)`: spend that is committed but not yet settled is never counted twice, and is never forgotten either.
- **Overage headroom:** 0 when overage is off. When it is on, it is `cap − debt` if a cap is set. With no cap it is a fixed safety ceiling (`BILLING_CREDIT_OVERAGE_DEFAULT_CEILING_CREDITS`, default 100,000 credits) rather than unlimited.
- **Re-seed triggers:**
  - after any grant, purchase, adjustment, expiry, or conversion, for that tenant;
  - after each settlement batch, for the tenants it touched;
  - on `window_unavailable`.
  The re-seed is the same idempotent `SeedBudgetWindow` call the IRR path uses.
- **Estimate:** the request's estimated breakdown goes through `credits.FromBreakdown` with the current rate card, rounded up once. The commit uses the actual breakdown in the same way, so the window and the ledger see the same millicredits.

### 3. Budget rules in credits (FR-5/6/7)

Migration `070_credit_budget_rules.sql` adds the following to `budget_limit_rules`:

| Column | Type | Default |
| --- | --- | --- |
| `limit_mc` | `bigint` | NULL |
| `on_exceed` | `varchar` | `block` |
| `alert_thresholds` | `jsonb` | `[80,100]` |

It also adds a check constraint on `on_exceed IN ('block','alert')`.

- Backfill: `block_on_budget_exceed = true → block`, `false → alert`. This keeps behavior: rules that did not block before still do not block.
- **API contract:**
  - `limit` is a decimal string in credits (two decimals) and maps to `limit_mc`; `on_exceed` and `alert_thresholds` are new.
  - `limit_to` (IRR) and `block_on_budget_exceed` keep working for one release.
  - A request that sets both `block_on_budget_exceed` and `on_exceed` inconsistently gets a 400 `budget_rule_inconsistent_exceed`.
- **Which limit applies:** credit tenants enforce `limit_mc`. IRR tenants enforce `limit_to` as today. The cutover fills `limit_mc` from `limit_to` (§6).
- **Rule windows:**
  - Credit-tenant rule windows use the prefix `credit_policy:`.
  - `alert` rules get a window too, so spend is counted and thresholds can fire, but reservation always succeeds.
  - Under BYOK, rule windows reserve the full credit equivalent, while the tenant pool reserves only the fee (§5).

### 4. Exceed behavior

- **`block`:**
  - Response: HTTP 402 `{"error":{"type":"budget_exceeded","code":"budget_exceeded","scope":"<scope>"}}`, shaped for each surface: OpenAI-compatible, Anthropic-native, Gemini-native, TokenBazaar.
  - `scope` is the rule's applies-per (`tenant`, `user`, `team`, `org_unit`, `model`, `virtualaccount`), or `tenant_balance` for the pool.
  - The response never includes amounts or internal ids. The three `err.Error()` leaks are fixed in the same PR.
- **`alert`:**
  - The request is allowed.
  - Crossing a threshold emits one log line `budget_alert` with rule id, scope, window, and threshold, plus `godarai_budget_alerts_total{threshold}`.
  - Dedupe uses a Redis `SET NX` on `(rule, scope value, window, threshold)` with the window's TTL. Nothing is sent to users (P1).
- **Tenant balance:** an exhausted balance always blocks (scope `tenant_balance`), unless overage is on and its headroom is left (P5).

### 5. BYOK (FR-11)

- BYOK requests are now reserved. Rule windows reserve the full equivalent. The pool reserves `ceil10(equivalent × fee%)`, with the fee from entitlement or override (default 5%, the same resolver as Phase 2 settlement).
- The ledger `consume` already stores the fee as `amount_mc` (Phase 2). This phase adds `credit_equivalent_mc` to the entry metadata, for reports.

### 6. Cache hits and side usage

- An exact or semantic cache hit commits 0 for the response, and `credits_avoided_mc` is recorded. This already works for spend logs; this phase checks that reservations are released on hits.
- **Judge calls and semantic-cache embeddings** (lookup and write, hit or miss):
  - Each one writes its own spend log with `product = guardrail` or `semantic_cache`, attributed to the originating request's user, team, org unit, and trace.
  - Each one commits post hoc to the pool and to matching rule windows, using reserve-and-commit in one call with no admission check. The possible small overshoot is documented.
  - If the pool is already exhausted, the semantic-cache lookup is skipped, and the request goes through normal `on_exceed`.
- The tenant-admin usage breakdown shows `guardrail` and `semantic_cache` as separate products.

### 7. `GET /v1/budget/me` (FR-14)

- **Response:** for the caller's identity, including OIDC member tokens, the endpoint returns each matching rule:
  - `rule_id`, `name`, `scope`, `limit`, `spent`, `remaining`, `window_end`, `on_exceed`
  - `unit: "credit"`
- The tenant pool is included only with billing-read permission.
- Amounts are strings, read from the Redis windows; the ledger is never scanned.

### 8. Tenant-admin UI (FR-13)

All of it is in fa and en, RTL-correct, built from existing `components/ui` and `app-ui`:

- **Balance:** by lot and expiry.
- **Usage:** by team, user, model, and product, over a time range.
- **Depletion forecast:** from the last 7 days, with explicit empty and uncertain states.
- **Credits saved by cache:** net of `semantic_cache` credits, with gross as secondary detail.
- **Plan page:** add-on purchase, overage toggle, and default `on_exceed`.
- **Budget rule form in credits:** with `on_exceed` (`block` or `alert`) and the alert thresholds. This includes updating `ruleValidation.ts`, the policy-rules simulator, and `loadtests/k6/budget-limit-rule.js`.

### 9. Cutover tool and fleet cutover (Phase 3 item 9, Phase 4 items 2–4)

Migration `071_credit_conversion.sql`:

- Adds ledger kind `convert` (positive amount, reason required) to the kind and sign checks.
- Adds a table `credit_conversions`:
  - Columns: `tenant_id`, `direction` (`to_credit` | `to_irr`), `rate_revision_id`, `wallet_irr`, `credits_mc`, `residue_irr`, `rules_converted`, `actor`, `reason`, `created_at`.
  - Unique on `(tenant_id, direction, sequence)`.

**Dry run.** For each tenant, it computes:
- The `migration` lot: `floor(balance_irr / 10 / list_toman_per_credit)` credits, using the active price revision. The residue is reported in IRR.
- Each rule's `limit_mc`: `limit_to` converted the same way, **rounded up** to a whole credit.

It also reports tenants with a negative wallet, pending top-ups, or open bank transfers.

**Execute.** One transaction per tenant, under key `convert:{tenant_id}:{n}`. It:
1. Grants the `migration` lot. The lot has **no expiry**: invariant 8 says converted value is never reduced.
2. Turns a negative wallet into debt instead.
3. Writes `limit_mc` for every rule.
4. Writes `wallet_transactions` `conversion_out` to zero the wallet.
5. Sets `billing_unit = credit`.
6. Writes `financial_audit_events`.
7. Re-seeds the windows and recompiles the profile after commit.

**Fleet.** A support-admin page lists the dry run for every IRR tenant and executes in batches of N. Each tenant is its own transaction, so a failure stops only that tenant. Re-running skips converted tenants.

**Rollback** (per tenant): it converts the remaining `available_mc − debt_mc` back to IRR at the rate stored with the conversion, as a `to_irr` row. It expires the remaining credit lots with reason `rollback`, and switches the unit back.

**Plan move** (P4), in the same transaction as the conversion:
- Cancel the active, trialing, or past-due subscription with event `credit_cutover`.
- Insert an `active` subscription on the published credit version of the same plan slug and billing interval, for the period now → now + interval. Its price snapshot holds the new price.
- Users already above the new `users.max` are grandfathered, as in Phase 2.
- A tenant with no subscription, or one whose plan has no published credit version, is reported by the dry run and keeps its subscription unchanged.

**Compensation lot** (Phase 4 item 3): not granted separately. Under P4 the free first period is the compensation.

**Deprecation** (Phase 4 item 4): after one full billing period, `x-godarai-cost-irr`, the IRR wallet fields, and `limit_to` get a `Deprecation` header and doc notes. The IRR runtime path goes behind `BILLING_IRR_RUNTIME_ENABLED` (default on until then).

### 10. Docs

- Update `cost-tracking.mdx`, `request-headers.mdx` (credit headers), and `guides/budget-limiting.mdx` (credits, `on_exceed`, the alert behavior under P1).
- Add a `budget/me` reference page.

## Feature flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `BILLING_CREDIT_SIDE_USAGE_ENABLED` | off | Spend logs and commits for the judge and the semantic-cache embedding. |
| `BILLING_CREDIT_OVERAGE_DEFAULT_CEILING_CREDITS` | 100000 | Headroom for overage with no cap. |
| `BILLING_IRR_RUNTIME_ENABLED` | on | Phase 4 kill switch for the IRR path. |

## Rollout

1. Merge PRs 3a–3d with all flags off. Nothing changes for IRR tenants.
2. In staging: enable the ledger and enforcement, convert test tenants, and run end to end (purchase → grant → consume → expire → reconcile, block, alert, BYOK, side usage, rollback). Run the load and chaos tests below.
3. Publish the credit plan versions in support-admin.
4. In production: enable the ledger and enforcement flags, then run the fleet dry run. Fix anything it reports.
5. Execute the fleet cutover in batches. Watch reconciliation and the `budget_exceeded` rate after each batch.
6. After one full billing period: deprecation headers, then the IRR flag off.

**Rollback:** per tenant with the rollback action, which a script can run fleet-wide. Credit enforcement has no flag of its own (owner decision, 2026-10-17): every credit tenant is always reserved in credits.

## PR split (stacked on #71)

| PR | Content |
| --- | --- |
| 3a | Billing unit in the profile, the credit pool window and its seeding, the settled table, BYOK reservation, rules in credits (`limit_mc`, `on_exceed` block/alert), 402 shaping and the three leak fixes, alerts as logs and metrics. |
| ~~3b~~ | Dropped (P5). |
| 3c | Side usage (judge, semantic cache), and `GET /v1/budget/me`. |
| 3d | Tenant-admin UI (balance, usage, forecast, cache savings, plan page, rule form) and docs. |
| 3e | The cutover tool, the fleet page in support-admin, rollback, compensation, and deprecation. |

## Tests

- **Unit:** mc estimation, window allowance math, both `on_exceed` branches and the 402 shape on all four API surfaces, BYOK split, conversion rounding (wallet rounded down, rules rounded up), and the residue.
- **Integration (Postgres + Redis):**
  - pool exhaustion → 402;
  - settle → re-seed → allowance stays correct across a month boundary;
  - overage headroom;
  - conversion run twice and concurrently gives one lot;
  - rollback restores IRR exactly;
  - the fleet batch keeps going after one failing tenant.
- **Load:** reservation p95 overhead ≤ 5 ms on the k6 profile, against the IRR baseline.
- **Chaos:**
  - Redis restart: windows re-seed with the same allowance.
  - Settlement worker killed mid-batch: no drift after the next pass.
  - Duplicate pack callbacks.

## Gate 3 as amended (P3)

- The staging end-to-end run is clean, with zero reconciliation difference.
- The production fleet dry run reports no unexplained case: every negative wallet, pending top-up, and open transfer is resolved or accepted.
- Load and chaos tests pass.
- After the cutover, reconciliation is clean every day and there are zero double charges. If either fails, stop the batches and roll back the affected tenants.

## Implementation notes (3a)

- **Forced commits.** A credit commit always succeeds, because the usage was already served. Commits use `BudgetCommitRequest.Force`, and alert rules use the new `RecordBudgetSpend`, which records spend with no admission check. Spend beyond an allowance is recorded as window overage (`shahrah_credit_budget_commit_overage_millicredits_total`), and the ledger debits it as debt.
  - Before this change, a commit beyond the allowance was refused, and several handlers then returned 402 *after* the provider call without writing a spend log. For a credit tenant, that request would never have been charged.
- **No wallet debit for credit tenants.** `DebitWallet` is the direct-debit fallback, used when a request reaches settlement without a reservation. It now returns without writing for credit tenants. The window settlement path was already guarded in Phase 2.
- **Switch guard.** Support-admin can move a tenant *to* credit only when `BILLING_CREDIT_LEDGER_ENABLED` is on, because the ledger's settlement is what debits a credit tenant. Moving back to IRR is always allowed. Credit enforcement itself has no flag (owner decision, 2026-10-17).
- **Profile and window refresh.**
  - Every billing-unit switch recompiles the tenant's runtime profile.
  - Every balance change re-seeds the current pool window: grants, purchases, bank-transfer approvals, adjustments, expiry, settlement, and overage-settings changes.
- **Rule windows** are re-seeded from `limit_mc` on every reservation, as IRR rule windows are, so a limit change applies at once.
- **Alerts** compare the rule window's committed spend before and after each commit against `alert_thresholds`. Each crossing is deduplicated with a cache counter keyed by window and threshold, and is logged as `budget_alert` plus `shahrah_credit_budget_alerts_total`.
- The credit rate-card cache is now built before the billing service, so reservations are priced with the published card.

## Implementation notes (3c)

- **Judge charges.**
  - The guardrail engine counts judge calls on a `guardrails.JudgeUsage` placed on the request context.
  - Only calls that reached the judge count; cached answers and failed calls are free.
  - The management "test policy" endpoint is never charged.
  - The input and output stages of `/v1/chat/completions` each charge their calls right after the stage, including when the guardrail blocks the request, because the judge ran.
- **Per charge for a credit tenant:**
  - **Budget windows.** The pool and every matching rule window are recorded after the fact through `RecordSideUsage`, with no admission check. The overshoot is at most one stage's judge calls.
  - **Spend log.** One spend log `{request_id}:{stage}` with product `guardrail`, credits as `fixed_price`, the caller's user, team, and org unit, and `trace_id` in the metadata. The ledger settles it like any request. The suffixed request id keeps the side call from replacing the request's own spend log in trace lookups.
  - IRR tenants are not charged for side usage.
- **Cache hits:** the chat cache lookup runs before reservation, so a hit reserves and commits nothing. `credits_avoided_mc` is already recorded (Phase 1).
- **`GET /v1/budget/me`** (route in `registerBudgetRoutes`), for any gateway credential:
  - Lists the caller's enabled credit rules whose applies-per is an identity scope (tenant, user, team, org unit, virtual account). Rules applied per model or per metadata value have one window per value and are left out.
  - Only identity conditions are evaluated. Request conditions (model, metadata, endpoint) cannot be known without a request, so a rule that matches the caller is listed regardless of them.
  - `remaining` subtracts both committed and reserved spend.
  - `tenant_balance` (remaining and reserved this month) is included for `tenant_admin`, `admin`, `billing:read`, and `billing:manage`.
  - IRR tenants get 409 `credit_billing_required`.
  - Data comes from the Redis windows plus a pool-allowance read when the month has no window yet. The response is `no-store`.

## Decisions made (conservative defaults, easy to flip)

- `block_on_budget_exceed = false` becomes `alert`, not `block`, so it keeps blocking less.
- Wallet credits are rounded down and rule limits up, as the prompt says. Both favor the tenant.
- Overage with no cap gets a 100,000-credit ceiling instead of unlimited spend.
- Migration lots never expire.
- A credit tenant whose credit windows are unavailable is refused, never served unenforced.
- IRR budget refusals carry no `scope`. The IRR path does not track which window refused, and it is left untouched.
- An unparseable `credits.byok_fee_percent` falls back to the 5% default.

## Open issues

1. ~~§8.4: plans and prices of existing subscriptions.~~ Resolved by P4: same-slug credit plan at cutover, first period free, new price from the next renewal.
2. ~~`infra/k8s/local/migrations` stopped at 050.~~ No script, manifest, or Makefile references it (only planning docs do), and Phase 2 added 066–069 without mirroring. New migrations go to `db/main/migrations` only.

# Execution Prompt: Credit-Based Pricing Framework (Shahrah / GodarAI)

Use this prompt with a coding agent that has full access to the Shahrah repository.
The LibreChat side is a separate prompt: `plans/credit-pricing-librechat-execution-prompt.md`.

---

You are a senior software engineer implementing GodarAI's **credit-based pricing framework**.
You own the design and the delivery quality: correctness of money, idempotency, backward
compatibility, and operability matter more than speed. You work in phases. Each phase ends
in a mergeable PR and a written report, and you **stop and wait for approval** before
starting the next phase.

## 1. Sources of Truth

Read these before writing any code:

1. Product spec (Persian): `docs/PRD چارچوب قیمتگذاری کردیتی گدارAI.md`. This prompt
   resolves its open questions; where they disagree, **this prompt wins**.
2. `AGENTS.md`: pnpm only, the Go lint rules, and the UI workflow (read
   `docs/design-system/DESIGN.md`, `docs/design-system/COMPONENT_USAGE.md`, and reuse
   `apps/tenant-admin/components/ui` and `components/app-ui`).
3. Related plans for background only: `plans/customer-budget-architecture.md`,
   `plans/budgets_rate_limits.md`, `plans/budget_reset_and_tz.md`,
   `plans/01_iranian_payment_system.md`, `plans/support-admin-tenant-subscription-execution-prompt.md`.

## 2. Problem in One Paragraph

Today the tenant wallet and `BudgetLimitRule.limit_to` are denominated in IRR. Each request
is priced in USD by the pricing engine and converted to IRR with the live `ExchangeRate`.
Every currency jump therefore moves the per-message cost, team budgets, and customer
forecasts at the same time. We are moving to a stable consumption unit, the **credit**:
`credits = 500 × official provider USD cost`. Only the toman price of a credit moves with
currency, and only inside a banded policy. IRR remains for purchases, invoices, and internal
accounting.

## 3. Locked Decisions

### 3.1 From the PRD

| Topic | Decision |
| --- | --- |
| Conversion | `credits = 500 × Σ_c q_c × p_c^USD` over every `UsageCostBreakdown.Components` entry (input, output, cache read, cache write, image, audio, search, other tools). The factor 500 is a constant, never tied to FX. |
| List price | 1 credit = 1,000 toman = **10,000 IRR** at base rate 270,000 toman/USD. |
| Storage | Balances and ledger amounts are `int64` **millicredits** (1 credit = 1,000 mc). |
| Rounding | Rates keep full fractional precision (decimal, not float). Round **once per request total**, **up**, to 0.01 credit (that is, to a multiple of 10 mc). |
| Display | Two decimal places; Persian digits and thousands separators in `fa`. |
| Provider ratios | Input/output, cache discount, cache write, and tool ratios pass through unchanged. A per-model multiplier is allowed only when GodarAI's real cost differs (egress, tunnel), and always with a stored reason. |
| Versioning | The rate card is versioned; every transaction stores `PricingRevision` plus the rate-card version. |
| Plans | Trial (14 days, 3 users, 500 one-time credits), Developer (3 users, 5,000,000 toman, 5,000 credits), Team (10, 15,000,000 toman, 16,500), Growth (30, 45,000,000 toman, 54,000), Business (100, 150,000,000 toman, 190,000), Enterprise (custom, ≥ 750 toman/credit). Prices exclude 10% VAT. Discounts are delivered as bonus credits, never as a lower per-credit list price. |
| Add-on packs | 1,000 credits / 1,000,000 toman; 5,000 / 4,500,000; 20,000 / 16,000,000. Valid for 12 months. |
| Expiry | Plan credits expire at the end of each subscription period. Add-ons after 12 months. Promo and trial credits are short-lived (expiry set per grant). |
| Consumption order | Earliest expiry first. On equal expiry: plan, then trial/promo/compensation, then add-on. |
| Overage | Off by default. The tenant admin can enable it; it is billed at the current list price. |
| Annual | At most 10% discount, credits delivered monthly, FX adjustment clause. |
| Rollover | Enterprise only, conditional on renewing at the same or a higher tier. |
| FX band | ±10% around the base rate. If the **7-day average** stays outside the band, the new list price = old price × (avg / base), rounded to the nearest 50 toman, and avg becomes the new base. Example: 300k → 1,100 toman. |
| Price-change notice | Plan price increases: 30 days' notice, effective from the next billing period. Add-on prices apply immediately to new purchases. Provider price cuts pass through immediately with a new rate-card version; provider increases get 30 days' notice. Credits already purchased never lose consumption value. |
| Changelog | Public changelog for every credit price or rate-card version change, with an effective date. |
| Budgets | Reserve → commit → release keeps the current cycle (`ReserveBudget`/`CommitBudget`/`ReleaseBudget`), now in credits. |

### 3.2 Open Questions Resolved by the Product Owner

| Question | Decision | Implementation consequence |
| --- | --- | --- |
| GodarAI cache hits (exact and semantic) | **Free** (the cached *response* only) | No debit or budget increment for the response itself. Record `credits_avoided` in analytics for the "credits saved by cache" metric. The semantic-cache embedding is still billed (next row), so a semantic hit costs only its embedding and an exact hit costs 0. |
| Side usage | **Both billed to the customer: the guardrail judge and the semantic cache embeddings.** | Each side call produces its own `consume` entry: `product=guardrail` for the judge, `product=semantic_cache` for the embedding (on lookup and on write, hit or miss). Both are attributed to the originating request's identity (user, team, org_unit, trace) and counted in matching budgets. `x-godarai-credits-used` and the telemetry `credits` field report the total for the request, including side calls; the ledger keeps them as separate entries. |
| BYOK (`external_billed`) | **Percentage platform fee**, configurable, default **5%** | The full credit equivalent is counted in budgets and reports. Only `ceil10(equivalent × fee%)` is debited from the balance. Global default, overridable by plan entitlement and tenant override. |
| Per-plan user cap (3/10/30/100) | **Hard limit** | Enforced server-side on user create, activate, and invite. See §6 Phase 2. |

### 3.3 Defaults This Prompt Chose (record them in the design note, easy to flip)

| Topic | Default |
| --- | --- |
| Publishing the 500-credit/USD ratio | **Internal.** The rate card and docs expose credit rates only. Do not show USD prices or the factor in public APIs, UI, or docs. Keep the factor in server config. |
| FX band reset | The job **detects and proposes** a new list price. A support-admin approves it (two-person rule is not required). Auto-apply exists as a flag, default off. |
| User-cap counting | Active human tenant members + pending invitations. Virtual users, service keys, and PATs are excluded. Tenants already over the cap at rollout are grandfathered: they cannot add users and cannot downgrade below their current count. |
| Downgrade accounting | A downgraded request is not charged to the rule that triggered the downgrade. It is charged to the tenant balance, to all other matching rules, and to a separate daily downgrade quota per (rule, scope value). When the quota is exhausted, the request is blocked. |

## 4. Current-State Code Map (verify before relying on it)

| Area | Location |
| --- | --- |
| Budget reserve/commit/release | `services/control-plane/internal/billing/budget_service.go`, `budget_policies.go` |
| Atomic budget windows (Redis + memory) | `services/control-plane/internal/cache/cache.go` (`DualCache.ReserveBudget` etc.), `memory.go` |
| Wallet settlement from windows | `services/control-plane/internal/billing/wallet_settlement_processor.go`, `internal/storage/wallet_budget_settlement_repo.go`, `wallet_live_balance.go`, `wallet_repo.go` |
| Pricing engine and strategies | `services/control-plane/internal/billing/pricing_engine.go`, `pricing_strategy_*.go`, `pricing_override.go`, `*_pricing_snapshot.go` |
| FX rates | `exchange_rate_cache.go`, `nobitex_exchange_rate.go`; migrations `012`, `043` |
| Cost types | `packages/contracts/platform/types.go`: `UsageCostBreakdown`, `UsageCostComponent` (decimal **strings**), `Wallet` (IRR), `SpendLog`, `BudgetReservation` |
| Budget rule contract | `packages/contracts/platform/budget_limit_rules.go` (`LimitTo float64`, `BlockOnBudgetExceed bool`, `Alerts`); migration `010_budget_limit_rules.sql` |
| Plans and entitlements | migration `021_commercial_plans_entitlements.sql` (`commercial_plans`, `plan_versions`, `plan_entitlements`, `plan_prices.amount_minor` in IRR, `tenant_subscriptions`, `tenant_entitlement_overrides`); `internal/storage/commercial_plan_repo.go`; `billing/subscription_processor.go` |
| Trial / signup credit | migration `025` (`commercial_signup_settings.signup_credit_irr`, `trial_days`) |
| Wallet transactions, invoices, gift/discount codes, audit | migration `037_financial_management.sql` |
| Request pipeline, BYOK branch | `internal/proxy/request_pipeline.go`, `billing_metadata.go`, `handler.go` (~L818 `external_billed`) |
| Budget error mapping | `internal/proxy/rate_limit_errors.go` (402 `budget_exceeded`). **Leaks `err.Error()`:** `anthropic_native_handler.go:285`, `gemini_native_handler.go:138`, `tokenbazaar_sync.go:84` |
| Headers / stream telemetry | `internal/apiheaders/headers.go` (`x-godarai-cost-usd`, `x-godarai-cost-irr`, `x-godarai-include-stream-telemetry`); `godarai.telemetry` event in `proxy/handler.go` |
| Delegated identity today | Master key + `x-godarai-user-id` in `internal/auth/middleware.go` (`handleMasterKeyRequest`) |
| Key types | `KeyTypePAT`, `KeyTypeVAT` in `types.go` |
| Cache savings | `AvoidedCostUSD` in `internal/observability/metrics/processor.go`, `storage/metrics_repo_clickhouse.go` |
| Semantic cache / judge | `internal/responsecache/` (embedder), `internal/semanticjudge/`, `internal/guardrails/` |
| Analytics schema | `infra/clickhouse/migrations/` |
| Postgres migrations | `db/main/migrations/` (goose; next number is **061**). `infra/k8s/local/migrations` stopped at 050; check whether it still needs mirroring before you add files there. |
| Tenant-admin billing UI | `apps/tenant-admin/app/(admin)/billing/*`, `features/billing/*`, `features/gateway-policies/budget-limiting/*`, messages in `messages/modules/{billing,budgetLimitRules,...}/{fa,en}.json` |
| Support-admin | `apps/support-admin/app/admin/pricing`, `.../tenants` |
| Public site / docs | `apps/website/app/pages/Pricing.tsx`, `CommercialPricing.tsx`; `apps/docs/docs/{cost-tracking,request-headers}.mdx`, `guides/budget-limiting.mdx` |
| Decimal library | `github.com/shopspring/decimal` already in `services/control-plane/go.mod` |

## 5. Engineering Invariants (non-negotiable)

1. **No floats for money.** Credit math uses `decimal.Decimal` from component strings
   (`UsageCostComponent.CostUSD`/`UnitPriceUSD`, `Quantity`), never `TotalCostUSD float64`.
   Persist `int64` millicredits. Amounts on the wire are decimal **strings** with exactly
   two decimals (`"4.50"`).
2. **Toman ≠ rial.** The PRD speaks toman; the DB stores IRR (`amount_minor`, `*_irr`).
   1 toman = 10 IRR. Every price constant and test must state its unit in the identifier
   (`listPriceTomanPerCredit`, `amountIRR`). Add a unit test that fails if a plan price is
   off by 10×.
3. **Append-only ledger.** No `UPDATE`/`DELETE` on ledger or allocation rows (enforce it
   with a DB trigger and a test). Corrections are new `adjust`/`refund` entries. Projections
   (lot balances, cached wallet) may be updated but must be rebuildable from the ledger.
4. **Idempotency everywhere.** Every commit, release, grant, expiry, refund, and conversion
   has a deterministic idempotency key with a unique index. Retrying any job or request must
   never debit twice. Write tests that run each operation twice and concurrently.
5. **Hot path budget.** Credit reservation adds ≤ 5 ms p95. Keep using the cached budget
   windows. Ledger writes are asynchronous (settlement processor / outbox), never inline in
   the request.
6. **Client-safe errors.** Budget and credit errors expose only a stable `code` and the
   `scope`, never amounts, internal IDs, or raw exception text, across all API surfaces
   (OpenAI-compatible, Anthropic-native, Gemini-native, TokenBazaar).
7. **Backward compatibility until sunset.** IRR headers and fields keep working until
   Phase 4 deprecates them. Each tenant has `billing_unit ∈ {irr, credit}` (default `irr`).
   Redis window keys must encode the unit (for example a `mc:` segment or a schema version)
   so IRR and credit windows never mix.
8. **Never reduce purchased value.** No price reset, migration, or rate-card change may
   reduce an existing lot's remaining credits.
9. **Auditability.** Every manual action (adjust, approve price, convert tenant, grant
   compensation) writes `financial_audit_events` with the actor and a reason.
10. **Go and UI rules** from `AGENTS.md` apply to every change: lint, `wsl_v5`, cyclomatic
    limits, closed bodies in tests, semantic Tailwind tokens, RTL, and fa/en strings.

## 6. Phases and Gates

The PRD's phase diagram was not available, so these four phases and three gates are the
plan. Each phase is one branch (`feat/credits-phase-<n>-<slug>`) and may be split into
several stacked PRs if it exceeds roughly 1,500 changed lines.

**Start of each phase:** write `plans/credit-pricing/phase-<n>-design.md` with the schema
DDL, contract changes, sequence diagrams for reserve/commit/settle, the idempotency keys,
the feature flags, the rollout and rollback plan, and an "Open issues" list. If the open
issues contain anything listed in §8, stop and ask before you code.

---

### Phase 1: Credit Core, Rate Card, Shadow Mode

**Goal:** compute credits for every request alongside IRR without changing enforcement.

1. **Credit math package** (`services/control-plane/internal/billing/credits` or similar):
   - `FromBreakdown(breakdown, rateCard) (Millicredits, CreditDetail, error)`. Per
     component: `qty × unitPriceUSD × 500 × modelMultiplier`, summed in decimal and then
     rounded up once to 10 mc.
   - Golden tests from the PRD's examples. Per 1K tokens: Opus 5.5 input 2 / output 10 /
     cache read 0.1 / search 5. Kimi K3 1.5 / 7.5 / 0.15. Sonnet 5 1 / 5 / 0.1 / 5.
     Grok 4.7 1 / 3 / 0.25 / 2.5. Light model 0.125 / 1. Typical message (2,000 in /
     500 out): light 0.75, Sonnet 5 4.50, Opus 5.5 9.00.
   - Property tests: monotonic in quantity, always a multiple of 10 mc, never below the
     exact value, no float path.
2. **Rate card versioning.** New tables:
   - `credit_rate_card_versions`: version, status (draft, published, superseded),
     effective_from, published_by, changelog entry, notice_until.
   - `credit_model_multipliers`: model, multiplier as decimal, `reason NOT NULL`, version.

   The published rate card is derived from the existing provider pricing snapshots ×
   500 × multiplier. Store `rate_card_version` next to `PricingRevision` on spend logs.
3. **Shadow computation.** On every request, after the existing USD/IRR pricing, compute
   millicredits and persist them on `SpendLog` / ClickHouse events: `credits_mc`,
   `rate_card_version`, `product`, `billing_mode`, `credits_avoided_mc` for cache hits.
   This is gated by the global flag `credits.shadow_enabled`.
4. **Product attribution.** Add a `product` dimension to every event: `gateway`,
   `librechat`, `guardrail`, `semantic_cache`, `embeddings`, `ocr`, `playground`. Define the derivation rules
   in the design note.
5. **Response surface (additive):** an `x-godarai-credits-used` header on non-stream
   responses and a `credits` field (string) in the `godarai.telemetry` stream event. Add
   the constants to `apiheaders`. IRR headers stay unchanged.
6. **Delegated identity through users' OIDC access tokens** (product decision; it
   replaces an on-behalf-of service-key design):
   - Delegated callers such as LibreChat send the signed-in member's GodarAI OIDC
     access token as the bearer credential. The gateway already authenticates these
     tokens with a user-scoped model policy and blocks management APIs.
   - Add a per-client `credit_product` (`oidc_app` default, or `librechat`) that tenant
     admins choose in "Connected apps". Stamp it into access tokens at issuance and
     carry it to spend logs as a trusted product label.
   - Attribute usage to the user and tenant only; do not derive a team.
   - The existing master-key `x-godarai-user-id` path is unchanged.
7. **`GET /v1/rate-card`:**
   - Public, cacheable with `ETag`; when authenticated, filtered to the tenant's
     available models.
   - Returns `version`, `effective_from`, `unit: "credit"`, and per model: rates per 1K
     tokens, per call, and per image as decimal strings, plus `typical_message_credits`
     (2,000 in / 500 out).
   - No USD and no factor (§3.3).
8. **Shadow dashboard (tenant-admin):** a banner and card on the billing page: "this month
   you would have consumed X credits under the new model". Show usage per model and per
   team in credits.
9. **Calibration report:**
   - A script or admin endpoint over the last 90 days of ClickHouse data.
   - Outputs the median, p75, and p95 of credits per user per month, per tenant and per
     plan bucket, plus the projected share of plan quota consumed.
   - Stored under `output/` or exposed in support-admin.

**Gate 1 (exit criteria)**

- Shadow credits run in staging and production for ≥ 7 days, with `credits_mc` within
  ±1 rounding step of `ceil10(TotalCostUSD × 500,000)` for ≥ 99.9% of requests;
  discrepancies are logged and explained.
- Measured p95 overhead of the credit computation ≤ 1 ms. The full reservation budget
  of 5 ms is reserved for Phase 3.
- The rate-card API is documented in `apps/docs`, and the calibration report has been
  produced and shared.
- No change in any IRR number, header, or enforcement decision (proved by tests).

---

### Phase 2: Ledger, Lots, Plans, Purchases, Expiry, Price Policy

**Goal:** the full credit balance model exists and can be granted, purchased, consumed
(for pilot tenants), and expired. Enforcement is still IRR for everyone else.

1. **Schema** (migration 06x, partition the high-volume tables by month):
   - `credit_lots`: lot_id, tenant_id, source (`plan`, `trial`, `promo`, `purchase_addon`,
     `migration`, `compensation`, `adjust`), amount_mc, expires_at, purchase_price_irr,
     price_per_credit_irr, source_ref (subscription, period, intent), created_at.
   - `credit_ledger`: append-only. entry_id, tenant_id, kind (`purchase`, `plan_grant`,
     `promo_grant`, `consume`, `refund`, `expire`, `adjust`), amount_mc (signed),
     idempotency_key UNIQUE, occurred_at, and for `consume`: request_id, trace_id, model,
     user_id, team_id, org_unit_id, product, billing_mode, pricing_revision,
     rate_card_version, internal_cost_usd (decimal string), credit_equivalent_mc (BYOK),
     reason/actor for manual kinds.
   - `credit_ledger_allocations`: append-only. entry_id, lot_id (nullable means
     unallocated debt), amount_mc. Allocations happen at settlement time in
     expiry-then-source order (§3.1). If a commit exceeds the available lots (estimate
     vs. actual race), the remainder is a debt allocation with `lot_id NULL`. Debt is
     paid by the next grant before that grant becomes spendable, through new allocation
     rows, never by editing.
   - `credit_lot_balances`: a projection, updated in the same transaction as the
     allocations, plus a nightly rebuild-and-compare job.
2. **Settlement:** extend the settlement processor pattern so committed credit windows
   flow into `consume` entries asynchronously.
   - Idempotency key: `consume:{tenant}:{request_id}:{product}`.
   - Daily reconciliation: the ledger sum must equal the cached balance and the projection.
     Any non-zero difference raises an alert with the tenant and amount, and is never
     auto-fixed.
3. **Expiry job:**
   - At each lot's `expires_at`, write `expire` for the lot's remaining amount.
   - Idempotency key: `expire:{lot_id}`; safe to run twice and concurrently.
   - Must not expire credits held by in-flight reservations until they settle: commit
     reallocates them to the next lot or to debt.
   - Invalidate the cached balance after it runs.
4. **Plans as data:**
   - Seed the five plans and Enterprise through the existing
     `commercial_plans` / `plan_versions` / `plan_prices` / `plan_entitlements`. Prices go
     in IRR (×10 from toman).
   - New `feature_catalog` keys, for example: `credits.monthly_grant`,
     `credits.bonus_percent`, `users.max_active`, `credits.overage_allowed`,
     `credits.rollover_allowed`, `credits.byok_fee_percent`, `credits.trial_grant`,
     `credits.trial_days`.
   - Enterprise values come from `tenant_entitlement_overrides`.
   - Each period start grants a `plan` lot with an idempotency key
     `plan_grant:{subscription_id}:{period_start}`. Annual plans deliver monthly.
   - Trial: 500 credits, 14 days, 3 users. Replace `commercial_signup_settings.signup_credit_irr`
     with credit fields; keep the old column until Phase 4.
5. **Add-on packs:**
   - A versioned `credit_packs` catalog.
   - Purchased through the existing payment flow (`wallet_topup_intents`, the gateway
     callback, bank transfer), with VAT handled as today.
   - A successful payment creates a `purchase` entry plus a 12-month lot, idempotent on
     the payment/intent id.
6. **Overage:**
   - Tenant setting `credits.overage_enabled` (default false) plus an optional
     `credits.overage_cap`.
   - When enabled and lots are exhausted, consumption accrues as debt that is invoiced at
     period end at the list price in effect, as an IRR line item.
7. **Hard user cap:**
   - Enforce `users.max_active` in the user create, activate, and invite handlers
     (server-side, inside a transaction or with a lock to avoid races), with error
     `plan_user_limit_reached` (403).
   - Block plan downgrades below the current count.
   - Grandfather tenants that are already over the cap (§3.3).
   - Show used/max in tenant-admin.
8. **Credit price policy:**
   - `credit_price_revisions`: price_toman_per_credit, base_usd_toman, band_pct,
     status (proposed, approved, active, rejected), effective_from, notice_sent_at,
     approved_by.
   - Daily job: compute the 7-day average from the existing FX source. When it is
     outside the band, propose a new price (rounded to 50 toman) per §3.1.
   - Support-admin approve/reject UI. Apply the notice rules: add-ons immediately, plan
     price changes 30 days ahead and from the next period.
9. **Public changelog:**
   - `GET /v1/pricing/changelog`: rate-card versions and credit price changes with
     effective dates.
   - A page in `apps/website` and a link from tenant-admin. Time-limited promotions
     show their end date.
10. **Support-admin tooling:**
    - Plan and pack catalog.
    - Rate-card multipliers with a required reason.
    - Manual `adjust` with a reason.
    - Per-tenant ledger viewer with lot drill-down.

**Gate 2 (exit criteria)**

- In staging, purchase → grant → consume → expire → reconcile runs end to end with zero
  reconciliation difference, for plan, trial, add-on, promo, and overage paths.
- Concurrency tests show no double grant, double expiry, or double debit under retries
  and parallel workers.
- Finance has signed off on a sample IRR invoice (fixed plan price, add-ons, overage, VAT).
- The user cap is enforced, with a test for each entry point.

---

### Phase 3: Credit Enforcement, Budget Rules, Visibility

**Goal:** tenants with `billing_unit=credit` are enforced in credits end to end. Pilot it
with a small set of internal and friendly tenants.

1. **Reservation in credits:** `ReserveBudget`/`CommitBudget`/`ReleaseBudget` operate on
   millicredits for credit tenants. The tenant pool's available amount = the lot
   projection − debt − active reservations. Keep the IRR path untouched for IRR tenants.
2. **Budget rules (FR-5/6/7):**
   - Add `limit_mc bigint`, `on_exceed` (`block`, `downgrade`, `alert`; default `block`),
     `downgrade_model` (validated against the tenant's available or virtual models),
     `downgrade_daily_limit_mc`, and `alert_thresholds` (default `[80, 100]`).
   - Migrate `block_on_budget_exceed`: `true → block`, `false → alert`. Keep reading the
     old field for one release, and reject requests that set both inconsistently.
   - Scopes and conditions (user, team, org_unit, model, virtualaccount, `metadata.*`)
     are unchanged.
   - Update the contract types, admin API, `ruleValidation.ts`, the form, the
     policy-rules simulator, and the k6 test `loadtests/k6/budget-limit-rule.js`.
3. **Exceed behavior:**
   - `block` (FR-9): HTTP 402
     `{"error":{"type":"budget_exceeded","code":"budget_exceeded","scope":"<scope>"}}`,
     shaped per API surface. Fix the three `err.Error()` leaks listed in §4.
   - `downgrade` (FR-8): route to `downgrade_model` within the daily downgrade quota
     (§3.3). Set `x-godarai-model-substituted: <model>` and
     `x-godarai-original-model: <model>`, and put the same in the telemetry event.
     Re-run the matching rules for the substituted model. When the quota is exhausted,
     block.
   - `alert`: allow, and emit alerts.
   - FR-10: if the tenant balance is zero and overage is off, apply the tenant default
     `credits.default_on_exceed`.
4. **Alerts (FR-7):**
   - Fire one alert per (rule, scope value, window, threshold), deduplicated with a
     unique key.
   - Recipients: tenant admins plus the admins of the affected team.
   - Find the existing outbound notification and email infrastructure first (only the OTP
     mailer and notifier in `internal/admin` are known). If there is no reusable
     notification channel, **stop and ask** which channel to build.
5. **BYOK (FR-11):**
   - Budgets count the full credit equivalent.
   - The balance is debited only by the fee (default 5%, from the entitlement or
     override).
   - The reservation reserves the equivalent against the rules and the fee against the
     tenant pool.
   - The ledger `consume` stores both `amount_mc` (the fee) and `credit_equivalent_mc`.
6. **Cache hits and side usage:**
   - GodarAI cache hits debit nothing for the cached response.
   - The guardrail judge (`product=guardrail`) and the semantic cache embedding
     (`product=semantic_cache`, on both lookup and write) each commit a post-hoc
     `consume` against the same identity and rules.
   - Side calls are not reserved in advance; document the possible small overshoot.
   - Exception: if the tenant balance is already exhausted, skip the semantic cache
     lookup instead of charging it. The request is then handled by the normal
     `on_exceed` path.
   - The tenant-admin usage breakdown must show `guardrail` and `semantic_cache` as
     separate products so customers can see what cache and guardrails cost them.
7. **`GET /v1/budget/me` (FR-14):**
   - For the caller's identity, including members calling with OIDC access tokens, return each matching rule:
     `rule_id`, `name`, `scope`, `limit`, `spent`, `remaining`, `window_end`, `on_exceed`,
     plus `unit: "credit"`.
   - The tenant pool appears only when the caller has billing-read permission.
   - Amounts are strings. Serve from the cache, not from a ledger scan.
8. **Tenant-admin UI (FR-13):**
   - Balance by lot and expiry.
   - Usage by team, user, model, and product, with a time range.
   - A depletion forecast from the last 7 days' trend, with the empty states and
     uncertainty spelled out.
   - "Credits saved by cache": `AvoidedCostUSD × 500` minus the `semantic_cache` credits
     for the same period, so the figure is net savings. Show the gross figure as
     secondary detail, and no USD.
   - Plan page with the five plans, add-on purchase, overage toggle, and default on_exceed.
   - Budget rule form in credits.
   - All of it in fa and en, RTL-correct, using existing components. Update the design
     system playground if you add a reusable pattern.
9. **Per-tenant cutover tool** (used for pilots now and the fleet in Phase 4):
   - A dry run converts the IRR wallet balance at the day's list price
     (`balance_irr / 10 / list_toman_per_credit`, rounded down, with the residue reported)
     into a `migration` lot.
   - It converts each rule's `limit_to` the same way, rounded up to the nearest whole
     credit.
   - The tenant admin approves in the UI. Execution is idempotent
     (`convert:{tenant_id}`), flips `billing_unit`, and writes an audit event.
   - Rollback switches the unit back to IRR and converts the remaining credits back
     to IRR at the same rate.
10. **Docs:** update `cost-tracking.mdx`, `request-headers.mdx`, `guides/budget-limiting.mdx`,
    and add rate-card and budget/me reference pages.

**Gate 3 (exit criteria)**

- At least 3 pilot tenants run in credit mode for one full billing period.
- Billing tickets in that period stay below 5% of pilot tenants. Zero double charges, and
  reconciliation is clean every day.
- Load test: reservation p95 overhead ≤ 5 ms at the current production RPS profile.
- Chaos tests (Redis restart, settlement worker crash mid-batch, duplicate callbacks)
  pass without balance drift.

---

### Phase 4: Fleet Migration and IRR Sunset

1. **Notification:** a direct message to each tenant admin with a calculator (their
   last 90 days in credits against the recommended plan) and the cutover date.
2. **Fleet cutover:** in batches, using the Phase 3 tool, with admin approval per tenant.
   Include a fallback policy for tenants that do not respond by the date; document it
   and ask the product owner to approve it before running.
3. **Compensation:** grant a `compensation` lot equal to one month of the tenant's plan
   quota. A 30-day window allows `refund` entries for unexpected charges, through a
   support-admin action with a reason.
4. **Deprecation:** after one full billing period in credits, mark `x-godarai-cost-irr`,
   the IRR wallet fields, and `limit_to` IRR as deprecated (docs plus a `Deprecation`
   header). Removal is a separate, later PR.
5. **Success metrics dashboard** (support-admin): gross margin on provider cost, median
   plan quota usage, the share of tenants hitting the cap, breakage, billing tickets, and
   churn against the previous 3 months (see the PRD's KPI table).

**Exit:** every active tenant is on `billing_unit=credit`, and the IRR runtime path is
behind a flag, ready for removal.

## 7. Testing Requirements (every phase)

- Go: unit tests for the math, allocation order, rounding, and the toman/IRR boundary;
  table tests for every `on_exceed` branch across all API surfaces; integration tests under
  `services/control-plane/tests/integration` for the ledger, settlement, expiry, and cutover
  against real Postgres and Redis; concurrency tests with `-race`.
- Contract tests for the new headers, the telemetry field, `/v1/rate-card`, and
  `/v1/budget/me` (OpenAI SDK contract suite under `services/control-plane/tests`).
- Frontend: vitest for the formatters (Persian digits, two decimals, separators) and the
  rule validation; DOM tests for the budget form and the balance views.
- Migrations: up and down on a copy of realistic data; the append-only trigger test.
- Run before finishing each phase:
  - In `services/control-plane` and `packages/contracts`: `go vet ./...`,
    `go test ./...`, and
    `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`.
  - In each changed app (`tenant-admin`, `support-admin`, `website`, `docs`):
    `pnpm lint`, typecheck, `pnpm test`, and `pnpm build`.

## 8. Stop and Ask (do not guess)

Stop and ask the product owner when any of these come up:

1. There is no reusable outbound notification channel for budget alerts.
2. LibreChat cannot forward a usable member access token for some request path (for
   example background jobs without a user session).
3. The policy for tenants that never approve their conversion by the cutover date.
4. Anything that would change an existing tenant's effective price, quota, or balance
   beyond what this prompt defines.
5. Retention, partitioning, or archival limits for `credit_ledger` that conflict with
   finance or legal requirements.
6. A provider whose pricing strategy does not produce `Components` with decimal strings
   (the credit math would otherwise fall back to floats).
7. Any FX source or rate other than the existing exchange-rate pipeline.

For anything else that is ambiguous, pick the conservative option (charge less, block
less, keep existing behavior), record it under "Decisions made" in the phase design note,
and call it out in the phase report.

## 9. Out of Scope

- Separate hard wallets per team (deferred to Enterprise later).
- Transferring credits between tenants.
- Rewriting the USD pricing engine. Credits are built on top of its output.
- LibreChat code (separate prompt).

## 10. Phase Report Format

At the end of each phase, reply with:

1. What shipped (PR links or branch, migrations, flags).
2. Gate checklist with evidence (test output, metrics, screenshots for UI).
3. Decisions made and assumptions, each one flagged for confirmation.
4. Remaining risks and debt, including UI debt per `AGENTS.md`.
5. What the next phase needs from humans (approvals, data, decisions).

Do not push branches or open PRs until the user approves. Commit messages follow
Conventional Commits, as in the existing history.

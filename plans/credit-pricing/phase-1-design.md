# Phase 1 Design: Credit Core, Rate Card, Shadow Mode

Status: in progress · Branch: `feat/credits-phase-1-core`
Prompt: `plans/credit-pricing-execution-prompt.md` §6 Phase 1

## Goal

Every request computes credits next to the existing USD/IRR cost. Nothing in enforcement,
wallets, IRR headers, or budget windows changes.

## Findings From the Code Read

- All nine registered providers (`internal/providerregistry/registry.go`) price through a
  snapshot (`anthropic` or `catalog`). Every strategy fills `UsageCostBreakdown.Components`
  with decimal-string `CostUSD`, so credit math can stay fully decimal.
- Legacy fallback: `PricingEngine.modelPrice` → `pricingCache`/registry for models outside
  the provider registry. These breakdowns may have no `Components`. Shadow mode falls back
  to `decimal.NewFromFloat(TotalCostUSD)` and tags the result `source=float_fallback`, with
  a counter metric. Enforcement (Phase 3) must fail closed on this source. That is
  stop-and-ask item 6 if any production model still lands there.
- Private deployment pricing (`PricingModePrivate`, `ApplyPricingOverride`) rewrites the
  component rates before billing. Credits use the final components, so a private price
  card flows into credits the same way it flows into IRR today.
- Spend logs: a single helper, `logSpendBestEffort` (`internal/proxy/handler.go`), is used by
  every handler. It writes Postgres `spend_logs` through the async batch writer. This is the
  central shadow hook.
- Analytics: ClickHouse `gateway_model_events` (per-user and per-team) and `request_logs_ch`.
  Control-plane validates the ClickHouse columns at startup but does not migrate them, so
  the ClickHouse migration must be applied **before** the deploy.
- Cost headers are set in one place (`completions_handler.go:814`). The stream telemetry
  event is written in `writeStreamTelemetryEvent` (`handler.go`).
- Delegated identity today: master key + `x-godarai-user-id`. Key types: `pat` and `vat`.

## Units and Math (`internal/domain/credits`)

- `Millicredits int64`: 1 credit = 1,000 mc.
- `CreditsPerUSD = 500` is a server constant. It is never exposed (prompt §3.3).
- `exact_mc = Σ component.CostUSD × 500 × 1000 × multiplier(model)`, all in `decimal`.
- `charged_mc = ceil(exact_mc / 10) × 10`: a single round-up to 0.01 credit per request.
  A zero cost stays 0.
- `Format(mc) → "4.50"`: always two decimals, for the wire format.
- Golden tests: the PRD rate table and the typical-message examples (0.75 / 4.50 / 9.00).

## Rate Card Versioning (Postgres migration 061)

```sql
credit_rate_card_versions(
  version integer PK, status text CHECK (draft|published|superseded),
  effective_from timestamptz, notice_until timestamptz,
  changelog_fa text, changelog_en text,
  published_by text, created_at, updated_at)
  -- unique partial index: one published
credit_model_multipliers(
  version integer FK, model text, multiplier numeric(10,4) CHECK (> 0),
  reason text NOT NULL CHECK (length(trim(reason)) > 0),
  PRIMARY KEY (version, model))
```

- Seed: version 1, published, effective now, no multipliers (identity).
- The runtime label is `rc-<version>`.
- Credit inputs come from provider prices, which are already versioned by `PricingRevision`.
  The rate-card version only versions the multiplier set and the public changelog.
- A cached reader (`RateCardCache`) loads the published version with its multipliers,
  using a TTL like the pricing cache. If the reader fails, it uses the last good version.
  On a cold failure it uses version 1 with an identity multiplier and logs a warning.

## Shadow Persistence

Postgres `spend_logs` (migration 061), additive and nullable:

| column | type | notes |
| --- | --- | --- |
| `credits_mc` | bigint | charged (rounded) millicredits |
| `credits_avoided_mc` | bigint | cache hits only |
| `rate_card_version` | integer | |
| `product` | text | see below |
| `billing_mode` | text | `platform_billed` / `external_billed` |
| `credit_source` | text | `components` / `float_fallback` |

ClickHouse (`infra/clickhouse/migrations/006_credit_shadow.sql`): the same columns on
`gateway_model_events` and `request_logs_ch` (`credits_mc Int64 DEFAULT 0`, …). These
columns are written only when the startup validation sees them, and are otherwise
skipped with a warning, so the deploy order is not a hard outage.

## Product Attribution

`product` derivation, first match wins:

1. Guardrail judge calls → `guardrail`.
2. Semantic-cache embedding calls → `semantic_cache`.
3. The request's auth key has on-behalf-of capability → the key's configured product
   (default `librechat`).
4. Playground (`x-godarai-playground-observability` present) → `playground`.
5. Request type `embeddings` → `embeddings`. `ocr` / layout parsing → `ocr`.
6. Otherwise → `gateway`.

Phase 1 only *labels* events. Guardrail-judge and embedding side calls are not yet
priced as separate events. They are wired in Phase 3, where they become billable.

## Response Surface

- `x-godarai-credits-used: "4.50"` is set next to the existing cost headers when
  `credits.shadow_enabled` is on.
- The telemetry event gets `"credits": "4.50"`. The existing fields are unchanged.

## Service Identity (`x-godarai-on-behalf-of`)

Decision: implement this as a **capability on VAT keys** (`on_behalf_of_enabled` plus a
`product` label) rather than a brand-new key type. VATs are already the service-account
key class, with lifecycle, UI, and RBAC. A new type would duplicate all of that. The
contract still behaves as specified:

- Only keys with the capability may send the header. Anyone else gets 403
  `on_behalf_of_not_allowed`.
- The value is the GodarAI OIDC `sub`. It is resolved through the same directory lookup
  as master-key delegation (`ResolveTenantUserAuth`), scoped to the key's tenant.
- An unknown or inactive user gets 403 `on_behalf_of_user_not_found`, with no detail.

Open: confirm the format of the OIDC `sub` (prompt §8.2). It is checked against
`060_oidc_provider.sql` and the OIDC token issuer during implementation.

## `GET /v1/rate-card`

- Public. It is registered outside auth, like `/.well-known/*`.
- Response:

  ```json
  {"object":"rate_card","version":"rc-1","effective_from":"…","unit":"credit",
   "models":[{"model":"anthropic/claude-sonnet-5","provider":"anthropic",
     "rates":[{"dimension":"input_tokens","per":"1k_tokens","credits":"1.00"}, …],
     "typical_message_credits":"4.50"}]}
  ```

- Rates per 1K tokens are shown with up to 4 decimals (cache reads can be 0.1 or 0.15).
  Per call and per image are shown with 2 decimals.
- `typical_message_credits` uses 2,000 input and 500 output tokens with the base
  (unconditional) rates.
- Source: the anthropic + catalog snapshot managers (`List` accessor to add).
- `ETag` = hash of (rate-card version + provider revisions);
  `Cache-Control: public, max-age=300`.
- The authenticated tenant filter is deferred to Phase 3 (budget/me). Phase 1 serves
  the global catalog.

## Tenant-Admin Shadow View

- A billing-page card "this month in credits": the month-to-date sum of `credits_mc`,
  a top-5 by model, and a top-5 by team.
- Data comes from a new management endpoint `GET /api/v1/credits/shadow-summary` over
  ClickHouse `gateway_model_events`.
- Shown only when the server reports `shadow_enabled`.

## Calibration Report

- `cmd/credit-calibration`: a read-only ClickHouse query over 90 days.
- Outputs credits per (tenant, user, month) with the median, p75, and p95 per tenant and
  overall, plus the projected plan-quota share.
- Writes CSV and Markdown to `output/credit-calibration/`.

## Flags

| Flag | Default | Effect |
| --- | --- | --- |
| `BILLING_CREDITS_SHADOW_ENABLED` (env) | `false` | Compute and persist credits, set the header and telemetry field, show the shadow card. |
| `BILLING_CREDIT_RATE_CARD_TTL_SECONDS` | `300` | Rate-card cache TTL |

## Rollback

- Flag off: the request path is byte-identical to today (tested).
- Migrations are additive only. Down migrations drop the new columns and tables.

## Decisions Made

1. On-behalf-of is a VAT capability, not a new key type (see above).
2. Rate-card version = multiplier set + changelog. Provider prices stay versioned by
   `PricingRevision`.
3. The `float_fallback` source is allowed in shadow mode only.

## Open Issues

- OIDC `sub` ↔ LibreChat `OPENIDID` (prompt §8.2). It does not block Phase 1 math or
  shadow work.

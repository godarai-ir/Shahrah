# Phase 1 Design: Credit Core, Rate Card, Shadow Mode

Status: implemented · Branch: `feat/credits-phase-1-core`
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
3. The request was authenticated with an OIDC access token → the OIDC client's
   configured product (`librechat` or the default `oidc_app`), stamped into the token.
4. Playground (`x-godarai-playground-observability` present) → `playground`.
5. Request type `embeddings` → `embeddings`. `ocr` / layout parsing → `ocr`.
6. Otherwise → `gateway`.

Phase 1 only *labels* events. Guardrail-judge and embedding side calls are not yet
priced as separate events. They are wired in Phase 3, where they become billable.

## Response Surface

- `x-godarai-credits-used: "4.50"` is set next to the existing cost headers when
  `credits.shadow_enabled` is on.
- The telemetry event gets `"credits": "4.50"`. The existing fields are unchanged.

## Delegated Identity: Users' OIDC Access Tokens

Product decision: delegated callers such as LibreChat send **the signed-in member's
GodarAI OIDC access token** (`Authorization: Bearer <access_token>`). A shared service
key with an on-behalf-of header was considered and dropped. The gateway already
accepted OIDC access tokens (`internal/auth/access_token.go`):
- Tokens are tenant-bound and last 10 minutes; refresh tokens are 14-day and
  rotating, with the `offline_access` scope.
- Management APIs are blocked.
- The model policy is user-scoped, exactly like the member's own PAT.

What Phase 1 adds:
- `oidc_clients.credit_product` (migration 062; `oidc_app` default or `librechat`,
  check constraint), chosen by tenant admins in "Connected apps".
- The product is stamped into each access token at issuance (claim `gdp`), so the
  gateway hot path needs no client lookup. A change takes effect within one access-token
  lifetime. Tokens without the claim, or with an unknown value, fall back to `oidc_app`.
- `AuthContext.CreditProduct` carries it into spend logs as a `credits.TrustedProduct`,
  which client-supplied metadata cannot spoof.
- Usage is attributed to the user and tenant only. No team is set, by product
  decision. Team rules do not match OIDC traffic.

LibreChat side: set `OPENID_REUSE_TOKENS=true` and use `{{LIBRECHAT_OPENID_ACCESS_TOKEN}}`
as the custom endpoint key. The fork resolves the placeholder per user and handles
expiry and re-authentication.

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
| ~~`BILLING_CREDITS_SHADOW_ENABLED`~~ | — | Removed after the fleet cutover (2026-10-17): credits are always computed, because the ledger settles only spend logs that carry them. |
| `BILLING_CREDIT_RATE_CARD_TTL_SECONDS` | `300` | Rate-card cache TTL |

## Rollback

- Flag off: the request path is byte-identical to today (tested).
- Migrations are additive only. Down migrations drop the new columns and tables.

## Decisions Made

1. Delegated identity uses members' OIDC access tokens, not an on-behalf-of service key
   (see above). A first on-behalf-of implementation was reverted in this branch.
2. Rate-card version = multiplier set + changelog. Provider prices stay versioned by
   `PricingRevision`.
3. The `float_fallback` source is allowed in shadow mode only. A zero-cost request
   with no breakdown (cache hit, blocked request) is labelled `no_cost`, so it does
   not inflate the fallback ratio used by Gate 1.
4. Multiplier key = canonical `provider/model-id` (lower-case, `tenant:account:` prefix
   and duplicated provider prefix stripped). The DB check constraint enforces the
   normalized form.
5. Two credit grades:
   - **Ledger grade**, on `spend_logs`: computed from decimal cost components by
     `CreditShadowSpendLogger`.
   - **Analytics grade**, on ClickHouse `gateway_model_events`: computed from the
     event's `cost_usd` in the metrics emitter, which avoids touching ~20 metric
     call sites.

   They agree except for float rounding of `cost_usd`, and Phase 2 consume entries
   use the ledger grade. `request_logs_ch` is unchanged.
6. The ClickHouse credit columns are optional. The control plane detects them at
   startup and writes them only when present, so the rollout order is not a hard
   dependency.
7. The response credit surface is a `/v1` middleware that puts the rate-card
   provider in `fiber` Locals. That avoids changing the signature of
   `setCostHeaders` (17 call sites) and avoids global state. Stream handlers
   capture the provider before the body stream writer starts.
8. The tenant summary endpoint is `GET /api/v1/billing/credits/summary?preset=`. It
   reuses the Jalali `BillingReportWindow` presets and sits next to the existing
   billing reports, instead of the `/api/v1/credits/shadow-summary` path planned
   above.
9. OIDC `sub` is the GodarAI `user_id` (`internal/oidc/access.go`). Prompt §8.2 no
   longer applies, because the gateway reads the identity from the verified token.
10. Calibration months are calendar months in Asia/Tehran (not Jalali). Partial months
    are excluded.

## Measurements

- `BenchmarkFromBreakdown` (5 components plus a multiplier): about 16 µs/op, 72
  allocs. At most three computations per request (header, spend log, metric), so
  roughly 50 µs, well inside the 1 ms Gate 1 budget.

## Open Issues

- ClickHouse metric events still label OIDC usage `gateway`; the trusted product is on
  spend logs only. Fix this with FR-13 in Phase 3.
- The gitops `migrate-clickhouse` job must include `006_gateway_model_event_credits.sql`.
- Pre-existing, not addressed here: the `registerProxyRoutes` gocyclo limit (53 > 50),
  gofmt drift in `pricing_strategy_catalog_tokens.go`, tenant-admin cross-feature
  import violations, and duplicate `@types/react` copies breaking `tsc`.

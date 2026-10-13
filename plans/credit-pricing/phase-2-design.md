# Phase 2 Design: Ledger, Lots, Plans, Purchases, Expiry, Price Policy

Status: Phase 2 implemented (2a ledger core, 2b plans and user cap, 2c packs and overage, 2d price policy, changelog, and support-admin) · Branch: `feat/credits-phase-2-ledger` (stacked on `feat/credits-phase-1-core`)

Source: `plans/credit-pricing-execution-prompt.md` §6 Phase 2. Phase 1 design:
`plans/credit-pricing/phase-1-design.md`.

## Goal and Non-Goals

Phase 2 builds the full credit balance model. A tenant's credits can be granted, purchased,
consumed, and expired. Enforcement stays IRR for every tenant. Credit reservations, budget
rules in credits, and the cutover tool are Phase 3.

## What a "credit tenant" means in Phase 2

- `tenants.billing_unit` is added now (`irr` default, or `credit`), as invariant 7 asks.
- Only support-admin can set `credit`. The call is refused unless
  `BILLING_CREDIT_LEDGER_ENABLED=true`. The flag defaults to off, and Phase 2 enables it
  only in staging and for internal pilot tenants.
- For a credit tenant:
  - Grants, purchases, consumption, and expiry are written to the credit ledger.
  - **The IRR wallet settlement skips the tenant**, so it is never charged twice.
  - Gateway enforcement still runs on the IRR budget windows. A credit tenant therefore
    needs a non-zero IRR wallet balance to pass the gateway's wallet check until Phase 3.
- For an IRR tenant nothing changes. The workers skip it, and no lot or ledger row is
  ever written for it.

## Data Model (migration 063, PR 2a)

All amounts are `bigint` millicredits (1 credit = 1,000 mc). IRR amounts are
`numeric(18,0)`. The sign of `amount_mc` follows the effect on the balance: grants are
positive; consumption and expiry are negative.

```sql
ALTER TABLE tenants ADD COLUMN billing_unit varchar(16) NOT NULL DEFAULT 'irr'
  CHECK (billing_unit IN ('irr','credit'));

CREATE TABLE credit_lots (
  lot_id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants ON DELETE CASCADE,
  source varchar(32) NOT NULL CHECK (source IN
    ('plan','trial','promo','purchase_addon','migration','compensation','adjust','overage')),
  amount_mc bigint NOT NULL CHECK (amount_mc > 0),
  expires_at timestamptz,                  -- NULL never expires (adjust, migration)
  purchase_price_irr numeric(18,0),        -- purchases only
  price_per_credit_irr numeric(18,4),
  source_ref jsonb NOT NULL DEFAULT '{}',  -- subscription_id, period_start, order_id ...
  grant_entry_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now());

-- Append-only, partitioned monthly by occurred_at.
CREATE TABLE credit_ledger (
  entry_id uuid NOT NULL,
  tenant_id uuid NOT NULL,
  kind varchar(32) NOT NULL CHECK (kind IN
    ('purchase','plan_grant','promo_grant','consume','refund','expire','adjust')),
  amount_mc bigint NOT NULL,
  idempotency_key varchar(200) NOT NULL,
  occurred_at timestamptz NOT NULL,
  lot_id uuid,                             -- the lot a grant created or an expire closed
  -- consume only
  request_id varchar(255), trace_id varchar(255), model varchar(255),
  user_id uuid, team_id uuid, org_unit_id uuid,
  product varchar(32), billing_mode varchar(32),
  pricing_revision varchar(128), rate_card_version integer,
  internal_cost_usd varchar(64),           -- decimal string, never a float
  credit_equivalent_mc bigint,             -- BYOK: full equivalent; amount_mc is the fee
  spend_log_id bigint,
  -- manual kinds
  reason text, actor varchar(255),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (entry_id, occurred_at)) PARTITION BY RANGE (occurred_at);

-- A unique index on a partitioned table must contain the partition key, so the
-- idempotency guarantee lives in a small unpartitioned table written in the same tx.
CREATE TABLE credit_ledger_keys (
  idempotency_key varchar(200) PRIMARY KEY,
  entry_id uuid NOT NULL, tenant_id uuid NOT NULL, occurred_at timestamptz NOT NULL);

-- Append-only, partitioned like the ledger. lot_id NULL means unallocated debt.
CREATE TABLE credit_ledger_allocations (
  allocation_id uuid NOT NULL,
  entry_id uuid NOT NULL, tenant_id uuid NOT NULL,
  lot_id uuid, amount_mc bigint NOT NULL,
  occurred_at timestamptz NOT NULL,
  PRIMARY KEY (allocation_id, occurred_at)) PARTITION BY RANGE (occurred_at);

-- Projections: updatable, rebuildable from allocations.
CREATE TABLE credit_lot_balances (
  lot_id uuid PRIMARY KEY REFERENCES credit_lots ON DELETE CASCADE,
  tenant_id uuid NOT NULL, remaining_mc bigint NOT NULL CHECK (remaining_mc >= 0),
  expires_at timestamptz, source_rank smallint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE credit_tenant_balances (
  tenant_id uuid PRIMARY KEY, available_mc bigint NOT NULL, debt_mc bigint NOT NULL,
  ledger_sum_mc bigint NOT NULL, updated_at timestamptz NOT NULL);

CREATE TABLE credit_consume_cursor (            -- settlement progress over spend_logs
  cursor_id varchar(64) PRIMARY KEY, last_spend_log_id bigint NOT NULL, updated_at timestamptz);
```

- Triggers: `BEFORE UPDATE OR DELETE` (row) and `BEFORE TRUNCATE` (statement) raise on
  `credit_ledger`, `credit_ledger_allocations`, `credit_ledger_keys`, and `credit_lots`.
  This enforces invariant 3. A test proves each trigger.
  - Row UPDATE and DELETE have no bypass.
  - TRUNCATE is allowed only after `SET LOCAL shahrah.credit_ledger_maintenance = 'on'`.
    The integration suite's reset helper uses it. It also stops an accidental
    `TRUNCATE tenants CASCADE` from wiping lots and allocations.
- `credit_lots` and `credit_tenant_balances` reference `tenants` without cascade.
  Tenants are never hard-deleted in production, and a tenant with credit history cannot be.
- Partitions: the migration creates partitions for the previous, current, and next 3
  months, plus a DEFAULT partition. The expiry worker keeps 3 months of partitions ahead.
  A row in DEFAULT is a bug: the reconciliation job raises an alert for it.
- `spend_logs` gains `user_id` (the column already exists and was never written) and
  `org_unit_id`. The ledger needs both for consumption attribution.
- Retention: the ledger is never pruned in Phase 2. Partitions only make a later archive
  cheap. No finance or legal rule is known that conflicts with this (prompt §8.5).

### Invariants the projections hold

- `lot.remaining = Σ allocations(lot)`, `debt = −Σ allocations(lot IS NULL)`.
- `Σ ledger.amount_mc = Σ lot.remaining − debt = available − debt`.
- For each entry, `Σ allocations(entry) = entry.amount_mc`.

## Allocation and Settlement (PR 2a)

### Consumption source

Consumption comes from `spend_logs` rows with `credits_mc > 0` for credit tenants. Phase 1
already persists those rows asynchronously with the trusted product, billing mode, rate-card
version, and pricing revision. This keeps ledger writes off the request path
(invariant 5) and gives one ledger entry per request, as the schema asks.

Phase 3 replaces the IRR windows with millicredit windows for enforcement. Those windows
never write ledger rows, so the ledger cannot be charged twice.

```
spend_logs (id ↑) ──► CreditSettlementProcessor (advisory lock, batch N)
                         │ for each row of a credit tenant with credits_mc > 0
                         │   key = consume:{tenant}:{request_id}:{product}
                         │   tx: INSERT credit_ledger_keys ON CONFLICT DO NOTHING
                         │       ├─ conflict → skip (already settled)
                         │       └─ new → INSERT ledger entry
                         │                SELECT lot balances FOR UPDATE (allocation order)
                         │                INSERT allocations, UPDATE projections
                         └─ advance cursor (same tx as the last row of the batch)
```

- Allocation order (§3.1): `expires_at ASC NULLS LAST`, then source rank (plan = 0;
  trial, promo, and compensation = 1; purchase_addon = 2; everything else = 3), then
  `created_at`.
- Shortfall: the remainder becomes a debt allocation (`lot_id NULL`). Both overage and the
  estimate-versus-actual race end up here.
- Debt repayment: a new grant first pays outstanding debt. It writes, for the oldest debt
  entries, a `(entry, NULL, +x)` allocation and an `(entry, new_lot, −x)` allocation. This
  appends rows only and keeps `Σ allocations(entry)` unchanged.
- BYOK (`billing_mode = external_billed`): `credit_equivalent_mc = credits_mc` and
  `amount_mc = −ceil10(credits_mc × fee%)`. The fee comes from `credits.byok_fee_percent`
  (entitlement, then tenant override, then the 5% default).
- Cache hits have `credits_mc = 0` and create no entry.
- Spend logs without a `request_id` use the key `consume:{tenant}:log:{spend_log_id}:{product}`.
- The cursor starts at the newest spend log on first run, so usage from before the
  ledger existed is never charged in credits.
- Rows younger than `BILLING_CREDIT_SETTLE_DELAY_SECONDS` (default 30s) are left for the
  next pass, so a slow writer's lower id is not passed over.
- A daily sweep re-settles the last 48 hours behind the cursor as the safety net.
  Settlement is idempotent, so the sweep only adds what was missed, and it counts the
  rows it recovers (`shahrah_credit_settlement_recovered_total`).
- A row that fails stops the pass at that row: the cursor stops before it, and the
  failure is counted. This charges late rather than skipping a charge.
- Spend logs now record `user_id` and `org_unit_id`. The streaming `/v1/completions`
  spend log was also missing `organization_id` and `team_id`; this PR fixes that. An id
  that is not a UUID is stored as NULL, so it cannot fail the spend-log batch.
- A duplicate `(request_id, product)` is skipped and counted in
  `credit_settlement_duplicates_total`. That under-charges, which is the conservative
  choice. The design note on the report gives the measured count.
- Concurrency: a `pg_try_advisory_xact_lock` serializes settlement workers. Lot rows are
  locked `FOR UPDATE` per tenant, in the same order as the expiry job, so neither can
  deadlock the other.

### Expiry (PR 2a)

- For each lot with `expires_at <= now` and `remaining_mc > 0`: write `expire` with key
  `expire:{lot_id}`, an allocation of `−remaining` to that lot, and set the projection to 0.
- Safe to run twice and concurrently (the key plus `FOR UPDATE SKIP LOCKED` on the lot).
- In-flight reservations do not exist in credits until Phase 3. A consumption that
  arrives after expiry is allocated to the next lot, or to debt.
- After each tenant it invalidates the cached balance (`credit:balance:{tenant}` in Redis).

### Reconciliation (PR 2a)

- Daily, per credit tenant: compare `Σ ledger`, `Σ lot projections − debt`, and
  `credit_tenant_balances`, and check each entry's allocation sum. Also check that the
  DEFAULT partitions are empty.
- Any difference: an error log with the tenant and amount, plus
  `credit_reconciliation_diff_mc{tenant}` (gauge) and `credit_reconciliation_failures_total`.
  Nothing is auto-fixed.
- Nightly rebuild-and-compare: recompute the projections from allocations in a read-only
  transaction and report differences the same way.

### Grants

Grant helper: `Grant(ctx, tenant, source, amount_mc, expires_at, key, ref)` writes the
key, the ledger entry, the lot, the lot's own `+amount` allocation, the debt repayment,
and the projections, all in one transaction.

| Grant | Kind / lot source | Idempotency key |
| --- | --- | --- |
| Plan period slice | `plan_grant` / `plan` | `plan_grant:{subscription_id}:{slice_start}` |
| Trial | `promo_grant` / `trial` | `trial_grant:{subscription_id}` |
| Add-on purchase | `purchase` / `purchase_addon` | `purchase:{order_id}` |
| Overage settlement | `purchase` / `overage` | `overage:{subscription_id}:{period_start}` |
| Manual adjust | `adjust` / `adjust` (if positive) | the caller's key, required |

## Plans as Data (PR 2b)

- **Existing data:** v1 of developer, team, growth, and enterprise is published, with
  `amount_minor NULL` (unpriced). `users.max` is already 3/10/30/null and is enforced
  (without a lock, and without counting invitations). There is no `business` plan.
- **The migration seeds drafts only.**
  - v2 drafts of developer, team, and growth, plus a v1 draft of a new `business` plan,
    with IRR prices (×10 from toman) and the credit entitlements. All other entitlements
    are copied from v1.
  - A support-admin publishes them through the existing `PublishPlanVersion`.
  - Publishing changes the price new purchases pay, so a human does it, not a migration
    (prompt §8.4).
  - A unit test fails if any seeded price is off by 10×.

| Plan | Users | Monthly IRR | Credits / month |
| --- | --- | --- | --- |
| developer | 3 | 50,000,000 | 5,000 |
| team | 10 | 150,000,000 | 16,500 |
| growth | 30 | 450,000,000 | 54,000 |
| business | 100 | 1,500,000,000 | 190,000 |
| enterprise | override | override (≥ 7,500 IRR/credit) | override |

- Intervals: `monthly`, and `yearly` with a 10% discount (the PRD maximum). Quarterly has
  no price in v2, so it is not offered. The v1 seed's yearly 20% does not carry over.
- New `feature_catalog` keys:
  - `credits.monthly_grant` (integer credits).
  - `credits.bonus_percent` (decimal, display only).
  - `credits.overage_allowed`, `credits.rollover_allowed` (boolean).
  - `credits.byok_fee_percent` (decimal, default 5).
  - `credits.trial_grant` (integer, 500), `credits.trial_days` (integer, 14).
  - `users.max` is reused for `users.max_active`. It already has that meaning, so there
    is no second key to keep in sync.
- Plan grant worker (runs with expiry, every `BILLING_CREDIT_EXPIRY_INTERVAL_SECONDS`):
  - For each current subscription of a credit tenant, it grants the monthly slice that
    contains `now`. Slices are calendar months from `current_period_start`, so annual
    plans deliver monthly. Only `active` subscriptions get plan credits; `past_due` does
    not, and `trialing` gets the trial grant instead.
  - The monthly amount is `credits.monthly_grant`: a tenant override first, then the
    subscription's plan version.
  - Each slice lot expires at the earlier of the slice end and the period end.
  - Subscriptions do not auto-renew today. A renewal or plan change is a new period,
    and the same worker covers it.
- Trial:
  - `commercial_signup_settings` gains `trial_credit_mc` (500,000), `trial_credit_days`
    (14), and `trial_max_users` (3).
  - Trial credits expire at the trial's end, capped at `trial_credit_days` after it started.
  - `trial_max_users` caps trialing **credit** tenants only. Trials of IRR tenants keep
    their plan's cap, so no existing tenant's quota changes (prompt §8.4).
  - `signup_credit_irr` stays for IRR tenants until Phase 4.
- Rollover (Enterprise only) is **deferred to Phase 3**. No Enterprise tenant is on
  credits in Phase 2, and the condition ("renews at the same or higher tier") is only
  known at renewal. Open issue below.

## User Cap (PR 2b)

- Count active human members plus pending, unexpired invitations. Virtual users, PATs,
  and service keys are excluded. This applies to every tenant, so IRR tenants now also
  count pending invitations and get the race-free check.
- Check it inside the same transaction as the write, under
  `pg_advisory_xact_lock(hashtext('users_cap:' || tenant))`, at every entry point:
  - invite
  - invitation accept
  - user reactivate
  - direct create
  - SSO/OIDC just-in-time provisioning, if that creates members
- Error: 403 `{"error":{"code":"plan_user_limit_reached"}}`, replacing the generic
  "feature users.max limit exceeded" message on these paths. The message carries no
  counts. `GET /api/v1/users/seats` (`user:list`) reports used, active, pending, and the
  limit.
- Entry points covered:
  - invite
  - invitation accept (the pending invitation already holds the seat)
  - activate
  - role assignment, which also reactivates a deactivated member
- New-tenant onboarding and social sign-up create the first member of a new tenant and
  need no check. There is no other create-member path.
- Downgrades below the current count are refused in `PrepareTenantSubscriptionChange`
  and `ChangeTenantSubscription`.
- Tenants already over the cap are grandfathered: nothing is removed. They simply cannot
  add users.
- Tenant-admin shows used/max seats on the users page. It reads the seats endpoint, so
  pending invitations count there too. The invite button opens the upgrade dialog when the
  plan is full, and refused invites, activations, and role changes show a localized message.
- Support-admin `UpsertTenantSubscription` is not guarded: it is an operator override.

## Add-on Packs and Overage (PR 2c)

- `credit_packs` (migration 065): a versioned catalog with slug, version, credits,
  `discount_percent` against the list price, `validity_months` = 12, and a status.
  Only one version per slug can be published.
  - Seeded and published: `credits-1000` at 0%, `credits-5000` at 10%, and
    `credits-20000` at 20%. At 10,000 IRR per credit these reproduce the PRD prices:
    10,000,000, 45,000,000, and 160,000,000 IRR. A test checks them against the toman
    prices.
  - Publishing changes nobody's price, because only credit tenants can buy packs.
- `credit_pack_orders` follows the IRR-native shape of `subscription_payment_intents`.
  - Each order snapshots the pack version, list price, discount, `price_irr`, VAT
    (`wallet_topup_settings.tax_percent`, as top-ups use), and `payable_irr`.
  - A check constraint enforces `payable = price + tax`.
  - The list price is the launch constant until 2d replaces it with the active revision.
- Gateway flow:
  - `POST /api/v1/billing/credits/orders` returns the gateway URL.
  - The public callback `/api/v1/credits/packs/payment/callback/:provider` verifies the
    payment **with the order's own provider**. The provider in the URL is ignored, so a
    zarinpal order cannot be "verified" by the mock gateway.
  - Completion grants the lot under key `purchase:{order_id}` in the same transaction, so
    a repeated callback grants nothing.
  - A completed order is never marked failed afterwards.
- The mock gateway approves every payment. Credit packs accept it only where it is the
  configured default provider, that is, in local and test environments. Wallet top-ups and
  subscriptions accept it from any request; that is flagged as a separate task.
- Bank transfer:
  - An order with `provider = bank_transfer` waits in `AWAITING_TRANSFER`.
  - `POST /wallet/bank-transfers` takes `credit_pack_order_id`, and the amount must equal
    the order's payable amount exactly.
  - Approval completes the order and grants the lot instead of crediting the IRR wallet.
  - A unique index allows only one approved transfer per order.
  - Rejection leaves the order open for another transfer.
- Every purchase writes `financial_audit_events` (`credit_pack.purchased`).
- Other tenant endpoints: `GET /billing/credits/packs`, `GET /billing/credits/orders[/:id]`,
  `GET` and `PUT /billing/credits/settings`. Errors carry stable codes, for example
  `credit_billing_not_enabled`, `credit_pack_transfer_amount_mismatch`, and
  `overage_not_allowed`.
- Overage:
  - `credit_tenant_settings` holds `overage_enabled` (default false) and an optional
    `overage_cap_mc`. Overage can be turned on only for credit tenants whose plan allows
    it (`credits.overage_allowed`). Every change is audited.
  - When a new plan slice is due, the grant transaction first checks for outstanding debt
    of an overage-enabled tenant. It invoices that debt in IRR at the list price, rounded
    up to the rial, with an `OVG-yyyymm-sub` invoice and a `credit_overage` line item. An
    `overage` lot, fully allocated to the debt under key
    `overage:{subscription}:{slice_start}`, clears it.
  - The new slice is then granted in full.
  - Without overage, the new slice repays the debt instead.
  - Enforcing the cap is Phase 3, because Phase 2 has no credit enforcement.
- Tenant-admin screens for packs and overage are Phase 3 (FR-13's plan page). 2c ships the
  APIs only.

## Credit Price Policy, Changelog, Support-Admin (PR 2d)

- FX samples (migration 066):
  - Every hour the worker stores the rate the pricing engine uses at that moment (the
    newest `exchange_rate_current` row) in `exchange_rate_samples`. This is the existing
    pipeline, not a new FX source (prompt §8.7).
  - A rate older than 6 hours is skipped, so a stalled feed leaves a gap instead of
    repeating one value.
  - The 7-day average uses the 7 complete UTC days before now. It requires at least one
    sample on each day.
- `credit_price_revisions`:
  - The seeded active revision is 1,000 toman at a 270,000 toman/USD base.
  - Statuses are `proposed`, `active`, `rejected`, and `superseded`. The prompt's
    "approved" collapses into `active`, because add-on prices change at approval.
  - Unique indexes allow one active and one open proposal.
  - When the average leaves ±10% of the base, the worker proposes
    `round50(price × avg / base)` with the average as the new base. The PRD example
    (300k → 1,100) is a test. An open proposal is refreshed, not duplicated.
  - `BILLING_CREDIT_PRICE_AUTO_APPLY` (default off) approves it as `price_policy`.
- Approval (support-admin, reason required, audited):
  - The old price becomes `superseded` and the new one `active` at once, so pack quotes
    use it immediately.
  - Every published credit plan gets a draft version with prices scaled from the list
    price that version was priced at. The draft's `effective_from` is approval + 30 days,
    and it is linked by `plan_versions.credit_price_revision_id`.
  - The worker publishes those drafts when due, and only while their revision is still
    active, so subscribers pay the new price from their next period.
  - Rejection closes the proposal.
- Public changelog: `GET /v1/pricing/changelog` (public, cached 5 minutes) lists rate-card
  versions (`scheduled`, `published`, `superseded`) and decided prices. It never includes a
  USD figure.
  - `apps/website` renders it at `/pricing/changelog` (RTL, Persian dates), linked from
    the pricing page and the sitemap.
  - Tenant-admin links to it from the credit card on the billing page.
- Rate cards:
  - Drafts copy the published multipliers, and there is one draft at a time.
  - Every multiplier needs a reason and at most 4 decimals. Models are normalized.
  - Publishing needs a Persian changelog. If any model gets more expensive than in the
    published card, the version is `scheduled` for 30 days (`notice_until`) and the worker
    publishes it then; otherwise it publishes at once. A newer due schedule replaces an
    older one.
- Packs: support-admin drafts a new version of a slug and publishes it, which archives the
  previous one. Orders keep their snapshot.
- Tenant credits (support-admin):
  - Overview (unit, available, debt, overage), lots with a per-lot movement drill-down,
    and a paged ledger.
  - Manual adjustments, signed with two decimals, with a reason and the caller's
    idempotency key, audited. A positive one creates an `adjust` lot (optional expiry); a
    negative one is spent from lots, with shortfall as debt.
  - The billing-unit switch needs a reason and is refused unless
    `BILLING_CREDIT_LEDGER_ENABLED` is on.
- Operator permissions in the support-admin proxy:
  - Reads need `billing:read` or `wallet:read`.
  - Adjustments need `wallet:adjust`.
  - Pack and rate-card changes need `catalog:write`.
  - Price decisions and the billing unit need `platform:admin`.
- The plan catalog stays in the existing support-admin "Pricing" page.

## Feature Flags

| Flag | Default | Effect |
| --- | --- | --- |
| `BILLING_CREDIT_LEDGER_ENABLED` | false | Starts the settlement, grant, expiry, and reconciliation workers, and allows `billing_unit=credit`. |
| `BILLING_CREDIT_SETTLEMENT_INTERVAL` | 30s | Consume settlement cadence. |
| `BILLING_CREDIT_PRICE_AUTO_APPLY` | false | Applies an FX-band proposal without approval. |
| `BILLING_CREDIT_PRICE_INTERVAL_SECONDS` | 3600 | FX sampling, price proposals, and scheduled plan and rate-card publishing. |

## Rollout and Rollback

1. Merge with the flag off. The migrations are additive, and IRR tenants are untouched.
2. Staging: flag on, flip internal tenants to `credit`, and run Gate 2.
3. Rollback:
   - Flip tenants back to `irr` and turn the flag off.
   - The ledger stays as history.
   - For pilot tenants, the IRR wallet resumes settlement from the current window state.
     Their usage while in credit mode is not re-charged in IRR, which is conservative.

## Decisions Made (conservative defaults; call them out in the report)

1. Phase 2 credit tenants: ledger only, with IRR wallet settlement skipped, as described above.
2. Consumption is settled from `spend_logs`, not from budget windows.
3. Ledger idempotency lives in `credit_ledger_keys`, because partitioned tables cannot
   carry a unique index without the partition key.
4. Duplicate `(request_id, product)` spend logs are skipped (this under-charges) and counted.
5. Plan seeds are drafts; a human publishes them.
6. Yearly discount is 10%, and quarterly is not offered in v2.
7. `users.max` is reused as the active-user cap.
8. Add-on pack prices are a discount on the active list price, so an approved price
   revision moves them automatically.
9. Rollover is deferred to Phase 3.
10. FX samples come from the pricing engine's existing source.
11. Business copies Growth's non-credit entitlements, with `users.max` set to 100.
12. On a mid-period plan change, the old plan's lot is kept until its own slice ends.
    This is conservative: it charges less.

## Open Issues

- Price-change notices are announced through the public changelog only. Direct notices to
  tenant admins (`notice_sent_at`) need an outbound notification channel, which prompt §8.1
  says to ask about in Phase 3.

- Tenant admins cannot revoke a pending invitation, so a mistaken invitation holds a seat
  until it expires after 7 days. A revoke action is a small follow-up.

- Rollover rules for Enterprise (Phase 3).
- The FX adjustment clause for annual plans is contract text; finance and legal own it.
- Finance must confirm that VAT on add-on packs uses `wallet_topup_settings.tax_percent`,
  as top-ups do (Gate 2 asks for sign-off on a sample invoice).
- Overage invoices carry no VAT line, like the existing period-end invoices. Finance
  should confirm this with the Gate 2 sample invoice.
- Debt of a tenant whose subscription lapses without renewal is never invoiced, because
  no new slice is due. It stays as debt in the ledger.
- None of the open issues falls under prompt §8, so implementation proceeds.

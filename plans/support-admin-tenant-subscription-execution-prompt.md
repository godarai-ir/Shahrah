# Execution Prompt: Tenant Subscription Management in Support Admin

Use this prompt with a coding agent that has full access to the GodarAI repository.

---

You are a senior product-minded full-stack engineer implementing **Tenant Subscription Management** for GodarAI/Shahrah's internal support-admin application.

## Mission

Turn `apps/support-admin` into the central operational surface for finding, understanding, and safely managing SaaS tenants. An authorized operator must be able to inspect a tenant's commercial and operational state, adjust its wallet through an auditable ledger entry, manage its subscription and discounts, review usage and billing history, see important events, and suspend or reactivate access.

This is an internal privileged workflow. Correct authorization, tenant isolation, financial integrity, auditability, and safe failure behavior are more important than implementation speed.

## Repository Context You Must Inspect First

- Repository root: `D:\source\Shahrah`
- Support application: `apps/support-admin` (Next.js App Router)
- Control plane: `services/control-plane` (Go/Fiber)
- Shared contracts: `packages/contracts/platform`
- Main database migrations: `db/main/migrations`
- Existing support-admin proxy: `apps/support-admin/app/api/admin/control-plane/[...path]/route.ts`
- Existing commercial-plan UI: `apps/support-admin/app/admin/pricing/page.tsx`
- Existing plan/subscription contracts: `packages/contracts/platform/commercial_plans.go`
- Existing plan handlers/repository:
  - `services/control-plane/internal/admin/handlers_commercial_plans.go`
  - `services/control-plane/internal/storage/commercial_plan_repo.go`
  - `services/control-plane/internal/storage/subscription_payment_repo.go`
- Existing wallet repository: `services/control-plane/internal/storage/wallet_repo.go`
- Existing tenant support projection: `apps/support-admin/prisma/schema.prisma`
- Existing auth/runbook: `docs/support-admin-auth-runbook.md`
- Product references:
  - `plans/01_iranian_payment_system.md`
  - `plans/vat-operationalization-requirements.md`

Before editing, read `apps/support-admin/AGENTS.md` and the local Next.js documentation under `node_modules/next/dist/docs/` for every App Router API you use.

## Product Decisions

1. Add a first-class **Tenants** workspace. Do not overload the existing Pricing page; Pricing manages the global catalog, while Tenants manages customer instances.
2. Use a tenant 360-degree detail view with these tabs:
   - Overview
   - Subscription
   - Usage
   - Wallet & Billing
   - Events
   - Account Controls
3. Treat wallet operations as append-only ledger adjustments. Never expose direct balance editing.
4. Model promotional discount codes independently from `plan_prices.discount_percent`. The current field is a catalog price attribute, not a redeemable coupon system.
5. Require a human-readable reason for every privileged mutation. High-risk actions also require an explicit confirmation step.
6. Keep tenant source-of-truth data in control-plane. The support-admin database may hold authentication, sessions, projections, and UI read models, but it must not become authoritative for tenant subscription, wallet, usage, billing, or status.
7. All money values use integer minor units in API and persistence contracts. IRR has zero fractional minor units for this product. Do not use JavaScript floating-point arithmetic or Go `float64` for new financial writes.
8. Dates are stored and transported in UTC and displayed in the operator's configured timezone. Make period boundaries explicit.

## MVP Scope

### Include

- Searchable, filterable, cursor-paginated tenant list.
- Tenant summary and operational health indicators.
- Current subscription, effective entitlements, period dates, pending change, and subscription history.
- Immediate plan assignment/change by authorized operator with a required reason.
- Schedule cancellation at period end and reverse a scheduled cancellation.
- Current-period usage versus entitled limits.
- Wallet balance, pending balance, and paginated transaction history.
- Manual wallet credit and debit adjustment with reason, reference, idempotency key, and audit trail.
- Billing/payment-intent history and status.
- Discount-code CRUD plus assign/remove actions, with validity and redemption constraints.
- Important tenant event timeline.
- Suspend/reactivate tenant with reason and confirmation.
- Operator-facing audit trail for all mutations.
- Loading, empty, error, stale-data, and permission-denied states.

### Defer

- Refund execution through a payment gateway.
- Credit notes and tax-compliant invoice accounting.
- Proration and backdated plan changes.
- Automatic dunning and collections workflows.
- Multiple concurrent subscriptions per tenant.
- Bulk tenant mutations.
- Impersonation or login-as-tenant.

### Explicitly Exclude

- Direct SQL/database access from support-admin to the main tenant database.
- Editing a wallet balance in place.
- Deleting financial transactions, billing events, subscription events, or audit records.
- Exposing payment gateway raw payloads, credentials, tokens, prompt bodies, or response bodies in the UI.
- Using `SupportTenantProjection` as mutation source of truth.

## Roles and Permissions

Replace the coarse `platform:admin` behavior for this domain with explicit permissions:

- `tenant:read`
- `tenant:status:write`
- `subscription:read`
- `subscription:write`
- `usage:read`
- `wallet:read`
- `wallet:adjust`
- `billing:read`
- `discount:read`
- `discount:write`
- `audit:read`

Default role policy:

- `support_agent`: tenant/subscription/usage/wallet/billing/event read access only.
- `catalog_admin`: existing catalog permissions; no tenant financial or status mutations by default.
- `super_admin`: all permissions.

Enforce permissions twice: in support-admin route handlers for UX/security and in control-plane service-token authorization scopes as the authoritative boundary. A UI-hidden button is not authorization.

## UX Requirements

### Tenant List: `/admin/tenants`

- Search by tenant name/slug, domain, tenant UUID, and known administrator email if supported without cross-tenant leakage.
- Filters: tenant status, plan, subscription status, low/negative wallet, usage threshold, and last activity range.
- Columns: tenant, status, plan, subscription status, wallet balance, current-period request usage, last activity, important alert count.
- Use cursor pagination and URL-backed filters. Do not fetch all tenants into the browser.
- Show projection freshness. Provide a safe source-of-truth refresh action if projection lag exceeds the defined threshold.

### Tenant Detail: `/admin/tenants/[tenantId]`

Header includes display name/slug, tenant ID copy action, account status, plan, wallet balance, last activity, projection freshness, and allowed quick actions.

Overview cards:

- Account and subscription status.
- Current plan and renewal/period end.
- Wallet balance and last adjustment.
- Current-period requests, tokens, and cost when data is available.
- Active users/keys/providers where existing stores can supply them efficiently.
- Open alerts: past due, low balance, quota near limit, failed payment, suspended account.

Do not invent data. Show `Unavailable` with a clear explanation when a metric does not exist.

### Subscription Tab

- Show current plan/version, billing interval, price snapshot, status, current period, cancel-at-period-end, pending version, effective entitlements, overrides, and recent subscription events.
- Plan change dialog previews old versus new plan, interval, price, entitlements, timing, and whether payment is bypassed by an operator action.
- MVP operator changes are immediate and non-prorated. Require a reason and confirmation.
- Cancellation is scheduled at period end; do not immediately erase access unless suspension is chosen separately.

### Usage Tab

- Show current billing period and comparison to the previous equivalent period.
- Required metrics when supported: requests, input tokens, output tokens, total tokens, cost USD, charged IRR, error rate, and top models.
- Usage limits come from resolved entitlements. Show used, limit, remaining, and percentage.
- Clearly distinguish metered usage from wallet transactions. Do not calculate authoritative billing totals in the browser.
- Use existing metrics/trace stores if appropriate; add bounded aggregate queries rather than scanning raw traces.

### Wallet & Billing Tab

- Display available and pending wallet balance.
- Paginated wallet ledger with type, signed amount, status, description, operator/reference, and timestamp.
- Manual adjustment dialog:
  - operation: credit or debit
  - positive `amount_minor`
  - required reason
  - optional external reference/ticket ID
  - client-generated idempotency key
  - preview resulting balance
  - explicit confirmation
- Reject a debit that would violate the chosen negative-balance policy. The default is no negative balance.
- Billing history includes subscription payment intents with status, amount, discount, provider, safe payment reference, failure reason, and completion time.

### Discount Codes

Add a global management page at `/admin/discounts` and tenant assignment UI in the subscription tab.

A discount code supports:

- normalized unique code
- description
- discount type: `percent` or `fixed_amount`
- percent basis points or fixed integer amount and currency
- valid-from and valid-until
- enabled flag
- global maximum redemptions
- per-tenant maximum redemptions
- optional eligible plan slugs and billing intervals
- optional minimum purchase amount
- created-by and timestamps

Rules:

- Snapshot the applied discount into the payment/subscription financial record.
- Never recompute historical amounts when a code changes.
- One active code per tenant subscription in MVP; do not stack discounts.
- Validate eligibility and redemption atomically in control-plane.
- Store normalized codes but preserve a display form if needed.

### Events Tab

Provide a single reverse-chronological timeline containing important, sanitized events:

- tenant created, suspended, reactivated
- subscription assigned/changed/cancellation scheduled/cancellation reversed/expired
- payment initiated/completed/failed
- wallet manual credit/debit and payment top-up
- discount assigned/removed/redeemed
- entitlement override changed
- quota exhausted, low balance, or other existing material platform alerts

Each event shows type, severity, timestamp, actor, reason, safe metadata, and correlation/request ID. Use typed event names and a stable schema; do not expose raw provider payloads.

### Account Controls Tab

- Suspend account with a mandatory reason and explicit confirmation containing the tenant slug.
- Reactivate with a mandatory reason.
- Suspension must block new gateway traffic and tenant-admin authenticated operations consistently, while leaving support-admin read access and billing records available.
- Define status transition rules. MVP: `active -> suspended -> active`; do not reuse subscription status for account status.
- Existing in-flight inference requests may finish; new requests after the committed status change must fail with a stable error code.

## Backend and Contract Work

Add typed platform DTOs for all new endpoints. Follow existing snake_case JSON conventions.

Create platform endpoints under `/api/v1/platform`:

- `GET /tenants?query=&status=&plan=&subscription_status=&wallet_health=&cursor=&limit=`
- `GET /tenants/:tenantID`
- `POST /tenants/:tenantID/refresh-projection` if a refresh mechanism is necessary
- `GET /tenants/:tenantID/subscription`
- `POST /tenants/:tenantID/subscription/change`
- `POST /tenants/:tenantID/subscription/cancel-at-period-end`
- `DELETE /tenants/:tenantID/subscription/cancel-at-period-end`
- `GET /tenants/:tenantID/usage?from=&to=&granularity=`
- `GET /tenants/:tenantID/wallet`
- `GET /tenants/:tenantID/wallet/transactions?cursor=&limit=`
- `POST /tenants/:tenantID/wallet/adjustments`
- `GET /tenants/:tenantID/billing/payment-intents?cursor=&limit=`
- `GET /tenants/:tenantID/events?types=&severity=&cursor=&limit=`
- `POST /tenants/:tenantID/suspend`
- `POST /tenants/:tenantID/reactivate`
- `GET /discount-codes?query=&status=&cursor=&limit=`
- `POST /discount-codes`
- `PATCH /discount-codes/:discountCodeID`
- `POST /tenants/:tenantID/discount-code`
- `DELETE /tenants/:tenantID/discount-code`

Endpoint details:

- Use opaque cursor pagination with a deterministic tie-breaker.
- Validate UUIDs, enums, ranges, dates, currency, and request size.
- Mutations accept `reason` and `idempotency_key`; return the authoritative updated resource plus a correlation ID.
- Use transactions and row locks for wallet adjustments, discount redemption, and conflicting subscription changes.
- Return stable machine-readable error codes alongside safe messages.
- Never trust actor IDs supplied in a JSON body. Derive actor and role from authenticated headers/context set by support-admin.

## Data Model

Create the next sequential migration under `db/main/migrations` and mirror it where this repository's local Kubernetes migration process requires it. Prefer extending current tables only when semantics match.

Add or evolve these concepts:

1. Tenant status integrity
   - constrain supported status values without breaking existing data
   - add status-change timestamps if useful

2. Wallet adjustment idempotency and audit metadata
   - keep `wallet_transactions` append-only
   - add a unique operator adjustment/idempotency record or constrained metadata columns
   - retain balance update and ledger insert in the same transaction

3. Discount codes
   - `discount_codes`
   - `tenant_discount_assignments`
   - `discount_redemptions`
   - database constraints for discount shape, validity, non-negative limits, uniqueness, and one active tenant assignment

4. Platform audit/events
   - reuse `subscription_events` only for subscription-domain events
   - introduce a general append-only `tenant_events` or `platform_audit_events` table for cross-domain tenant operations
   - include actor ID/role, action, tenant ID, reason, request/correlation ID, sanitized before/after summaries, and timestamp

Do not duplicate the complete tenant source of truth in support-admin Prisma. Extend `SupportTenantProjection` only with fields necessary for fast list filtering and freshness, and update it through an explicit projection/sync mechanism.

## Financial Integrity Rules

- New write contracts use integer `amount_minor`.
- A manual wallet adjustment is atomic: lock wallet row, validate policy, insert immutable ledger transaction, update wallet, insert audit event, commit.
- Idempotency scope is tenant plus operation plus idempotency key. A repeated identical request returns the original result; a conflicting payload returns `409`.
- Never silently convert currency. If conversion is needed, persist the exchange-rate snapshot and source.
- Existing float-backed wallet fields are technical debt. Do not perform a risky global migration without a compatibility plan and tests; isolate conversion at existing boundaries and document it.
- Store before/after balances on the adjustment result or audit summary for reconciliation.

## Suspension Enforcement

Audit all request-entry paths. Ensure tenant status is checked in shared tenant/auth middleware for:

- OpenAI-compatible gateway endpoints
- tenant-admin APIs and session refresh
- virtual keys/PAT/VAT paths
- background operations that initiate tenant-billable work

Use cache invalidation or a bounded short TTL so suspension takes effect promptly across replicas. Return a stable `tenant_suspended` error without leaking the suspension reason externally.

## Support-Admin Implementation

- Add Tenants and Discounts to navigation.
- Prefer React Server Components for initial reads where practical; use client components only for interactive filters/dialogs.
- Route every browser request through authenticated same-origin Next route handlers.
- Extend the control-plane proxy to preserve query strings; allow only known methods and safe response headers.
- Do not expose `CONTROL_PLANE_SERVICE_TOKEN` to client code.
- Add a reusable confirmation dialog, money formatter, status badge, pagination controls, permission-aware action wrapper, and error presentation if equivalents do not exist.
- Maintain keyboard accessibility, labeled form controls, focus management, and visible destructive-action language.
- Avoid optimistic UI for financial, subscription, discount, and suspension mutations. Render the authoritative response and revalidate affected reads.

## Audit and Observability

- Every mutation emits an audit event in the same transaction when possible.
- Include `x-request-id`/correlation ID from support-admin through control-plane logs and responses.
- Add structured logs and metrics for endpoint latency, mutation outcome, permission denial, idempotency replay, wallet adjustment totals by direction (not tenant label), suspension count, and discount redemption failures.
- Do not use tenant ID, discount code, actor email, or reason as unbounded metric labels.
- Redact sensitive fields in logs and events.

## Required Tests

### Go Unit/Integration

- Tenant list/detail isolation and cursor pagination.
- Permission enforcement for every read and mutation family.
- Wallet credit/debit atomicity, insufficient balance, concurrent updates, idempotent retry, and conflicting idempotency payload.
- Subscription change, schedule cancellation, reverse cancellation, invalid transitions, and event creation.
- Discount date, eligibility, plan/interval, maximum-redemption, concurrent-redemption, no-stacking, and snapshot behavior.
- Suspend/reactivate transitions and immediate blocking of new gateway and tenant-admin requests.
- Audit event creation and sensitive-field sanitization.
- Cross-tenant access attempts return not found or forbidden according to the established policy without leaking existence.

### Support-Admin

- Permission matrix and proxy authorization.
- Query-string forwarding and route encoding.
- Tenant list filtering/pagination.
- Tenant detail loading, empty, stale, error, and forbidden states.
- Mutation dialogs require reason and confirmation.
- Financial forms reject invalid precision, amount, currency, and debit-over-balance.
- Successful mutations render authoritative server results and refresh affected panels.

### Regression

- Existing commercial plan, entitlement, tenant subscription payment, wallet top-up, billing debit, auth, and support-admin tests still pass.
- Existing tenant-admin subscription and wallet flows remain compatible.

## Delivery Phases

Implement in this order and leave every phase deployable:

1. Contracts, read-only tenant APIs, permission model, tenant list/detail UI.
2. Usage, wallet ledger, billing history, and unified tenant events as read-only surfaces.
3. Audited manual wallet adjustments.
4. Subscription operator actions and cancellation scheduling.
5. Suspension/reactivation plus enforcement at all entry points.
6. Discount-code model, management, assignment, redemption, and price snapshots.
7. Hardening, concurrency tests, observability, runbook, and UI accessibility pass.

If a required source metric or business rule does not exist, do not fabricate it. Implement a typed `unavailable` state, document the gap, and continue with the remaining deployable scope.

## Verification Commands

Determine exact package commands from repository manifests, then run at minimum:

```powershell
go test ./services/control-plane/...
pnpm --filter @shahrah/support-admin lint
pnpm --filter @shahrah/support-admin typecheck
pnpm --filter @shahrah/support-admin build
```

Also run targeted integration tests for new migrations and concurrent financial mutations. Render and inspect the principal support-admin pages at desktop and narrow widths using the available browser workflow.

## Definition of Done

- Authorized operators can find a tenant and understand its account, subscription, wallet, billing, usage, and recent important events from one tenant workspace.
- All privileged mutations are reasoned, confirmed, idempotent where applicable, audited, and authorized in control-plane.
- Wallet adjustments preserve ledger and balance consistency under retries and concurrency.
- Suspension reliably blocks new tenant activity without deleting tenant data or financial history.
- Discounts are validated atomically and snapshotted into historical financial records.
- No cross-tenant data leakage or service-token exposure exists.
- Migrations apply cleanly and safely to existing data.
- New and regression tests pass.
- Operator runbook documents permissions, wallet adjustment/reconciliation, suspension/reactivation, discount troubleshooting, projection freshness, and rollback considerations.

## Required Final Report

Return:

1. Product behavior delivered, grouped by delivery phase.
2. Files and migrations changed.
3. API endpoints and permission matrix.
4. Financial and concurrency invariants implemented.
5. Test, lint, typecheck, build, migration, and browser-verification results.
6. Screenshots or concise visual verification notes for the tenant list and every tenant-detail tab.
7. Remaining gaps, each labeled `BLOCKER`, `DEFERRED`, or `FOLLOW-UP`.

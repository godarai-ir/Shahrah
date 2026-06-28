# PRD and Implementation Plan: VAT Operationalization

## Context

GodarAI already has the first pieces of Virtual Account Token support: PAT/VAT key types, `/api/v1/keys/vat`, virtual-user CRUD, provider-account access resolution, and tenant-scoped key authentication. The current gap is that VAT is not yet fully operational as a production service identity. A tenant admin can create a VAT-like team key, but the key is not strongly tied to a `virtual_user_id`, VAT lifecycle operations are incomplete, and runtime quota/budget behavior is not yet enforceable at the VAT/team/key boundary.

TrueFoundry's relevant product idea is simple: developers should use one gateway token instead of provider keys; PATs are for people and development; VATs are for applications, CI/CD, and production services; virtual accounts are non-user identities with explicit model/provider access and operational lifecycle controls.

For Shahrah, we should include the parts that improve day-1 operability and adoption, and defer enterprise-only secret-store automation until the core identity, access, audit, and limits are solid.

## Inputs Reviewed

- TrueFoundry docs:
  - `.truefoundry-docs/markdown/docs__ai-gateway__authentication.md`
  - `.truefoundry-docs/markdown/docs__platform__virtual-account-management.md`
  - `.truefoundry-docs/markdown/docs__ai-gateway__gateway-access-control.md`
  - `.truefoundry-docs/markdown/docs__generating-truefoundry-api-keys.md`
- GodarAI plans:
  - `plans/readme.md`
  - `plans/tenant-organization-team-design.md`
  - `plans/phase2-tenant-admin-iam-rbac.md`
  - `plans/phase3-provider-model-catalogue.md`
  - `plans/shahrah_budgets_rate_limits.md`
- GodarAI code inspected:
  - `packages/contracts/platform/types.go`
  - `services/control-plane/internal/admin/handlers_onboarding_keys.go`
  - `services/control-plane/internal/storage/auth_repo.go`
  - `services/control-plane/internal/storage/iam_repo.go`
  - `services/control-plane/internal/auth/middleware.go`
  - `services/control-plane/internal/billing/service.go`
  - `services/control-plane/internal/server/server.go`

## Product Decisions

| Capability | Decision | GodarAI adaptation | Reason |
|---|---|---|---|
| VAT as non-human production identity | Include now | VAT must be owned by `virtual_users.virtual_user_id`, optionally scoped to team/org unit | Production apps must not depend on a human user's PAT |
| Provider-account/model access for VAT | Include now | Use existing provider-account and virtual-model access resolver; reject VAT creation if requested models exceed virtual user's scope | Matches GodarAI provider-account direction and avoids provider key exposure |
| VAT list/get/update/revoke APIs | Include now | Add canonical VAT management APIs separate from PAT APIs | Operators need lifecycle control before production use |
| Runtime auth context includes virtual user | Include now | `AuthContext` must expose `VirtualUserID` and reject inactive/deleted virtual users | Required for audit, policy, billing, and support |
| Audit and last-used visibility | Include now | Emit audit records for create/update/revoke/rotate and expose `last_used_at` | Operators need to answer "which service used what?" |
| Per-VAT rate limit and budget | Include now, MVP narrow | Implement RPM/TPM and spend checks for VAT/key/team using existing policy/rule direction; wallet check remains tenant-level backstop | VAT is not production-safe without abuse limits |
| Manual rotation | Include now | Add rotate endpoint with grace period; return new raw key once | Solves common credential hygiene without scheduler complexity |
| Auto-rotation | Defer | Store fields only if needed, but no scheduler in this phase | Useful, but not required for first operational VAT release |
| Notification on rotation | Defer | Audit events first; Slack/email later | Notification integrations widen scope |
| Secret-store sync | Defer | Document manual secret update; no Vault/Cloud secret integrations in MVP | Enterprise feature with high ops surface |
| Multi-cloud secret-manager integrations | Exclude for now | Not in GodarAI MVP | Not aligned with focused Go gateway MVP |

## Goals

- Make VAT a real service identity, not just a team-scoped key alias.
- Ensure production applications can authenticate with a VAT, receive only allowed models, and be blocked when revoked, expired, inactive, over budget, or over rate limit.
- Give tenant operators clear CRUD, revoke, rotate, audit, and last-used visibility.
- Preserve OpenAI-compatible inference behavior: clients continue to send `Authorization: Bearer <vat>`.
- Keep provider credentials centralized; VAT never exposes provider API keys.

## Non-Goals

- No full TrueFoundry parity.
- No automatic secret-manager sync in this phase.
- No Slack/email notification integration in this phase.
- No broad redesign of provider catalogue or virtual models.
- No PAT behavior expansion beyond ensuring PAT and VAT boundaries stay distinct.

## Users and Use Cases

| User | Use case | Success signal |
|---|---|---|
| Tenant admin | Create one virtual user per production app and issue VATs | App can call only authorized models; key is visible and manageable |
| Team admin | Manage service tokens for a team-owned workload | Can rotate/revoke without touching user PATs |
| Platform operator | Investigate abuse, spend spike, or failed auth | Can trace request to tenant, virtual user, VAT, team, model, and policy |
| Developer | Use VAT through OpenAI SDK | Only changes `OPENAI_API_KEY` and base URL |

## Functional Requirements

1. P0: `virtual_users` must be the owning identity for VAT.
   - VAT creation requires `virtual_user_id`.
   - `virtual_user_id` must belong to the current tenant.
   - If the virtual user has `team_id` or `org_unit_id`, the VAT scope must match or be narrower.
   - Inactive virtual users cannot create new VATs and existing VATs must fail auth.

2. P0: VAT persistence must store and return ownership fields.
   - Add `virtual_user_id` to `verification_tokens` if missing.
   - Store `key_type='vat'`, `team_id`, `organization_id`/`org_unit_id`, `virtual_user_id`, `created_by_user_id`, `token_hash`, `key_alias`, `key_name`, `expires`, `metadata`, `last_used_at`, `revoked_at`.
   - Never store the raw token.
   - Full raw token is returned only on create/rotate.

3. P0: Runtime auth must resolve VAT context correctly.
   - `GetKeyByHash` must join/verify `virtual_users`.
   - Auth context must include `VirtualUserID`, `TeamID`, `OrgUnitID`, `KeyType`, `KeyID`, `KeyAlias`, allowed models, and policy metadata.
   - Revoked, expired, blocked, inactive virtual-user, or cross-tenant keys must return normalized 401/403 errors.

4. P0: VAT model access must be explicit and scoped.
   - Requested `allowed_models` must be a subset of the virtual user's provider-account/virtual-model access.
   - If `allowed_models` is empty, VAT inherits all models available to the virtual user scope.
   - Runtime `/v1/models` must show only models available to the VAT.

5. P0: VAT lifecycle APIs must be canonical.
   - `POST /api/v1/virtual-users`
   - `GET /api/v1/virtual-users`
   - `GET /api/v1/virtual-users/{id}`
   - `PATCH /api/v1/virtual-users/{id}`
   - `DELETE /api/v1/virtual-users/{id}`
   - `POST /api/v1/virtual-users/{id}/tokens`
   - `GET /api/v1/virtual-users/{id}/tokens`
   - `PATCH /api/v1/virtual-users/{id}/tokens/{token_id}`
   - `POST /api/v1/virtual-users/{id}/tokens/{token_id}/revoke`
   - `POST /api/v1/virtual-users/{id}/tokens/{token_id}/rotate`
   - Keep `/api/v1/keys/vat` as compatibility alias during rollout.

6. P0: Manual rotation must be production-safe.
   - Rotation creates a new token with same policy fields.
   - Old token enters grace period if requested, otherwise revoked immediately.
   - Response returns the new raw token once.
   - Audit event records actor, old key id, new key id, virtual user id, and grace expiry.

7. P0: VAT rate limits must be enforced on the request path.
   - Enforce RPM by VAT key and optionally by team.
   - Enforce TPM using estimated input plus max output before forwarding; reconcile final usage after provider response where possible.
   - Return 429 with normalized OpenAI-compatible error body and rate-limit headers.

8. P0: VAT spend checks must be enforced enough for MVP.
   - Tenant wallet remains the final payment backstop.
   - VAT/key spend must be tracked per request.
   - If a key/team/org budget rule applies, block before provider call when estimated cost exceeds remaining budget.
   - Return 402 or existing GodarAI `budget_exceeded` error consistently.

9. P1: Audit and observability must identify VAT usage.
   - Request logs and traces include `key_id`, `key_type`, `virtual_user_id`, `team_id`, `org_unit_id`, and model.
   - Admin APIs expose `last_used_at`, `spend`, `revoked_at`, `expires_at`, and status.
   - Metrics must allow grouping by tenant, virtual user, key type, and model.

10. P1: Tenant-admin UI should expose minimum VAT operations.
    - Virtual users list/detail.
    - Token list per virtual user.
    - Create token, revoke token, rotate token.
    - Copy raw key only immediately after create/rotate.
    - Clear warnings for production use and rotation grace period.

## Non-Functional Requirements

- Security: token hashing only; constant-time comparison where practical; no raw token logs.
- Isolation: every query must include tenant scope.
- Latency: auth lookup and policy resolution should stay cache-friendly; target added P95 under 5 ms after warm cache.
- Reliability: if rate-limit cache fails closed for configured strict mode, otherwise fail open only with explicit operator flag.
- Auditability: all management mutations must emit audit records even if UI is not yet complete.
- Compatibility: inference remains OpenAI-compatible and provider-neutral.

## API Behavior

### Create VAT

`POST /api/v1/virtual-users/{virtual_user_id}/tokens`

Request:

```json
{
  "key_alias": "billing-api-prod",
  "allowed_models": ["openai:default:gpt-4o-mini"],
  "expires_in": "90d",
  "metadata": {
    "environment": "production",
    "owner": "payments"
  }
}
```

Response:

```json
{
  "id": "key_uuid",
  "key": "sk-...",
  "key_name": "sk-...abcd",
  "key_alias": "billing-api-prod",
  "key_type": "vat",
  "virtual_user_id": "virtual_user_uuid",
  "team_id": "team_uuid",
  "org_unit_id": "org_uuid",
  "allowed_models": ["openai:default:gpt-4o-mini"],
  "expires_at": "2026-09-14T00:00:00Z",
  "created_at": "2026-06-16T00:00:00Z"
}
```

### Rotate VAT

`POST /api/v1/virtual-users/{virtual_user_id}/tokens/{token_id}/rotate`

Request:

```json
{
  "grace_period": "24h"
}
```

Behavior:

- New key is returned once.
- Old key remains valid until `rotation_grace_expires_at` if grace period is set.
- If grace period is empty, old key is revoked immediately.

### Error Model

- `400 invalid_request`: malformed body, missing `virtual_user_id`, invalid duration.
- `401 authentication_error`: missing/invalid/revoked/expired token.
- `403 authorization_error`: caller cannot manage virtual account or requested model is out of scope.
- `404 not_found`: virtual user or token not found in tenant.
- `409 conflict`: duplicate alias where uniqueness is enforced, invalid rotation state.
- `429 rate_limit_exceeded`: VAT/team RPM or TPM exceeded.
- `402 budget_exceeded`: wallet or configured budget policy exceeded.

## Data Model

Minimum migration:

- Add `virtual_user_id UUID NULL` to `verification_tokens`.
- Add FK to `virtual_users(virtual_user_id)` with tenant-safe repository checks.
- Add `rotated_from_token_id UUID NULL`, `rotation_grace_expires_at TIMESTAMPTZ NULL`, `revoked_reason TEXT NULL` if manual rotation is implemented in the same phase.
- Add indexes:
  - `(tenant_id, virtual_user_id)`
  - `(tenant_id, key_type)`
  - `(tenant_id, revoked_at)`
  - `(token_hash)` already exists and must remain unique.

Repository behavior:

- `CreateVATToken(ctx, tenantID, virtualUserID, params)`
- `ListVATTokens(ctx, tenantID, virtualUserID, limit, offset)`
- `UpdateVATToken(ctx, tenantID, virtualUserID, tokenID, params)`
- `RevokeVATToken(ctx, tenantID, virtualUserID, tokenID, actorID, reason)`
- `RotateVATToken(ctx, tenantID, virtualUserID, tokenID, actorID, gracePeriod)`

## Architecture

- Auth layer:
  - Extend `platform.AuthContext` with `VirtualUserID`.
  - Resolve VAT by token hash and validate virtual user active state.
  - Keep PAT and VAT semantics separate.

- Admin/control layer:
  - Add canonical virtual-user token handlers.
  - Keep `/api/v1/keys/vat` as a wrapper to reduce compatibility risk.
  - Enforce `virtual_account:*` permissions.

- Persistence layer:
  - Add migration and storage methods.
  - Preserve existing `verification_tokens` table as the token source of truth.
  - Store lifecycle fields needed for revoke and manual rotation.

- Policy layer:
  - Reuse provider-account access resolver.
  - Validate allowed models during token creation and request handling.

- Billing/rate layer:
  - Implement VAT-aware Redis/cache keys:
    - `ratelimit:tenant:{tenant_id}:key:{key_id}:rpm:{window}`
    - `ratelimit:tenant:{tenant_id}:key:{key_id}:tpm:{window}`
    - optional team aggregate keys for team limits.
  - Keep tenant wallet check as required final gate.

- Observability:
  - Add virtual user dimensions to request log, spend log, trace metadata, and metrics.

## Work Breakdown

### Phase 1: Foundation and Data Model

- [ ] Add `VirtualUserID` to `platform.AuthContext` and return models.
- [ ] Add `virtual_user_id` and rotation lifecycle columns to `verification_tokens`.
- [ ] Add indexes and migration tests for existing non-VAT keys.
- [ ] Update `CreateKeyParams`/`CreatedKey` normalization so VAT requires `virtual_user_id`.
- [ ] Update storage `CreateKey` so `virtual_user_id` is persisted for VAT.
- [ ] Update `GetKeyByHash` to join/validate `virtual_users.active`.

Acceptance criteria:

- Existing PAT tests still pass.
- Existing `/api/v1/keys/vat` test fails without `virtual_user_id` unless compatibility shim creates/uses an explicit virtual user.
- VAT auth context includes `virtual_user_id`.

### Phase 2: Canonical VAT APIs

- [ ] Add storage methods for VAT token list/get/update/revoke/rotate.
- [ ] Add handlers under `/api/v1/virtual-users/{id}/tokens`.
- [ ] Keep `/api/v1/keys/vat` as compatibility alias that delegates to the canonical flow.
- [ ] Enforce `virtual_account:create/read/manage/delete` permissions.
- [ ] Add normalized errors and no raw token exposure after create/rotate.

Acceptance criteria:

- Tenant admin can create, list, revoke, and rotate VAT.
- Team member without permission cannot manage VAT.
- Revoked VAT cannot call `/v1/chat/completions`.

### Phase 3: Access Scope and Runtime Enforcement

- [ ] Ensure VAT creation validates requested models against provider-account and virtual-model scope.
- [ ] Ensure runtime model access checks include VAT virtual-user/team scope.
- [ ] Ensure `/v1/models` returns VAT-filtered models.
- [ ] Add request log/spend log fields for `virtual_user_id`.
- [ ] Add integration tests for out-of-scope model denial.

Acceptance criteria:

- VAT can call allowed model.
- VAT receives 403 for disallowed model.
- Tenant admin can see which virtual user/key generated usage.

### Phase 4: Rate Limit and Budget MVP

- [ ] Implement `CheckRateLimit` for key-level RPM and TPM using existing cache abstraction.
- [ ] Add response headers for remaining request/token capacity where available.
- [ ] Connect budget-limit rules to VAT/key/team dimensions or add a narrow key/team budget check if rule engine is not ready.
- [ ] Track VAT spend and update key/team aggregates after request finalization.
- [ ] Add tests for 429 and budget exceeded behavior.

Acceptance criteria:

- VAT exceeding RPM receives 429 before provider call.
- VAT exceeding TPM receives 429 before provider call.
- VAT exceeding budget receives `budget_exceeded` before provider call.
- Tenant wallet insufficiency still blocks all requests.

### Phase 5: Tenant-Admin UI and Operational Hardening

- [ ] Add token management to virtual-user detail UI.
- [ ] Add create, revoke, rotate flows with one-time key reveal.
- [ ] Show `last_used_at`, status, spend, model scope, expiry.
- [ ] Add audit events for create/update/revoke/rotate/auth-denied.
- [ ] Add runbook notes for rotating production VATs and updating application secrets.

Acceptance criteria:

- Operator can complete create-copy-use-rotate-revoke flow from UI.
- Audit log captures actor and target for every VAT mutation.

## Testing Strategy

- Unit tests:
  - VAT request validation.
  - key type normalization.
  - auth context derivation.
  - rate limit key construction.

- Integration tests:
  - create virtual user -> create VAT -> call allowed model.
  - inactive virtual user blocks existing VAT.
  - revoke VAT blocks existing key.
  - rotate VAT old/new key behavior with and without grace period.
  - cross-tenant virtual user/token access is rejected.
  - out-of-scope model is rejected.

- Migration tests:
  - existing PAT and legacy VAT rows remain readable.
  - new `virtual_user_id` index exists.
  - null `virtual_user_id` is allowed only for PAT/legacy rows.

- Security tests:
  - raw key never appears in list/get logs.
  - tenant A cannot manage tenant B virtual user or token.
  - team member without permission cannot create/revoke VAT.

## Rollout Plan

1. Add migration and code behind compatibility behavior.
2. Deploy with existing `/api/v1/keys/vat` still working for current tests.
3. Enable canonical `/virtual-users/{id}/tokens` APIs.
4. Migrate UI to canonical APIs.
5. Warn on VAT creation through legacy `/keys/vat`.
6. After one release, require `virtual_user_id` for VAT creation everywhere.

Rollback:

- Keep old columns nullable.
- Keep old auth lookup behavior for legacy rows.
- Disable canonical token mutation routes if needed while existing keys continue to authenticate.

## Metrics

- Number of active virtual users per tenant.
- Number of active VATs per tenant/team.
- VAT request count, token count, spend, and error rate.
- Auth failures by reason: revoked, expired, inactive virtual user, out-of-scope model.
- Rotation count and stale grace-period keys.
- Rate-limit and budget-denial count.

## Open Questions

- Should VAT creation always require an existing virtual user, or should `/api/v1/keys/vat` create an implicit virtual user during compatibility rollout?
- Should team admins be allowed to manage only VATs under their team, while tenant admins manage all?
- What is the default rotation grace period: none, 1 hour, or 24 hours?
- Should VAT budget rules be implemented through existing `budget_limit_rules` only, or should `verification_tokens` regain narrow key-budget fields?
- Should inactive virtual users immediately block all existing tokens, or should deletion/revocation be explicit and separate?

## Execution Priority for Next Step

Start with Phase 1 and Phase 2. The first implementation target is:

1. Persist `virtual_user_id` on VAT rows.
2. Resolve `virtual_user_id` in auth context.
3. Add canonical VAT token create/list/revoke endpoints under virtual users.
4. Add integration tests proving VAT ownership and revoke behavior.

Do not start auto-rotation, notification, or secret-store sync until the above is complete.

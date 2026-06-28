# Implementation Plan: Tenant Admin IAM/RBAC (Users, Org Units, Teams, Workspaces, Invitations)

## Objective

Deliver MVP-grade identity and access management for `tenant-admin` so tenant operators can manage users, org units, teams, workspaces, role assignments, invitations, and access control with clear auditability. This phase prioritizes production-safe multi-tenant control-plane behavior over feature parity.

## Inputs Reviewed

- TrueFoundry docs:
  - `.truefoundry-docs/markdown/docs__platform__user-management.md`
  - `.truefoundry-docs/markdown/docs__platform__team-management.md`
  - `.truefoundry-docs/markdown/docs__platform__user-team-account-management.md`
  - `.truefoundry-docs/markdown/docs__platform__manage-user-roles-and-permissions.md`
  - `.truefoundry-docs/markdown/docs__api-reference__users__invite-user.md`
  - `.truefoundry-docs/markdown/docs__api-reference__users__list-users.md`
  - `.truefoundry-docs/markdown/docs__api-reference__teams__create-or-update-a-team.md`
  - `.truefoundry-docs/markdown/docs__api-reference__workspaces__create-or-update-workspace.md`
- GodarAI plans:
  - `plans/readme.md`
  - `plans/phase0-tenant-provider-account-contract.md`
  - `plans/phase1-tenant-provider-account-backend-foundation.md`
  - `plans/shahrah_access_management.md`
  - `plans/shahrah_access_control.md`

## Architecture

Core split for this phase:

- `services/control-plane`: source of truth for IAM policy and membership
- `packages/database`: shared schema + migrations for IAM entities
- `apps/tenant-admin`: UI and app-route proxy for IAM management

Policy model:

- Subject: `user`, `team`, `service_account`
- Scope: `tenant`, `org_unit`, `team`, `workspace`
- Permission check path: request -> tenant resolve -> session/key resolve -> policy evaluator -> handler

## Work Breakdown

### Phase 2.1: Foundation and Contracts

- [ ] Add IAM contracts in `packages/contracts/platform`:
  - users, teams, org units, workspaces, roles, invitations, role bindings
- [ ] Add DB migration(s) for:
  - `org_units`
  - `team_memberships`
  - `workspaces`
  - `workspace_memberships`
  - `roles`
  - `permissions`
  - `role_permissions`
  - `role_bindings`
  - `invitations`
  - `audit_logs`
- [ ] Add repository interfaces and storage implementations in `services/control-plane/internal/storage`

### Phase 2.2: Core IAM APIs

- [ ] Users:
  - `GET /api/v1/users`
  - `POST /api/v1/users/invite`
  - `POST /api/v1/users/:id/activate`
  - `POST /api/v1/users/:id/deactivate`
- [ ] Teams and org units:
  - `GET|POST|PUT /api/v1/teams`
  - `GET|POST|PUT /api/v1/org-units`
- [ ] Workspaces:
  - `GET|POST|PUT /api/v1/workspaces`
- [ ] Roles:
  - `GET|POST|PUT /api/v1/roles`
  - `GET|POST|PUT /api/v1/role-bindings`

### Phase 2.3: Enforcement and Tenant Admin UI

- [ ] Add centralized policy evaluator middleware (deny-by-default).
- [ ] Enforce permissions in existing tenant-admin-sensitive endpoints.
- [ ] Add `tenant-admin` pages:
  - `/users`
  - `/org-units`
  - `/teams`
  - `/workspaces`
  - `/roles`
  - `/invitations`
- [ ] Add invitation accept flow with one-time token + TTL.

### Phase 2.4: Auditability, Security, and Hardening

- [ ] Emit audit logs for:
  - invite, accept, deactivate, membership changes, role changes, denied access
- [ ] Add rate limiting for invite and sensitive role mutation endpoints.
- [ ] Add anti-abuse checks (idempotency for invite, replay protections for accept-token).
- [ ] Add operational dashboards/alerts for access-denied spikes and invite failures.

## API Contract

MVP behavior rules:

- `service_account` key creation requires `team_id`.
- invitation token is one-time and expires.
- role updates are atomic and audited.
- resource mutation requires explicit permission for resource type + action.

Error model:

- `400`: validation errors
- `401`: unauthenticated
- `403`: unauthorized by policy
- `404`: resource not found in tenant scope
- `409`: conflict (duplicate identity, invalid transition)

## Data Model

Minimum entity set:

- `users` (tenant-scoped identities + lifecycle state)
- `org_units` (hierarchy root)
- `teams` + `team_memberships`
- `workspaces` + `workspace_memberships`
- `roles` + `permissions` + `role_permissions`
- `role_bindings` (subject-role-scope mapping)
- `invitations` (token hash, expires_at, accepted_at, invited_by)
- `audit_logs` (actor, action, target, tenant_id, payload)

Index focus:

- `(tenant_id, email)` for users
- `(tenant_id, team_alias)` for teams
- `(tenant_id, workspace_alias)` for workspaces
- `(tenant_id, subject_id, scope_type, scope_id)` for role bindings
- `(tenant_id, token_hash, expires_at)` for invitations

## Testing Strategy

- Unit tests:
  - policy evaluator, role resolution, invitation lifecycle
- Integration tests:
  - user/team/workspace CRUD in tenant scope
  - permission enforcement paths (`403` coverage)
- Migration tests:
  - forward migration on non-empty DB
  - rollback of failed migration
- Security tests:
  - invitation replay
  - horizontal privilege escalation attempts

## Rollout Plan

- Phase gate by feature flags:
  - `IAM_USERS_ENABLED`
  - `IAM_ROLES_ENABLED`
  - `IAM_WORKSPACES_ENABLED`
- Rollout sequence:
  1. Deploy DB migrations
  2. Deploy read-only APIs
  3. Enable write APIs for internal tenant
  4. Enable UI flows for pilot tenants
  5. Gradual global enablement

Rollback:

- keep old auth path active behind flag
- disable write APIs first
- preserve audit log writes during rollback

## Open Questions

- Should org-unit hierarchy support deep nesting in MVP or one-level?
- Should one user have multiple concurrent roles per same scope in MVP?
- Should invitation acceptance require OTP in addition to token?
- Should team-owned service account keys be rotatable by team manager only or tenant admin too?

## Execution Snapshot (2026-04-28)

- [x] `tenant-admin` key route now supports `service_account`, `team_id`, and `organization_id`.
- [x] `control-plane` key creation now validates:
  - `key_alias` required
  - at least one model required
  - role in `virtual_key|service_account`
  - `team_id` required for `service_account`
- [x] role derivation in storage now returns `service_account` for team-scoped keys.
- [x] integration tests added for service-account key validation and creation.
- [x] Added tenant-admin IAM contracts:
  - `TenantAdminUser`, list response, invite request/response, invitation model.
- [x] Added `control-plane` APIs:
  - `GET /api/v1/users`
  - `POST /api/v1/users/invite`
- [x] Added `storage` repository for tenant-admin users and invitations.
- [x] Added migration `007_tenant_admin_invitations.sql`.
- [x] Added tenant-admin proxy routes:
  - `GET /api/users`
  - `POST /api/users/invite`
- [x] Added integration tests for user listing and invite flow.
- [x] Added invitation acceptance API:
  - `POST /api/v1/users/invitations/accept`
- [x] Added user status APIs:
  - `POST /api/v1/users/:userID/activate`
  - `POST /api/v1/users/:userID/deactivate`
- [x] Added tenant-admin proxy routes for invitation accept and user activation toggles.
- [x] Added integration tests for invitation accept and activate/deactivate lifecycle.

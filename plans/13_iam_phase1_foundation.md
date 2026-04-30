# IAM Phase 1 Foundation

> Scope: finalize domain naming, compatibility mapping, permission matrix, ownership rules, and the stable API contract for Shahrah IAM.

## 1. Final Domain Decisions

This document is the Phase 1 decision baseline for Shahrah IAM.

### 1.1 Canonical entity names

- `Tenant`: root isolation boundary
- `OrgUnit`: optional business subdivision inside a tenant
- `Team`: workspace under a tenant or under an org unit
- `User`: human actor
- `VirtualUser`: non-human service identity
- `PAT`: personal access token, owned by a `User`
- `VAT`: virtual account token, owned by a `VirtualUser`

### 1.2 Transitional naming rules

Current implementation still uses older names in parts of the schema and codebase. Until migration is complete:

- `Organization` maps to canonical `OrgUnit`
- `VerificationToken` maps to canonical `AccessToken`
- `service_account` keys map to canonical `VAT`
- current user-bound keys map to canonical `PAT`

### 1.3 Canonical scope rules

- Every mutable IAM entity belongs to exactly one `Tenant`
- A `Team` may have `org_unit_id = null`
- A `User` may exist at tenant scope even with zero memberships
- A `VirtualUser` may exist at tenant scope or be attached to one team
- A `PAT` must reference `user_id`
- A `VAT` must reference `virtual_user_id`

## 2. Compatibility Map

This table is the required compatibility contract between current storage/code and the target domain.

| Current name | Current location | Canonical name | Migration intent |
|---|---|---|---|
| `organizations` | DB schema | `org_units` | rename or compatibility view in Phase 2 |
| `organization_memberships` | DB schema | `org_unit_memberships` | rename in Phase 2 |
| `organization_id` | contracts / DB / handlers | `org_unit_id` | keep request alias during transition |
| `verification_tokens` | DB schema | `access_tokens` or typed PAT/VAT store | split or type-discriminate in Phase 2 |
| `role=virtual_key` | control-plane key creation | `PAT` by default | replace with explicit `key_type` |
| `role=service_account` | control-plane key creation | `VAT` | replace with explicit `key_type` |
| `TenantAdminUser` | tenant-admin UI auth model | `TenantConsoleUser` | optional rename later, no Phase 1 migration |

### 2.1 Request compatibility rules

During transition, public management APIs should accept both:

- `organization_id` and `org_unit_id`
- legacy `role` in key creation requests
- canonical `key_type` in key creation requests

If both legacy and canonical fields are sent:

- canonical field wins
- server returns canonical field names in responses

## 3. Canonical Roles

### 3.1 Tenant-scoped roles

- `tenant_admin`: full IAM and budget access inside one tenant
- `tenant_user`: tenant-scoped human user without org/team admin authority

### 3.2 Org-unit-scoped roles

- `org_admin`: manages one org unit and its teams

### 3.3 Team-scoped roles

- `team_admin`: manages one team
- `team_member`: regular team member; permissions come from team policy

### 3.4 Non-canonical legacy roles

These must be treated as compatibility aliases only:

- `admin` -> `tenant_admin`
- `user` -> `tenant_user`
- `viewer` -> not part of IAM Phase 1 canonical model; if retained, model it as a separate console/read-only concern

## 4. Ownership Rules

This section is the authoritative ownership matrix for CRUD and read access.

| Entity | Create | Update | Delete | Read details/list | Read spend/budgets | Issue keys |
|---|---|---|---|---|---|---|
| `Tenant` | platform only | platform only | platform only | `tenant_admin` | `tenant_admin` | no direct keys |
| `OrgUnit` | `tenant_admin` | `tenant_admin`, owning `org_admin` | `tenant_admin` | `tenant_admin`, owning `org_admin` | `tenant_admin`, owning `org_admin` | none |
| `Team` | `tenant_admin`, owning `org_admin` | `tenant_admin`, owning `org_admin`, owning `team_admin` (within allowed fields) | `tenant_admin`, owning `org_admin` | tenant, owning org, team members where relevant | `tenant_admin`, owning `org_admin`, owning `team_admin` | none |
| `User` | `tenant_admin` | `tenant_admin`, self for profile-safe fields | `tenant_admin` | `tenant_admin`, owning org/team admins for scoped users, self | `tenant_admin`, scoped admins, self where allowed | `PAT` for self or scoped users by admins |
| `VirtualUser` | `tenant_admin`, owning `org_admin`, owning `team_admin` | same scoped admins | `tenant_admin`, owning `org_admin` | scoped admins | scoped admins | `VAT` |
| `OrgUnitMembership` | `tenant_admin`, owning `org_admin` | same | same | scoped admins and self | scoped admins | none |
| `TeamMembership` | `tenant_admin`, owning `org_admin`, owning `team_admin` | same | same | scoped admins and self | scoped admins | none |
| `PAT` | `tenant_admin`; self if allowed; scoped admins | issuer or scoped admin | issuer or scoped admin | owner + scoped admins | owner + scoped admins | n/a |
| `VAT` | `tenant_admin`, owning `org_admin`, owning `team_admin` | same | same | scoped admins | scoped admins | n/a |

### 4.1 Team-admin field restrictions

`team_admin` can update only:

- team display metadata
- allowed team member permissions
- memberships
- team-scoped virtual users
- team-scoped keys

`team_admin` cannot:

- move a team across tenants
- change tenant ownership
- expand model access beyond parent constraints
- set budgets beyond parent constraints

## 5. Permission Matrix

This is the stable route-level authorization baseline for IAM APIs.

Legend:

- `TAdmin` = `tenant_admin`
- `OAdmin` = `org_admin` for owning org unit
- `TeamAdmin` = `team_admin` for owning team
- `MemberPerm` = `team_member` with explicit permission
- `Self` = the user acts on their own record or own PAT

| Endpoint group | Action | TAdmin | OAdmin | TeamAdmin | MemberPerm | Self |
|---|---|---:|---:|---:|---:|---:|
| `/org-units` | list/details/create | yes | yes scoped | no | no | no |
| `/org-units/{id}` | update | yes | yes scoped | no | no | no |
| `/org-units/{id}` | delete | yes | no | no | no | no |
| `/org-units/{id}/members` | add/update/remove | yes | yes scoped | no | no | no |
| `/teams` | list/details/create | yes | yes scoped | no | no | no |
| `/teams/{id}` | update | yes | yes scoped | yes scoped-limited | no | no |
| `/teams/{id}` | delete | yes | yes scoped | no | no | no |
| `/teams/{id}/members` | add/update/remove | yes | yes scoped | yes scoped | no | no |
| `/teams/{id}/permissions` | view/update | yes | yes scoped | yes scoped | no | no |
| `/users` | list/details/create | yes | yes scoped | yes scoped-members only | no | self details only |
| `/users/{id}` | update | yes | scoped limited | no | no | self-safe only |
| `/users/{id}` | deactivate/delete | yes | no | no | no | no |
| `/users/{id}/memberships` | view | yes | yes scoped | yes scoped | no | yes |
| `/virtual-users` | list/details/create | yes | yes scoped | yes team-scoped | no | no |
| `/virtual-users/{id}` | update/deactivate | yes | yes scoped | yes team-scoped | no | no |
| `/virtual-users/{id}` | delete | yes | yes scoped | no | no | no |
| `/keys/pat` | create | yes | yes scoped | yes scoped for team users | yes `/key/generate` | yes for self if policy allows |
| `/keys/vat` | create | yes | yes scoped | yes team-scoped | no | no |
| `/keys` | list/details | yes | yes scoped | yes scoped | yes `/key/list` or `/key/info` | yes own keys |
| `/keys/{id}` | update | yes | yes scoped | yes scoped | yes `/key/update` | yes own PAT if policy allows |
| `/keys/{id}` | revoke/delete | yes | yes scoped | yes scoped | yes `/key/delete` | yes own PAT if policy allows |
| `/keys/{id}/rotate` | rotate | yes | yes scoped | yes scoped | yes `/key/regenerate` | yes own PAT if policy allows |
| `/audit-logs` | view | yes | yes scoped | yes team scoped | no | no |

### 5.1 Team member permission vocabulary

Canonical permission strings:

- `/key/info`
- `/key/list`
- `/key/generate`
- `/key/update`
- `/key/delete`
- `/key/regenerate`
- `/key/block`
- `/key/unblock`
- `/key/health`

Default `team_member` permissions:

- `/key/info`
- `/key/health`

## 6. Stable API Contract

This is the contract Phase 2 and UI work must build against.

### 6.1 Common rules

- All list endpoints support `limit`, `offset`, `query`
- All write endpoints are tenant-scoped through auth context, never through caller-provided tenant id
- Public response field names use canonical names
- During transition, requests may still accept `organization_id`

### 6.2 Common response envelope

List responses:

```json
{
  "data": [],
  "total": 0,
  "limit": 25,
  "offset": 0,
  "next_offset": 25
}
```

Mutation responses:

```json
{
  "data": {}
}
```

Error shape:

```json
{
  "error": {
    "code": "forbidden",
    "message": "You do not have access to this resource."
  }
}
```

### 6.3 OrgUnit contract

Canonical endpoints:

- `GET /api/v1/org-units`
- `POST /api/v1/org-units`
- `GET /api/v1/org-units/{org_unit_id}`
- `PATCH /api/v1/org-units/{org_unit_id}`
- `DELETE /api/v1/org-units/{org_unit_id}`
- `GET /api/v1/org-units/{org_unit_id}/members`
- `POST /api/v1/org-units/{org_unit_id}/members`
- `PATCH /api/v1/org-units/{org_unit_id}/members/{user_id}`
- `DELETE /api/v1/org-units/{org_unit_id}/members/{user_id}`

Create request:

```json
{
  "alias": "engineering",
  "display_name": "Engineering",
  "budget_id": "optional-budget-id",
  "metadata": {}
}
```

### 6.4 Team contract

Canonical endpoints:

- `GET /api/v1/teams`
- `POST /api/v1/teams`
- `GET /api/v1/teams/{team_id}`
- `PATCH /api/v1/teams/{team_id}`
- `DELETE /api/v1/teams/{team_id}`
- `GET /api/v1/teams/{team_id}/members`
- `POST /api/v1/teams/{team_id}/members`
- `PATCH /api/v1/teams/{team_id}/members/{user_id}`
- `DELETE /api/v1/teams/{team_id}/members/{user_id}`
- `GET /api/v1/teams/{team_id}/permissions`
- `PATCH /api/v1/teams/{team_id}/permissions`

Create request:

```json
{
  "alias": "platform-prod",
  "display_name": "Platform Production",
  "org_unit_id": "optional-org-unit-id",
  "workspace_type": "production",
  "allowed_models": ["gpt-4o-mini"],
  "budget_id": "optional-budget-id",
  "metadata": {}
}
```

### 6.5 User contract

Canonical endpoints:

- `GET /api/v1/users`
- `POST /api/v1/users`
- `GET /api/v1/users/{user_id}`
- `PATCH /api/v1/users/{user_id}`
- `DELETE /api/v1/users/{user_id}`
- `POST /api/v1/users/{user_id}/activate`
- `POST /api/v1/users/{user_id}/deactivate`
- `GET /api/v1/users/{user_id}/memberships`
- `POST /api/v1/users/invitations`
- `POST /api/v1/users/invitations/accept`
- `POST /api/v1/users/invitations/{invitation_id}/resend`
- `POST /api/v1/users/invitations/{invitation_id}/revoke`

Create request:

```json
{
  "email": "user@company.com",
  "mobile": "",
  "display_name": "User Name",
  "role": "tenant_user",
  "metadata": {}
}
```

Invitation request:

```json
{
  "email": "user@company.com",
  "role": "tenant_user",
  "accept_invite_client_url": "https://tenant-admin.example.com/invite/accept"
}
```

### 6.6 VirtualUser contract

Canonical endpoints:

- `GET /api/v1/virtual-users`
- `POST /api/v1/virtual-users`
- `GET /api/v1/virtual-users/{virtual_user_id}`
- `PATCH /api/v1/virtual-users/{virtual_user_id}`
- `DELETE /api/v1/virtual-users/{virtual_user_id}`
- `POST /api/v1/virtual-users/{virtual_user_id}/activate`
- `POST /api/v1/virtual-users/{virtual_user_id}/deactivate`

Create request:

```json
{
  "alias": "ci-cd-service",
  "display_name": "CI/CD Service",
  "team_id": "optional-team-id",
  "metadata": {}
}
```

### 6.7 Key contract

Canonical endpoints:

- `GET /api/v1/keys`
- `GET /api/v1/keys/{key_id}`
- `POST /api/v1/keys/pat`
- `POST /api/v1/keys/vat`
- `PATCH /api/v1/keys/{key_id}`
- `DELETE /api/v1/keys/{key_id}`
- `POST /api/v1/keys/{key_id}/rotate`

PAT request:

```json
{
  "key_alias": "alice-local-dev",
  "user_id": "user-id",
  "team_id": "optional-team-id",
  "allowed_models": ["gpt-4o-mini"],
  "max_budget": 50,
  "rpm_limit": 60,
  "tpm_limit": 30000,
  "expires_in": "30d",
  "metadata": {}
}
```

VAT request:

```json
{
  "key_alias": "platform-prod-service",
  "virtual_user_id": "virtual-user-id",
  "team_id": "team-id",
  "allowed_models": ["gpt-4o-mini"],
  "max_budget": 500,
  "rpm_limit": 300,
  "tpm_limit": 120000,
  "expires_in": "90d",
  "metadata": {}
}
```

Transitional compatibility request accepted in Phase 2:

```json
{
  "key_alias": "legacy-service-key",
  "role": "service_account",
  "organization_id": "legacy-org-id",
  "team_id": "team-id",
  "models": ["gpt-4o-mini"]
}
```

### 6.8 Membership contract

Org-unit membership payload:

```json
{
  "user_id": "user-id",
  "role": "org_admin",
  "budget_id": "optional-budget-id",
  "metadata": {}
}
```

Team membership payload:

```json
{
  "user_id": "user-id",
  "role": "team_member",
  "team_member_permissions": ["/key/info", "/key/health"],
  "budget_id": "optional-budget-id",
  "metadata": {}
}
```

## 7. Required Contract Changes to Current Code

The following current contracts are now officially outdated and must be migrated in Phase 2:

- `packages/contracts/platform/types.go:14`
  - `AuthContext.OrganizationID` should become canonical `OrgUnitID`
- `packages/contracts/platform/types.go:210`
  - `CreateKeyParams.Role` should be replaced by explicit `KeyType`
- `packages/contracts/platform/types.go:217`
  - `CreateKeyParams.Models` should become `AllowedModels`
- `packages/contracts/platform/types.go:222`
  - `CreateKeyParams.OrgID` should become `OrgUnitID`
- `packages/contracts/platform/tenant_admin_iam.go:1`
  - `TenantAdminUser` models console access only; it is not the canonical tenant `User` entity

## 8. Phase 1 Exit Criteria

Phase 1 is complete when:

- naming decisions are fixed
- compatibility map is documented
- route-level permission matrix is documented
- ownership rules are documented
- API contract is stable enough for backend implementation and `tenant-admin` UI design

This document satisfies that Phase 1 baseline.

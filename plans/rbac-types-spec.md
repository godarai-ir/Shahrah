# GodarAI RBAC Types Specification

> **For:** Coding Agent Implementation  
> **Module:** `shahrah/types/rbac_types.py`  
> **Dependencies:** `pydantic >= 2.0`

---

## 1. Overview

This module defines all type definitions, enums, and Pydantic models used in Shahrah's Role-Based Access Control (RBAC) system. The types cover user roles, team memberships, organizations, virtual keys, and authentication tokens.

---

## 2. Enums

### 2.1 `ShahrahUserRole`

Global proxy roles that apply across the entire platform.

| Value | Description |
|-------|-------------|
| `proxy_admin` | Full platform control. Create/manage organizations, teams, users, keys, budgets |
| `proxy_admin_viewer` | Read-only access to all platform resources. Cannot create/delete keys or modify settings |
| `internal_user` | Create/delete own keys, view own spend. Can become team/org admin if assigned |
| `internal_user_viewer` | **DEPRECATED** - View-only access. Use team/org specific roles instead |

**Implementation Notes:**
- Inherit from `str` and `Enum` for string comparison compatibility
- Use lowercase snake_case for values

### 2.2 `OrganizationTeamRole`

Organization and team specific roles (scoped to specific org/team).

| Value | Description |
|-------|-------------|
| `org_admin` | Admin over a specific organization. Can create teams and users within their org |
| `team_admin` | Admin over a specific team. Can manage team members and create keys |
| `team_member` | Regular team member with permissions defined by team admin |

### 2.3 `KeyPermission`

Permissions that can be granted to team members for key management.

| Value | Method | Description |
|-------|--------|-------------|
| `/key/info` | GET | View key information |
| `/key/health` | GET | Check key health |
| `/key/list` | GET | List all team keys |
| `/key/generate` | POST | Create new keys |
| `/key/service-account/generate` | POST | Create service account keys |
| `/key/update` | POST | Modify existing keys |
| `/key/delete` | POST | Delete keys |
| `/key/regenerate` | POST | Regenerate keys |
| `/key/block` | POST | Block keys |
| `/key/unblock` | POST | Unblock keys |

### 2.4 `VirtualKeyType`

Types of virtual keys based on their associations.

| Value | Description |
|-------|-------------|
| `user_only` | Has `user_id`, no `team_id` |
| `team_service_account` | Has `team_id`, no `user_id` |
| `user_and_team` | Has both `user_id` and `team_id` |

---

## 3. Default Constants

### `DEFAULT_TEAM_MEMBER_PERMISSIONS`

Default permissions granted to team members (read-only):

```python
DEFAULT_TEAM_MEMBER_PERMISSIONS = [
    KeyPermission.KEY_INFO,
    KeyPermission.KEY_HEALTH,
]
```

---

## 4. Budget & Rate Limit Models

### 4.1 `BudgetInfo`

Defines budget information with a time period and limit.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `budget_limit` | `float` | Yes | The maximum budget amount in USD |
| `time_period` | `str` | Yes | Duration string like `"1d"`, `"30d"`, `"1mo"` |

**Supported duration formats:** `30s`, `30m`, `30h`, `30d`, `7d`, `1mo`

### 4.2 `BudgetLimitEntry`

Entry for multiple budget windows (layered budgets).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `budget_duration` | `str` | Yes | e.g., `"24h"`, `"30d"` |
| `max_budget` | `float` | Yes | Max spend in USD for this window |

### 4.3 `RateLimitConfig`

Rate limit configuration for an entity.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tpm_limit` | `Optional[int]` | No | Tokens per minute limit |
| `rpm_limit` | `Optional[int]` | No | Requests per minute limit |
| `max_parallel_requests` | `Optional[int]` | No | Max parallel requests |

### 4.4 `ModelRateLimitConfig`

Per-model rate limit configuration.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model_rpm_limit` | `Optional[Dict[str, int]]` | No | RPM limit per model |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | No | TPM limit per model |

---

## 5. Entity Models

### 5.1 `Organization`

Organization entity - top-level business unit.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `organization_id` | `str` | Required | Unique organization identifier |
| `organization_alias` | `Optional[str]` | `None` | Human-readable name |
| `models` | `List[str]` | `[]` | Allowed models for org |
| `max_budget` | `Optional[float]` | `None` | Organization max budget |
| `tpm_limit` | `Optional[int]` | `None` | Organization TPM limit |
| `rpm_limit` | `Optional[int]` | `None` | Organization RPM limit |
| `spend` | `float` | `0.0` | Current organization spend |
| `budget_duration` | `Optional[str]` | `None` | Budget reset duration |
| `budget_reset_at` | `Optional[datetime]` | `None` | Next budget reset time |
| `budget_id` | `Optional[str]` | `None` | Associated budget ID |
| `metadata` | `Dict[str, Any]` | `{}` | Additional metadata |

### 5.2 `TeamMembership`

Represents a user's membership in a team.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `team_id` | `str` | Required | Team identifier |
| `user_id` | `str` | Required | User identifier |
| `role` | `OrganizationTeamRole` | `TEAM_MEMBER` | User role in team |
| `max_budget_in_team` | `Optional[float]` | `None` | User's max budget within team |
| `spend` | `float` | `0.0` | User's spend within team |
| `tpm_limit` | `Optional[int]` | `None` | User's TPM limit in team |
| `rpm_limit` | `Optional[int]` | `None` | User's RPM limit in team |

### 5.3 `Team`

Team entity - working group within an organization.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `team_alias` | `Optional[str]` | `None` | Human-readable team name |
| `team_id` | `str` | Required | Unique team identifier |
| `organization_id` | `Optional[str]` | `None` | Parent organization ID |
| `members` | `List[str]` | `[]` | Team member user IDs |
| `admins` | `List[str]` | `[]` | Team admin user IDs |
| `members_with_roles` | `List[Dict[str, str]]` | `[]` | Members with roles `{"role": "admin"\|"user", "user_id": "..."}` |
| `models` | `List[str]` | `[]` | Allowed models for team |
| `max_budget` | `Optional[float]` | `None` | Team max budget |
| `soft_budget` | `Optional[float]` | `None` | Soft budget threshold for alerts |
| `tpm_limit` | `Optional[int]` | `None` | Team TPM limit |
| `rpm_limit` | `Optional[int]` | `None` | Team RPM limit |
| `max_parallel_requests` | `Optional[int]` | `None` | Max parallel requests |
| `spend` | `float` | `0.0` | Current team spend |
| `budget_duration` | `Optional[str]` | `None` | Budget reset duration |
| `budget_reset_at` | `Optional[datetime]` | `None` | Next budget reset time |
| `budget_id` | `Optional[str]` | `None` | Associated budget ID |
| `metadata` | `Dict[str, Any]` | `{}` | Additional metadata |
| `model_rpm_limit` | `Optional[Dict[str, int]]` | `None` | RPM limit per model |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | `None` | TPM limit per model |
| `team_member_permissions` | `List[str]` | `DEFAULT_TEAM_MEMBER_PERMISSIONS` | Permissions for regular team members |

### 5.4 `InternalUser`

Internal user entity - individual platform user.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `user_id` | `str` | Required | Unique user identifier |
| `user_email` | `Optional[str]` | `None` | User email address |
| `user_alias` | `Optional[str]` | `None` | Human-readable alias |
| `user_role` | `ShahrahUserRole` | `INTERNAL_USER` | User's global role |
| `models` | `List[str]` | `[]` | Allowed models for user |
| `max_budget` | `Optional[float]` | `None` | User's personal max budget |
| `tpm_limit` | `Optional[int]` | `None` | User's TPM limit |
| `rpm_limit` | `Optional[int]` | `None` | User's RPM limit |
| `spend` | `float` | `0.0` | Current user spend |
| `budget_duration` | `Optional[str]` | `None` | Budget reset duration |
| `budget_reset_at` | `Optional[datetime]` | `None` | Next budget reset time |
| `budget_id` | `Optional[str]` | `None` | Associated budget ID |
| `teams` | `List[str]` | `[]` | Team IDs the user belongs to |
| `metadata` | `Dict[str, Any]` | `{}` | Additional metadata |
| `model_max_budget` | `Dict[str, BudgetInfo]` | `{}` | Model-specific budgets |

### 5.5 `VirtualKey`

Virtual API key for authentication.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `key` | `str` | Required | The actual API key value (sk-...) |
| `key_name` | `Optional[str]` | `None` | Masked key name for display (sk-...xxxx) |
| `key_alias` | `Optional[str]` | `None` | Human-readable alias |
| `models` | `List[str]` | `[]` | Allowed models for key |
| `user_id` | `Optional[str]` | `None` | Associated user ID |
| `team_id` | `Optional[str]` | `None` | Associated team ID |
| `max_budget` | `Optional[float]` | `None` | Key max budget |
| `soft_budget` | `Optional[float]` | `None` | Soft budget for alerts |
| `tpm_limit` | `Optional[int]` | `None` | Key TPM limit |
| `rpm_limit` | `Optional[int]` | `None` | Key RPM limit |
| `max_parallel_requests` | `Optional[int]` | `None` | Max parallel requests |
| `spend` | `float` | `0.0` | Current key spend |
| `budget_duration` | `Optional[str]` | `None` | Budget reset duration |
| `budget_reset_at` | `Optional[datetime]` | `None` | Next budget reset time |
| `budget_id` | `Optional[str]` | `None` | Associated budget ID |
| `budget_limits` | `Optional[List[BudgetLimitEntry]]` | `None` | Multiple budget windows |
| `model_max_budget` | `Dict[str, BudgetInfo]` | `{}` | Model-specific budgets |
| `model_rpm_limit` | `Optional[Dict[str, int]]` | `None` | RPM limit per model |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | `None` | TPM limit per model |
| `metadata` | `Dict[str, Any]` | `{}` | Additional metadata |
| `permissions` | `Dict[str, Any]` | `{}` | Key permissions |
| `config` | `Dict[str, Any]` | `{}` | Key configuration |
| `blocked` | `bool` | `False` | Whether key is blocked |
| `expires` | `Optional[datetime]` | `None` | Key expiration time |
| `duration` | `Optional[str]` | `None` | Key duration string |

---

## 6. Auth Token Model

### `AuthToken`

Represents a validated authentication token with all associated context. This is returned after successful API key validation.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `api_key` | `Optional[str]` | `None` | The validated API key |
| `token` | `Optional[str]` | `None` | The token string |
| `user_role` | `ShahrahUserRole` | `INTERNAL_USER` | User's global role |
| `user_id` | `Optional[str]` | `None` | User ID |
| `team_id` | `Optional[str]` | `None` | Team ID |
| `organization_id` | `Optional[str]` | `None` | Organization ID |
| `end_user_id` | `Optional[str]` | `None` | End user/customer ID |
| `organization_alias` | `Optional[str]` | `None` | |
| `organization_max_budget` | `Optional[float]` | `None` | |
| `organization_tpm_limit` | `Optional[int]` | `None` | |
| `organization_rpm_limit` | `Optional[int]` | `None` | |
| `team_alias` | `Optional[str]` | `None` | |
| `team_models` | `List[str]` | `[]` | |
| `team_tpm_limit` | `Optional[int]` | `None` | |
| `team_rpm_limit` | `Optional[int]` | `None` | |
| `team_max_budget` | `Optional[float]` | `None` | |
| `team_soft_budget` | `Optional[float]` | `None` | |
| `team_metadata` | `Dict[str, Any]` | `{}` | |
| `team_blocked` | `bool` | `False` | |
| `team_member_permissions` | `List[str]` | `[]` | Permissions granted to this user as team member |
| `user_email` | `Optional[str]` | `None` | |
| `user_tpm_limit` | `Optional[int]` | `None` | |
| `user_rpm_limit` | `Optional[int]` | `None` | |
| `user_max_budget` | `Optional[float]` | `None` | |
| `user_spend` | `Optional[float]` | `None` | |
| `team_member_spend` | `Optional[float]` | `None` | |
| `team_member_tpm_limit` | `Optional[int]` | `None` | |
| `team_member_rpm_limit` | `Optional[int]` | `None` | |
| `end_user_tpm_limit` | `Optional[int]` | `None` | |
| `end_user_rpm_limit` | `Optional[int]` | `None` | |
| `end_user_max_budget` | `Optional[float]` | `None` | |
| `end_user_model_max_budget` | `Optional[Dict]` | `None` | |
| `key_alias` | `Optional[str]` | `None` | |
| `models` | `List[str]` | `[]` | Models the key can access |
| `spend` | `float` | `0.0` | Current key spend |
| `max_budget` | `Optional[float]` | `None` | Key max budget |
| `soft_budget` | `Optional[float]` | `None` | Key soft budget |
| `tpm_limit` | `Optional[int]` | `None` | |
| `rpm_limit` | `Optional[int]` | `None` | |
| `model_max_budget` | `Dict` | `{}` | |
| `model_rpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `metadata` | `Dict[str, Any]` | `{}` | |
| `permissions` | `Dict[str, Any]` | `{}` | |
| `config` | `Dict[str, Any]` | `{}` | |
| `blocked` | `bool` | `False` | |
| `expires` | `Optional[datetime]` | `None` | |
| `request_route` | `Optional[str]` | `None` | The route being accessed |
| `parent_otel_span` | `Optional[Any]` | `None` | OpenTelemetry span |
| `jwt_claims` | `Optional[Dict]` | `None` | JWT claims if JWT auth was used |

---

## 7. Request/Response Models for API Endpoints

### 7.1 `CreateOrganizationRequest`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `organization_alias` | `Optional[str]` | `None` | Human-readable name |
| `models` | `List[str]` | `[]` | Allowed models |
| `max_budget` | `Optional[float]` | `None` | Organization max budget |
| `tpm_limit` | `Optional[int]` | `None` | Organization TPM limit |
| `rpm_limit` | `Optional[int]` | `None` | Organization RPM limit |
| `budget_duration` | `Optional[str]` | `None` | Budget reset duration |
| `metadata` | `Dict[str, Any]` | `{}` | Additional metadata |

### 7.2 `CreateTeamRequest`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `team_alias` | `Optional[str]` | `None` | |
| `organization_id` | `Optional[str]` | `None` | Parent organization ID |
| `members_with_roles` | `List[Dict[str, str]]` | `[]` | `{"role": "admin"\|"user", "user_id": "..."}` |
| `models` | `List[str]` | `[]` | |
| `max_budget` | `Optional[float]` | `None` | |
| `soft_budget` | `Optional[float]` | `None` | |
| `tpm_limit` | `Optional[int]` | `None` | |
| `rpm_limit` | `Optional[int]` | `None` | |
| `max_parallel_requests` | `Optional[int]` | `None` | |
| `budget_duration` | `Optional[str]` | `None` | |
| `model_rpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `team_member_permissions` | `Optional[List[str]]` | `None` | |
| `metadata` | `Dict[str, Any]` | `{}` | |

### 7.3 `CreateInternalUserRequest`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `user_id` | `Optional[str]` | `None` | |
| `user_email` | `Optional[str]` | `None` | |
| `user_alias` | `Optional[str]` | `None` | |
| `user_role` | `ShahrahUserRole` | `INTERNAL_USER` | |
| `models` | `List[str]` | `[]` | |
| `max_budget` | `Optional[float]` | `None` | |
| `tpm_limit` | `Optional[int]` | `None` | |
| `rpm_limit` | `Optional[int]` | `None` | |
| `budget_duration` | `Optional[str]` | `None` | |
| `teams` | `List[str]` | `[]` | |
| `metadata` | `Dict[str, Any]` | `{}` | |
| `model_max_budget` | `Dict[str, BudgetInfo]` | `{}` | |

### 7.4 `GenerateKeyRequest`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `user_id` | `Optional[str]` | `None` | |
| `team_id` | `Optional[str]` | `None` | |
| `key_alias` | `Optional[str]` | `None` | |
| `models` | `List[str]` | `[]` | |
| `max_budget` | `Optional[float]` | `None` | |
| `soft_budget` | `Optional[float]` | `None` | |
| `tpm_limit` | `Optional[int]` | `None` | |
| `rpm_limit` | `Optional[int]` | `None` | |
| `max_parallel_requests` | `Optional[int]` | `None` | |
| `budget_duration` | `Optional[str]` | `None` | |
| `budget_limits` | `Optional[List[BudgetLimitEntry]]` | `None` | Multiple budget windows |
| `model_max_budget` | `Dict[str, BudgetInfo]` | `{}` | |
| `model_rpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `model_tpm_limit` | `Optional[Dict[str, int]]` | `None` | |
| `metadata` | `Dict[str, Any]` | `{}` | |
| `permissions` | `Dict[str, Any]` | `{}` | |
| `config` | `Dict[str, Any]` | `{}` | |
| `duration` | `Optional[str]` | `None` | Key expiration duration |

### 7.5 `TeamMemberAddRequest`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `team_id` | `str` | Required | Team ID to add member to |
| `max_budget_in_team` | `Optional[float]` | `None` | Member's budget within team |
| `tpm_limit` | `Optional[int]` | `None` | |
| `rpm_limit` | `Optional[int]` | `None` | |
| `member` | `Dict[str, str]` | Required | `{"role": "admin"\|"user", "user_id": "..."}` |

---

## 8. File Structure

```
shahrah/
└── types/
    ├── __init__.py          # Re-export all types
    └── rbac_types.py        # Main type definitions
```

---

## 9. Module Exports (`__init__.py`)

All types should be re-exported from `__init__.py` for clean imports:

```python
fromGodarAI.types.rbac_types import (
    ShahrahUserRole,
    OrganizationTeamRole,
    KeyPermission,
    VirtualKeyType,
    DEFAULT_TEAM_MEMBER_PERMISSIONS,
    BudgetInfo,
    BudgetLimitEntry,
    RateLimitConfig,
    ModelRateLimitConfig,
    Organization,
    TeamMembership,
    Team,
    InternalUser,
    VirtualKey,
    AuthToken,
    CreateOrganizationRequest,
    CreateTeamRequest,
    CreateInternalUserRequest,
    GenerateKeyRequest,
    TeamMemberAddRequest,
)
```

---

## 10. Implementation Notes

1. **Pydantic v2 Compatibility**: Use `Field()` for all field definitions with proper `description` attributes
2. **Default Factories**: Use `default_factory=list` and `default_factory=dict` for mutable defaults
3. **Type Hints**: Use `Optional[T]` for nullable fields, `List[T]` for lists, `Dict[K, V]` for dicts
4. **Enum Values**: Inherit from both `str` and `Enum` for string comparison compatibility
5. **Imports**: Import `datetime`, `List`, `Dict`, `Optional`, `Any`, `Literal` from standard library

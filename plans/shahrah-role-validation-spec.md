# GodarAI Role Validation Specification

> **For:** Coding Agent Implementation  
> **Module:** `shahrah/auth/role_validation.py`  
> **Dependencies:** `shahrah.types.rbac_types`, `fastapi.HTTPException`, `fastapi.Request`

---

## 1. Overview

This module provides utility functions for validating user roles and checking permissions in the GodarAI RBAC system. All functions are pure (no side effects) and raise `HTTPException` on validation failures.

---

## 2. Admin Role Checks

### 2.1 `is_proxy_admin(user_role: ShahrahUserRole) -> bool`

**Purpose:** Check if the user role is proxy admin.

**Input:**
- `user_role`: A `ShahrahUserRole` enum value

**Output:**
- `True` if `user_role == ShahrahUserRole.PROXY_ADMIN`, `False` otherwise

### 2.2 `is_proxy_admin_viewer(user_role: ShahrahUserRole) -> bool`

**Purpose:** Check if the user role is proxy admin viewer.

**Output:**
- `True` if `user_role == ShahrahUserRole.PROXY_ADMIN_VIEWER`, `False` otherwise

### 2.3 `is_internal_user(user_role: ShahrahUserRole) -> bool`

**Purpose:** Check if the user role is an internal user (active or deprecated viewer).

**Output:**
- `True` if `user_role` is `INTERNAL_USER` or `INTERNAL_USER_VIEWER`, `False` otherwise

### 2.4 `is_internal_user_viewer(user_role: ShahrahUserRole) -> bool`

**Purpose:** Check if the user role is the deprecated internal user viewer.

**Output:**
- `True` if `user_role == ShahrahUserRole.INTERNAL_USER_VIEWER`, `False` otherwise

### 2.5 `is_org_admin_or_higher(auth_token: AuthToken) -> bool`

**Purpose:** Check if user is proxy admin or org admin.

**Logic:**
1. If `is_proxy_admin(auth_token.user_role)` returns `True`, return `True`
2. If `auth_token.jwt_claims` exists, check if `jwt_claims.get("user_role") == "org_admin"`, return `True` if so
3. Otherwise return `False`

### 2.6 `is_team_admin_or_higher(auth_token: AuthToken) -> bool`

**Purpose:** Check if user is proxy admin, org admin, or team admin.

**Logic:**
1. If `is_proxy_admin(auth_token.user_role)` returns `True`, return `True`
2. If `auth_token.jwt_claims` exists, check if `jwt_claims.get("user_role")` is `"org_admin"` or `"team_admin"`, return `True` if so
3. Otherwise return `False`

---

## 3. Route Classification Constants

### 3.1 `MASTER_KEY_ONLY_ROUTES`

**Type:** `Set[str]`

**Description:** Routes that require the master key (proxy admin only).

**Contents:**
```python
MASTER_KEY_ONLY_ROUTES = {
    "/organization/new",
    "/organization/delete",
    "/organization/update",
    "/team/new",
    "/team/delete",
    "/team/update",
    "/user/new",
    "/user/delete",
    "/user/update",
    "/model/add",
    "/model/delete",
    "/config/update",
}
```

### 3.2 `PUBLIC_ROUTES`

**Type:** `Set[str]`

**Description:** Routes that are public (no auth required).

**Contents:**
```python
PUBLIC_ROUTES = {
    "/health",
    "/health/services",
    "/metrics",
    "/routes",
    "/test",
}
```

### 3.3 `INTERNAL_USER_VIEWER_ALLOWED_ROUTES`

**Type:** `Set[str]`

**Description:** Routes that the deprecated internal_user_viewer role can access.

**Contents:**
```python
INTERNAL_USER_VIEWER_ALLOWED_ROUTES = {
    "/key/info",
    "/key/health",
    "/user/info",
    "/spend/calculate",
}
```

---

## 4. Route Classification Functions

### 4.1 `is_master_key_only_route(route: str) -> bool`

**Purpose:** Check if route requires master/proxy admin key.

**Output:** `route in MASTER_KEY_ONLY_ROUTES`

### 4.2 `is_public_route(route: str) -> bool`

**Purpose:** Check if route is public.

**Output:** `route in PUBLIC_ROUTES`

### 4.3 `is_internal_user_viewer_allowed_route(route: str) -> bool`

**Purpose:** Check if internal user viewer role is allowed for this route.

**Output:** `route in INTERNAL_USER_VIEWER_ALLOWED_ROUTES`

---

## 5. Authorization Check Functions

### 5.1 `check_route_authorization(route: str, auth_token: AuthToken) -> None`

**Purpose:** Main authorization entry point. Checks if the authenticated user is authorized to access the given route.

**Raises:** `HTTPException(status_code=403, detail=...)` if unauthorized.

**Logic Flow:**

```
1. PUBLIC ROUTE CHECK
   If is_public_route(route):
       return (allow)

2. PROXY ADMIN CHECK
   If is_proxy_admin(auth_token.user_role):
       return (allow - admin has access to everything)

3. DEPRECATED VIEWER ROLE CHECK
   If is_internal_user_viewer(auth_token.user_role):
       If NOT is_internal_user_viewer_allowed_route(route):
           raise HTTPException(403, detail="Role 'internal_user_viewer' is deprecated...")
       return (allow)

4. MASTER KEY ROUTE CHECK
   If is_master_key_only_route(route):
       raise HTTPException(403, detail="This route requires proxy admin privileges.")

5. ADMIN VIEWER MUTATION CHECK
   If is_proxy_admin_viewer(auth_token.user_role):
       If _is_mutation_route(route):
           raise HTTPException(403, detail="Proxy admin viewer role does not allow modifications.")
       return (allow - read-only access)

6. TEAM MEMBER PERMISSION CHECK
   If route.startswith("/key/") AND auth_token.team_id is set:
       _check_team_member_permissions(route=route, auth_token=auth_token)

7. ALLOW (passed all checks)
```

### 5.2 `_is_mutation_route(route: str) -> bool` (Private)

**Purpose:** Check if route performs mutations (create, update, delete).

**Logic:**
- Define `mutation_keywords = ["/new", "/delete", "/update", "/create", "/block", "/unblock", "/regenerate"]`
- Return `True` if any keyword is found in `route`

### 5.3 `_check_team_member_permissions(route: str, auth_token: AuthToken) -> None` (Private)

**Purpose:** Check if a team member has permission to access the key route.

**Logic Flow:**

```
1. TEAM ADMIN BYPASS
   If is_team_admin_or_higher(auth_token):
       return (allow - team admins have full access)

2. GET ALLOWED PERMISSIONS
   allowed_permissions = set(auth_token.team_member_permissions)

3. ROUTE TO PERMISSION MAPPING
   route_to_permission = {
       "/key/generate": "/key/generate",
       "/key/update": "/key/update",
       "/key/delete": "/key/delete",
       "/key/regenerate": "/key/regenerate",
       "/key/block": "/key/block",
       "/key/unblock": "/key/unblock",
       "/key/list": "/key/list",
       "/key/info": "/key/info",
       "/key/health": "/key/health",
   }

4. FIND REQUIRED PERMISSION
   For each (route_prefix, permission) in route_to_permission:
       If route.startswith(route_prefix):
           required_permission = permission
           break

5. CHECK PERMISSION
   If required_permission AND required_permission not in allowed_permissions:
       raise HTTPException(
           status_code=403,
           detail="Team member does not have permission '{required_permission}'. Contact your team admin..."
       )
```

---

## 6. Budget Validation Functions

### 6.1 `validate_budget_hierarchy(child_budget: Optional[float], parent_budget: Optional[float], child_name: str = "Entity", parent_name: str = "Parent") -> None`

**Purpose:** Validate that child budget does not exceed parent budget.

**Raises:** `HTTPException(status_code=400, detail=...)` if child > parent.

**Logic:**
```
If child_budget is not None AND parent_budget is not None:
    If child_budget > parent_budget:
        raise HTTPException(
            status_code=400,
            detail=f"{child_name} budget ({child_budget}) cannot exceed {parent_name} budget ({parent_budget})"
        )
```

### 6.2 `validate_team_budget_against_org(team_budget: Optional[float], org_budget: Optional[float]) -> None`

**Purpose:** Validate that team budget does not exceed organization budget.

**Implementation:** Wrapper around `validate_budget_hierarchy` with `child_name="Team"`, `parent_name="Organization"`.

### 6.3 `validate_key_budget_against_team(key_budget: Optional[float], team_budget: Optional[float]) -> None`

**Purpose:** Validate that key budget does not exceed team budget.

**Implementation:** Wrapper around `validate_budget_hierarchy` with `child_name="Key"`, `parent_name="Team"`.

---

## 7. Model Access Validation Functions

### 7.1 `validate_model_access(models: List[str], allowed_models: List[str], entity_name: str = "Entity") -> None`

**Purpose:** Validate that requested models are in the allowed models list.

**Special Case:** If `allowed_models` contains `"all-models"`, all models are allowed.

**Raises:** `HTTPException(status_code=400, detail=...)` if any model is not allowed.

**Logic:**
```
1. If "all-models" in allowed_models:
       return (allow all)

2. disallowed = [m for m in models if m not in allowed_models]

3. If disallowed:
       raise HTTPException(
           status_code=400,
           detail=f"{entity_name} is not allowed to access models: {disallowed}. Allowed models: {allowed_models}"
       )
```

### 7.2 `validate_team_models_against_org(team_models: List[str], org_models: List[str]) -> None`

**Purpose:** Validate that team can only access models allowed by organization.

**Implementation:** Wrapper: `validate_model_access(models=team_models, allowed_models=org_models, entity_name="Team")`

### 7.3 `validate_key_models_against_team(key_models: List[str], team_models: List[str]) -> None`

**Purpose:** Validate that key can only access models allowed by team.

**Implementation:** Wrapper: `validate_model_access(models=key_models, allowed_models=team_models, entity_name="Key")`

---

## 8. File Structure

```
shahrah/
└── auth/
    ├── __init__.py              # Re-export public functions
    └── role_validation.py       # Main role validation utilities
```

---

## 9. Module Exports (`__init__.py`)

Public functions to re-export:

```python
fromGodarAI.auth.role_validation import (
    is_proxy_admin,
    is_proxy_admin_viewer,
    is_internal_user,
    is_internal_user_viewer,
    is_org_admin_or_higher,
    is_team_admin_or_higher,
    is_master_key_only_route,
    is_public_route,
    check_route_authorization,
    validate_budget_hierarchy,
    validate_team_budget_against_org,
    validate_key_budget_against_team,
    validate_model_access,
    validate_team_models_against_org,
    validate_key_models_against_team,
)
```

---

## 10. Implementation Notes

1. **Error Handling**: All validation failures should raise `HTTPException` with appropriate status codes (400 for bad request, 403 for forbidden)
2. **Error Messages**: Messages should be descriptive and help the user understand what went wrong and how to fix it
3. **Private Functions**: Functions prefixed with `_` are internal helpers and should NOT be exported in `__init__.py`
4. **Pure Functions**: All functions should be pure (no side effects, no DB calls) - they only validate and raise exceptions
5. **Type Hints**: Use proper type hints for all function signatures
6. **Deprecation Warnings**: When the deprecated `internal_user_viewer` role is used, include a clear deprecation message

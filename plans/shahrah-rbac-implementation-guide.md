# GodarAI RBAC Implementation Guide

> **For:** Coding Agent  
> **Implementation Order:** Sequential - each module depends on the previous

---

## 1. Overview

This guide provides specifications for implementing Shahrah's Role-Based Access Control (RBAC) system. The implementation is inspired by LiteLLM's patterns but branded and adapted specifically forGodarAI.

---

## 2. Implementation Order

The Coding Agent should implement modules in this order:

| Step | Module | Specification File | Dependencies |
|------|--------|-------------------|--------------|
| 1 | Types | [`shahrah-rbac-types-spec.md`](./shahrah-rbac-types-spec.md) | None |
| 2 | Role Validation | [`shahrah-role-validation-spec.md`](./shahrah-role-validation-spec.md) | Types |
| 3 | RBAC Middleware | [`shahrah-rbac-middleware-spec.md`](./shahrah-rbac-middleware-spec.md) | Types, Role Validation |
| 4 | Tests | [`shahrah-rbac-tests-spec.md`](./shahrah-rbac-tests-spec.md) | All above |

---

## 3. File Structure

```
shahrah/
├── __init__.py                    # Main GodarAI package exports
├── types/
│   ├── __init__.py                # Re-export all types
│   └── rbac_types.py              # Enum and Pydantic model definitions
├── auth/
│   ├── __init__.py                # Re-export public auth functions
│   └── role_validation.py         # Role validation utilities
├── middleware/
│   ├── __init__.py                # Re-export middleware components
│   └── rbac_middleware.py         # FastAPI middleware and decorators
└── ... (future modules)

tests/
└── shahrah/
    ├── __init__.py
    ├── types/
    │   ├── __init__.py
    │   └── test_rbac_types.py
    ├── auth/
    │   ├── __init__.py
    │   └── test_role_validation.py
    └── middleware/
        ├── __init__.py
        └── test_rbac_middleware.py
```

---

## 4. Key Design Decisions

### 4.1 Deprecated Features (DO NOT Implement)

The following features from LiteLLM are **deprecated** and should NOT be implemented:

| Feature | Reason |
|---------|--------|
| `internal_user_viewer` role | Deprecated in favor of team/org specific roles |
| Enterprise premium features (JWT team/org admin claims) | Mark as premium/enterprise-only with clear error messages |

### 4.2 GodarAI vs LiteLLM Branding

All references to "LiteLLM" should be replaced with "Shahrah":

| LiteLLM | GodarAI |
|---------|---------|
| `LitellmUserRoles` | `ShahrahUserRole` |
| `LiteLLM_TeamTable` | `Team` |
| `LiteLLM_UserTable` | `InternalUser` |
| `LiteLLM_VerificationToken` | `VirtualKey` |
| `x-litellm-*` headers | `x-shahrah-*` headers |

### 4.3 Type Safety

- Use Pydantic v2 models with proper `Field()` definitions
- All enums inherit from both `str` and `Enum` for string comparison
- Use `Optional[T]` for nullable fields, never `None` as default for required fields

### 4.4 Error Handling

- Validation failures raise `HTTPException` with appropriate status codes
- Error messages should be descriptive and actionable
- Use 400 for bad request validation, 403 for authorization failures, 401 for authentication failures

---

## 5. Specification Files

Each specification file contains:

1. **Overview** - Purpose and dependencies
2. **Detailed Specifications** - Function signatures, logic flows, expected behavior
3. **File Structure** - Where files should be placed
4. **Module Exports** - What should be exported from `__init__.py`
5. **Implementation Notes** - Specific guidance for the Coding Agent

### 5.1 Types Specification

**File:** [`plans/shahrah-rbac-types-spec.md`](./shahrah-rbac-types-spec.md)

Defines:
- 4 Enums: `ShahrahUserRole`, `OrganizationTeamRole`, `KeyPermission`, `VirtualKeyType`
- 4 Budget/Rate Limit Models: `BudgetInfo`, `BudgetLimitEntry`, `RateLimitConfig`, `ModelRateLimitConfig`
- 5 Entity Models: `Organization`, `TeamMembership`, `Team`, `InternalUser`, `VirtualKey`
- 1 Auth Token Model: `AuthToken`
- 5 Request Models: `CreateOrganizationRequest`, `CreateTeamRequest`, `CreateInternalUserRequest`, `GenerateKeyRequest`, `TeamMemberAddRequest`

### 5.2 Role Validation Specification

**File:** [`plans/shahrah-role-validation-spec.md`](./shahrah-role-validation-spec.md)

Defines:
- 6 Admin role check functions
- 3 Route classification constants
- 3 Route classification functions
- 1 Main authorization function with detailed logic flow
- 3 Budget validation functions
- 3 Model access validation functions

### 5.3 RBAC Middleware Specification

**File:** [`plans/shahrah-rbac-middleware-spec.md`](./shahrah-rbac-middleware-spec.md)

Defines:
- `RBACMiddleware` class with `dispatch` method
- 3 Dependency injection functions for FastAPI `Depends()`
- 2 Decorators for enterprise features

### 5.4 Tests Specification

**File:** [`plans/shahrah-rbac-tests-spec.md`](./shahrah-rbac-tests-spec.md)

Defines:
- 50+ test cases across 4 test files
- Test fixtures for AuthTokens, mock requests, and auth functions
- Pytest configuration requirements

---

## 6. Integration with Existing Codebase

### 6.1 Where to Place the GodarAI Module

The `shahrah/` directory should be created at the project root level:

```
d:/source/litellm/
├── shahrah/                    # NEW - GodarAI RBAC module
│   ├── types/
│   ├── auth/
│   └── middleware/
├── litellm/                    # Existing LiteLLM code
├── tests/
│   └── shahrah/                # NEW - GodarAI tests
└── ...
```

### 6.2 How to Use in FastAPI App

```python
from fastapi import FastAPI
fromGodarAI.middleware import RBACMiddleware, require_proxy_admin
fromGodarAI.types import AuthToken, ShahrahUserRole

app = FastAPI()

# Add RBAC middleware
app.add_middleware(
    RBACMiddleware,
    get_auth_token_fn=your_auth_function,
    exclude_paths={"/docs", "/openapi.json", "/health"},
)

# Use dependency injection for route-level auth
@app.delete("/users/{user_id}")
async def delete_user(user_id: str, auth: AuthToken = Depends(require_proxy_admin)):
    # Only proxy admins can delete users
    pass
```

---

## 7. Future Enhancements (Out of Scope)

The following features are planned for future iterations:

1. **API Key Authentication** - Actual API key validation against database
2. **JWT Authentication** - JWT token parsing and claim validation
3. **OAuth2 Integration** - OAuth2 token introspection
4. **Budget Management** - Budget tracking, reset jobs, and enforcement
5. **Rate Limiting** - TPM/RPM enforcement with Redis
6. **Audit Logging** - Request/response logging for compliance
7. **Admin Dashboard UI** - React-based admin panel for managing users/teams/keys

---

## 8. References

- Original LiteLLM auth patterns: `litellm/proxy/auth/`
- Budget management: `litellm/proxy/common_utils/reset_budget_job.py`
- Timezone handling: `litellm/proxy/common_utils/timezone_utils.py`
- Duration parsing: `litellm/litellm_core_utils/duration_parser.py`

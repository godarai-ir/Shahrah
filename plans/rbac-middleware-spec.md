# GodarAI RBAC Middleware Specification

> **For:** Coding Agent Implementation  
> **Module:** `shahrah/middleware/rbac_middleware.py`  
> **Dependencies:** `shahrah.types.rbac_types`, `shahrah.auth.role_validation`, `fastapi`, `starlette`

---

## 1. Overview

This module provides FastAPI middleware for role-based access control. It intercepts HTTP requests and validates that the authenticated user has the appropriate role to access the requested route before the request reaches the endpoint handler.

---

## 2. `RBACMiddleware` Class

### 2.1 Description

A Starlette `BaseHTTPMiddleware` subclass that performs authorization checks on every request.

### 2.2 Constructor

```python
def __init__(
    self,
    app,
    get_auth_token_fn: Callable,
    exclude_paths: Optional[set] = None,
):
```

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `app` | `ASGIApp` | Yes | N/A | The ASGI application (FastAPI app) |
| `get_auth_token_fn` | `Callable` | Yes | N/A | Async function that takes a `Request` and returns an `AuthToken`. This is the authentication function that validates API keys/JWTs. |
| `exclude_paths` | `Optional[set]` | No | `set()` | Set of route paths to skip authorization for |

### 2.3 `dispatch` Method

```python
async def dispatch(
    self, request: Request, call_next: Callable
) -> Response:
```

**Purpose:** Process the request and check authorization.

**Logic Flow:**

```
1. EXTRACT ROUTE
   route = request.url.path

2. EXCLUDED PATHS CHECK
   If route in self.exclude_paths:
       return await call_next(request)

3. PUBLIC ROUTE CHECK
   If is_public_route(route):
       return await call_next(request)

4. AUTHENTICATION
   Try:
       auth_token = await self.get_auth_token_fn(request)
   Except HTTPException as e:
       return JSONResponse(status_code=e.status_code, content={"detail": e.detail})
   Except Exception as e:
       return JSONResponse(status_code=401, content={"detail": f"Authentication failed: {str(e)}"})

5. AUTHORIZATION CHECK
   Try:
       check_route_authorization(route=route, auth_token=auth_token)
   Except HTTPException as e:
       return JSONResponse(status_code=e.status_code, content={"detail": e.detail})

6. STORE AUTH TOKEN
   request.state.auth_token = auth_token

7. PROCEED WITH REQUEST
   return await call_next(request)
```

### 2.4 Usage Example

```python
from fastapi import FastAPI
fromGodarAI.middleware.rbac_middleware import RBACMiddleware
fromGodarAI.auth.api_key_auth import authenticate_api_key  # your auth function

app = FastAPI()

app.add_middleware(
    RBACMiddleware,
    get_auth_token_fn=authenticate_api_key,
    exclude_paths={"/docs", "/openapi.json", "/redoc"},
)
```

---

## 3. Dependency Injection Functions

These functions are designed to be used with FastAPI's `Depends()` for route-level authorization.

### 3.1 `require_proxy_admin(request: Request) -> None`

**Purpose:** Dependency that requires proxy admin role.

**Logic:**
```
1. auth_token = getattr(request.state, "auth_token", None)
2. If auth_token is None:
       raise HTTPException(401, "Not authenticated")
3. If auth_token.user_role != ShahrahUserRole.PROXY_ADMIN:
       raise HTTPException(403, "Proxy admin role required for this action.")
```

**Usage:**
```python
@router.post("/admin/config/update")
async def update_config(auth_token: Annotated[AuthToken, Depends(require_proxy_admin)]):
    ...
```

### 3.2 `require_proxy_admin_or_viewer(request: Request) -> None`

**Purpose:** Dependency that requires proxy admin or admin viewer role.

**Logic:**
```
1. auth_token = getattr(request.state, "auth_token", None)
2. If auth_token is None:
       raise HTTPException(401, "Not authenticated")
3. If auth_token.user_role not in (PROXY_ADMIN, PROXY_ADMIN_VIEWER):
       raise HTTPException(403, "Proxy admin or admin viewer role required.")
```

### 3.3 `require_internal_user(request: Request) -> None`

**Purpose:** Dependency that requires at least internal user role.

**Logic:**
```
1. auth_token = getattr(request.state, "auth_token", None)
2. If auth_token is None:
       raise HTTPException(401, "Not authenticated")
3. If auth_token.user_role not in (PROXY_ADMIN, PROXY_ADMIN_VIEWER, INTERNAL_USER):
       raise HTTPException(403, "Internal user role required.")
```

---

## 4. Decorators for Enterprise Features

### 4.1 `check_team_admin_required(fn: Callable) -> Callable`

**Purpose:** Decorator to check if user is a team admin. Used for enterprise/premium features.

**Logic:**
```
1. Extract auth_token from request.state
2. If auth_token is None:
       raise HTTPException(401, "Not authenticated")
3. If is_proxy_admin(auth_token.user_role):
       return await fn(*args, request=request, **kwargs)  # Admin bypass
4. If auth_token.jwt_claims and jwt_claims.get("user_role") == "team_admin":
       return await fn(*args, request=request, **kwargs)
5. raise HTTPException(403, "Team admin role required for this action. This is a premium feature.")
```

**Usage:**
```python
@router.post("/team/{team_id}/update")
@check_team_admin_required
async def update_team(request: Request, team_id: str, data: dict):
    ...
```

### 4.2 `check_org_admin_required(fn: Callable) -> Callable`

**Purpose:** Decorator to check if user is an organization admin. Used for enterprise/premium features.

**Logic:**
```
1. Extract auth_token from request.state
2. If auth_token is None:
       raise HTTPException(401, "Not authenticated")
3. If is_proxy_admin(auth_token.user_role):
       return await fn(*args, request=request, **kwargs)  # Admin bypass
4. If is_org_admin_or_higher(auth_token):
       return await fn(*args, request=request, **kwargs)
5. raise HTTPException(403, "Organization admin role required for this action. This is a premium feature.")
```

---

## 5. File Structure

```
shahrah/
└── middleware/
    ├── __init__.py              # Re-export public components
    └── rbac_middleware.py       # Main middleware implementation
```

---

## 6. Module Exports (`__init__.py`)

```python
fromGodarAI.middleware.rbac_middleware import (
    RBACMiddleware,
    require_proxy_admin,
    require_proxy_admin_or_viewer,
    require_internal_user,
    check_team_admin_required,
    check_org_admin_required,
)
```

---

## 7. Integration with FastAPI App

### 7.1 Basic Integration

```python
from fastapi import FastAPI, Request
fromGodarAI.middleware.rbac_middleware import RBACMiddleware
fromGodarAI.types.rbac_types import AuthToken
from typing import Optional

app = FastAPI()

async def get_auth_token(request: Request) -> AuthToken:
    """
    Your authentication function.
    Extract and validate API key/JWT, return AuthToken.
    """
    api_key = request.headers.get("Authorization", "").replace("Bearer ", "")
    # ... authenticate and return AuthToken
    return AuthToken(user_id="test-user", user_role=ShahrahUserRole.INTERNAL_USER)

app.add_middleware(
    RBACMiddleware,
    get_auth_token_fn=get_auth_token,
    exclude_paths={"/docs", "/openapi.json", "/redoc", "/health"},
)
```

### 7.2 Route-Level Authorization

```python
from fastapi import Depends
fromGodarAI.middleware.rbac_middleware import require_proxy_admin
fromGodarAI.types.rbac_types import AuthToken

@router.delete("/user/{user_id}")
async def delete_user(
    user_id: str,
    auth_token: Annotated[AuthToken, Depends(require_proxy_admin)],
):
    # Only proxy admins can delete users
    ...
```

---

## 8. Implementation Notes

1. **Starlette BaseHTTPMiddleware**: Use `BaseHTTPMiddleware` from Starlette, not FastAPI's built-in middleware pattern
2. **Request State**: Store the `AuthToken` in `request.state.auth_token` for downstream use
3. **Error Responses**: Return `JSONResponse` with `{"detail": "..."}` format for consistency with FastAPI error handling
4. **Async Functions**: All functions must be async since they handle async requests
5. **Type Hints**: Use `Optional[Any]` for `auth_token` variable in dependency functions
6. **Callable Import**: Import `Callable` from `typing` for type hints
7. **Decorator Pattern**: Use `functools.wraps` or async wrapper pattern for decorators

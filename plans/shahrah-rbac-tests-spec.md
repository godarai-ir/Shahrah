# Shahrah RBAC Tests Specification

> **For:** Coding Agent Implementation  
> **Test Framework:** `pytest`  
> **Test Location:** `tests/shahrah/`

---

## 1. Overview

This document specifies the test suite for Shahrah's RBAC components. Tests should cover all public functions and edge cases in the types, role validation, and middleware modules.

---

## 2. Test Structure

```
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

## 3. Types Tests (`test_rbac_types.py`)

### 3.1 Enum Tests

| Test Name | Description |
|-----------|-------------|
| `test_shahrah_user_role_values` | Verify all `ShahrahUserRole` enum values are correct strings |
| `test_organization_team_role_values` | Verify all `OrganizationTeamRole` enum values |
| `test_key_permission_values` | Verify all `KeyPermission` enum values |
| `test_virtual_key_type_values` | Verify all `VirtualKeyType` enum values |
| `test_enum_string_comparison` | Verify enum values can be compared with strings |

### 3.2 Model Instantiation Tests

| Test Name | Description |
|-----------|-------------|
| `test_budget_info_creation` | Test creating `BudgetInfo` with valid data |
| `test_budget_limit_entry_creation` | Test creating `BudgetLimitEntry` |
| `test_rate_limit_config_creation` | Test `RateLimitConfig` with optional fields |
| `test_organization_creation` | Test `Organization` with defaults and custom values |
| `test_team_creation` | Test `Team` with default member permissions |
| `test_internal_user_creation` | Test `InternalUser` with various role values |
| `test_virtual_key_creation` | Test `VirtualKey` with budget_limits and model budgets |
| `test_auth_token_creation` | Test `AuthToken` with full context |

### 3.3 Validation Tests

| Test Name | Description |
|-----------|-------------|
| `test_budget_info_required_fields` | Verify `BudgetInfo` requires both fields |
| `test_team_default_member_permissions` | Verify team defaults to `KEY_INFO` and `KEY_HEALTH` |
| `test_virtual_key_default_blocked_false` | Verify keys are not blocked by default |
| `test_virtual_key_budget_limits` | Test creating key with multiple budget windows |

---

## 4. Role Validation Tests (`test_role_validation.py`)

### 4.1 Admin Role Check Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_is_proxy_admin_true` | `is_proxy_admin(PROXY_ADMIN)` | `True` |
| `test_is_proxy_admin_false` | `is_proxy_admin(INTERNAL_USER)` | `False` |
| `test_is_proxy_admin_viewer_true` | `is_proxy_admin_viewer(PROXY_ADMIN_VIEWER)` | `True` |
| `test_is_proxy_admin_viewer_false` | `is_proxy_admin_viewer(PROXY_ADMIN)` | `False` |
| `test_is_internal_user_true` | `is_internal_user(INTERNAL_USER)` | `True` |
| `test_is_internal_user_viewer_deprecated` | `is_internal_user_viewer(INTERNAL_USER_VIEWER)` | `True` |

### 4.2 JWT Claims Role Check Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_is_org_admin_or_higher_with_proxy_admin` | AuthToken with PROXY_ADMIN role | `True` |
| `test_is_org_admin_or_higher_with_jwt_claim` | AuthToken with `jwt_claims={"user_role": "org_admin"}` | `True` |
| `test_is_org_admin_or_higher_false` | AuthToken with no claims and INTERNAL_USER role | `False` |
| `test_is_team_admin_or_higher_with_proxy_admin` | AuthToken with PROXY_ADMIN role | `True` |
| `test_is_team_admin_or_higher_with_org_admin_claim` | AuthToken with `jwt_claims={"user_role": "org_admin"}` | `True` |
| `test_is_team_admin_or_higher_with_team_admin_claim` | AuthToken with `jwt_claims={"user_role": "team_admin"}` | `True` |
| `test_is_team_admin_or_higher_false` | AuthToken with no claims and INTERNAL_USER role | `False` |

### 4.3 Route Classification Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_is_master_key_only_route_true` | Route `/organization/new` | `True` |
| `test_is_master_key_only_route_false` | Route `/key/info` | `False` |
| `test_is_public_route_true` | Route `/health` | `True` |
| `test_is_public_route_false` | Route `/user/new` | `False` |
| `test_is_internal_user_viewer_allowed_route_true` | Route `/key/info` | `True` |
| `test_is_internal_user_viewer_allowed_route_false` | Route `/user/new` | `False` |

### 4.4 Authorization Check Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_check_route_authorization_public_route` | Public route, any role | No exception |
| `test_check_route_authorization_proxy_admin` | Any route, PROXY_ADMIN | No exception |
| `test_check_route_authorization_internal_user_master_key_route` | `/user/new` with INTERNAL_USER | HTTPException 403 |
| `test_check_route_authorization_admin_viewer_mutation` | `/user/new` with PROXY_ADMIN_VIEWER | HTTPException 403 |
| `test_check_route_authorization_admin_viewer_read` | `/user/info` with PROXY_ADMIN_VIEWER | No exception |
| `test_check_route_authorization_deprecated_viewer_restricted_route` | `/user/new` with INTERNAL_USER_VIEWER | HTTPException 403 |
| `test_check_route_authorization_deprecated_viewer_allowed_route` | `/key/info` with INTERNAL_USER_VIEWER | No exception |

### 4.5 Team Member Permission Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_team_member_can_read_key_info` | Route `/key/info` with `KEY_INFO` permission | No exception |
| `test_team_member_can_check_key_health` | Route `/key/health` with `KEY_HEALTH` permission | No exception |
| `test_team_member_cannot_generate_key` | Route `/key/generate` without `KEY_GENERATE` permission | HTTPException 403 |
| `test_team_member_cannot_delete_key` | Route `/key/delete` without `KEY_DELETE` permission | HTTPException 403 |
| `test_team_admin_bypasses_permission_check` | Route `/key/generate` with team_admin JWT claim | No exception |
| `test_non_key_routes_skip_team_permission` | Route `/user/info` with team_id set | No exception |

### 4.6 Budget Validation Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_validate_budget_hierarchy_valid` | child=50, parent=100 | No exception |
| `test_validate_budget_hierarchy_equal` | child=100, parent=100 | No exception |
| `test_validate_budget_hierarchy_exceeded` | child=150, parent=100 | HTTPException 400 |
| `test_validate_budget_hierarchy_child_none` | child=None, parent=100 | No exception |
| `test_validate_budget_hierarchy_parent_none` | child=50, parent=None | No exception |
| `test_validate_team_budget_against_org_valid` | team_budget=1000, org_budget=5000 | No exception |
| `test_validate_team_budget_against_org_exceeded` | team_budget=6000, org_budget=5000 | HTTPException 400 |
| `test_validate_key_budget_against_team_valid` | key_budget=50, team_budget=100 | No exception |
| `test_validate_key_budget_against_team_exceeded` | key_budget=150, team_budget=100 | HTTPException 400 |

### 4.7 Model Access Validation Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_validate_model_access_all_allowed` | models=["gpt-4"], allowed=["gpt-4", "gpt-3.5"] | No exception |
| `test_validate_model_access_some_disallowed` | models=["gpt-4", "claude-3"], allowed=["gpt-4"] | HTTPException 400 |
| `test_validate_model_access_wildcard` | models=["any-model"], allowed=["all-models"] | No exception |
| `test_validate_team_models_against_org_valid` | team=["gpt-4"], org=["gpt-4", "gpt-3.5"] | No exception |
| `test_validate_team_models_against_org_disallowed` | team=["claude-3"], org=["gpt-4"] | HTTPException 400 |
| `test_validate_key_models_against_team_valid` | key=["gpt-3.5"], team=["gpt-4", "gpt-3.5"] | No exception |
| `test_validate_key_models_against_team_disallowed` | key=["claude-3"], team=["gpt-4"] | HTTPException 400 |

---

## 5. Middleware Tests (`test_rbac_middleware.py`)

### 5.1 RBACMiddleware Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_middleware_excluded_path` | Request to path in `exclude_paths` | Call next, no auth check |
| `test_middleware_public_route` | Request to `/health` | Call next, no auth check |
| `test_middleware_successful_auth` | Valid API key, authorized route | Call next, auth_token in request.state |
| `test_middleware_failed_auth` | Invalid API key | 401 JSONResponse |
| `test_middleware_unauthorized_route` | Valid key, master-key-only route | 403 JSONResponse |
| `test_middleware_auth_exception` | Auth function raises Exception | 401 JSONResponse with error message |
| `test_middleware_auth_http_exception` | Auth function raises HTTPException | Pass through original status code and detail |

### 5.2 Dependency Injection Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_require_proxy_admin_success` | AuthToken with PROXY_ADMIN | No exception |
| `test_require_proxy_admin_failure` | AuthToken with INTERNAL_USER | HTTPException 403 |
| `test_require_proxy_admin_no_token` | No auth_token in request.state | HTTPException 401 |
| `test_require_proxy_admin_or_viewer_admin` | AuthToken with PROXY_ADMIN | No exception |
| `test_require_proxy_admin_or_viewer_viewer` | AuthToken with PROXY_ADMIN_VIEWER | No exception |
| `test_require_proxy_admin_or_viewer_failure` | AuthToken with INTERNAL_USER | HTTPException 403 |
| `test_require_internal_user_success` | AuthToken with INTERNAL_USER | No exception |
| `test_require_internal_user_viewer_fails` | AuthToken with INTERNAL_USER_VIEWER | HTTPException 403 |
| `test_require_internal_user_no_token` | No auth_token in request.state | HTTPException 401 |

### 5.3 Decorator Tests

| Test Name | Description | Expected |
|-----------|-------------|----------|
| `test_check_team_admin_required_with_proxy_admin` | Proxy admin calls decorated function | Function executes |
| `test_check_team_admin_required_with_team_admin_jwt` | JWT claim `user_role: team_admin` | Function executes |
| `test_check_team_admin_required_without_role` | Internal user without team_admin claim | HTTPException 403 |
| `test_check_org_admin_required_with_proxy_admin` | Proxy admin calls decorated function | Function executes |
| `test_check_org_admin_required_with_org_admin_jwt` | `is_org_admin_or_higher` returns True | Function executes |
| `test_check_org_admin_required_without_role` | Internal user without org_admin | HTTPException 403 |

---

## 6. Test Fixtures

### 6.1 AuthToken Fixtures

```python
@pytest.fixture
def proxy_admin_token():
    return AuthToken(user_role=ShahrahUserRole.PROXY_ADMIN, user_id="admin-1")

@pytest.fixture
def proxy_admin_viewer_token():
    return AuthToken(user_role=ShahrahUserRole.PROXY_ADMIN_VIEWER, user_id="viewer-1")

@pytest.fixture
def internal_user_token():
    return AuthToken(user_role=ShahrahUserRole.INTERNAL_USER, user_id="user-1")

@pytest.fixture
def internal_user_viewer_token():
    return AuthToken(user_role=ShahrahUserRole.INTERNAL_USER_VIEWER, user_id="viewer-user-1")

@pytest.fixture
def team_member_token():
    return AuthToken(
        user_role=ShahrahUserRole.INTERNAL_USER,
        user_id="member-1",
        team_id="team-1",
        team_member_permissions=["/key/info", "/key/health"],
    )

@pytest.fixture
def team_admin_jwt_token():
    return AuthToken(
        user_role=ShahrahUserRole.INTERNAL_USER,
        user_id="admin-1",
        jwt_claims={"user_role": "team_admin"},
    )
```

### 6.2 Mock Request Fixture

```python
@pytest.fixture
def mock_request():
    """Create a mock FastAPI Request with configurable state."""
    class MockState:
        def __init__(self, **kwargs):
            for k, v in kwargs.items():
                setattr(self, k, v)
    
    class MockRequest:
        def __init__(self, auth_token=None):
            self.state = MockState(auth_token=auth_token)
            self.url = type("MockURL", (), {"path": "/key/info"})()
    
    return MockRequest
```

### 6.3 Mock Auth Function Fixture

```python
@pytest.fixture
def mock_auth_function():
    """Mock authentication function for middleware tests."""
    async def auth_fn(request):
        api_key = request.headers.get("Authorization", "")
        if api_key == "Bearer valid-key":
            return AuthToken(user_role=ShahrahUserRole.INTERNAL_USER, user_id="test-user")
        elif api_key == "Bearer admin-key":
            return AuthToken(user_role=ShahrahUserRole.PROXY_ADMIN, user_id="admin")
        raise HTTPException(401, "Invalid API key")
    
    return auth_fn
```

---

## 7. Test Configuration

### 7.1 Dependencies

```toml
# pyproject.toml
[project.optional-dependencies]
test = [
    "pytest>=7.0",
    "pytest-asyncio>=0.21",
    "httpx>=0.24",  # For async test client
    "pydantic>=2.0",
]
```

### 7.2 Pytest Configuration

```toml
# pyproject.toml
[tool.pytest.ini_options]
testpaths = ["tests/shahrah"]
asyncio_mode = "auto"
```

---

## 8. Implementation Notes

1. **Async Tests**: All middleware tests should be async since they test async functions
2. **HTTPException Testing**: Use `pytest.raises(HTTPException)` and check `value.status_code` and `value.detail`
3. **Mock Requests**: Use mock request objects or httpx AsyncClient for middleware testing
4. **Isolation**: Each test should be independent - don't rely on test order
5. **Coverage**: Aim for 100% code coverage on `shahrah/types/`, `shahrah/auth/`, and `shahrah/middleware/`
6. **Edge Cases**: Test None values, empty lists, and boundary conditions for budget validation

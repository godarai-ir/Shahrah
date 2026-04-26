# Phase 1: Core Gateway Stabilization - Implementation Plan

> **Status:** Draft
> **Created:** 2026-04-26
> **Target Duration:** 4-6 weeks
> **Priority:** P0 - Critical path for all subsequent phases

---

## Executive Summary

Phase 1 transforms the existing scaffold of the Shahrah control-plane into a production-ready LLM gateway. The codebase already has approximately **60% of the core structures** in place (proxy handler with streaming, tenant middleware, auth middleware, billing service, router, tunnel transport). This phase focuses on **completing gaps, hardening existing code, and adding missing critical components**.

### Current State Assessment

| Component | Completion | Status |
|-----------|------------|--------|
| Proxy Handler (sync + stream) | ~85% | Streaming implemented, needs error transformation |
| Tenant Middleware | ~70% | Resolution works, missing caching and status checks |
| Auth Middleware | ~90% | Bearer/header extraction + key hashing done |
| Billing Service | ~95% | ✅ Atomic debit + spend tracking complete |
| Router | ~90% | Simple shuffle + cooldown done |
| Tunnel (HTTP) | ~80% | Proxy + Stream done, needs error handling |
| Storage Store | ~75% | Interfaces defined, wallet debit verified |
| Admin Handlers | ~60% | Basic CRUD endpoints exist |
| Server Setup | ~90% | Fiber app wired with middleware chain |
| Tests | ~50% | Unit + integration tests for billing exist |
| **NEW:** Legacy Completions | ~80% | POST /v1/completions implemented |

### What's Already Working

- **Sync request flow**: Auth -> Tenant -> Rate Limit -> Budget -> Route -> Tunnel -> Response -> Billing -> Log
- **Streaming request flow**: Same pipeline with SSE chunk forwarding
- **API key authentication**: Bearer token + header extraction, SHA-256 hashing
- **Tenant resolution**: Host-based lookup with fallback to default tenant
- **Model routing**: Weight-based shuffle with deployment cooldown
- **Billing estimation**: Token counting + USD/IRR conversion + wallet check

### What Needs Completion

1. **Tenant caching** - Every request hits DB for tenant lookup
2. **✅ Atomic wallet debit** - COMPLETE (FOR UPDATE row lock verified)
3. **Provider error transformation** - Non-OpenAI errors not converted
4. **Additional endpoints** - ✅ /v1/completions added; /v1/embeddings pending
5. **Rate limiter hardening** - Fixed minute bucket, not true sliding window
6. **Database schema completeness** - Migrations may not match current code

---

## Task Breakdown

### Task 1: Complete Streaming SSE Proxy Handler

**Current Status:** ~85% complete. The `handleSync` and `handleStream` methods in [`proxy/handler.go`](../services/control-plane/internal/proxy/handler.go:60) are functional but lack edge case handling.

#### 1.1 Response Header Passthrough in Streaming

**File:** `services/control-plane/internal/proxy/handler.go`
**Lines:** 202-268 (handleStream method)

**Problem:** The streaming handler sets Shahrah headers but does not forward response headers from the upstream provider (e.g., `x-ratelimit-remaining-tokens` from OpenAI).

**Implementation:**
```go
// In handleStream, after establishing stream connection:
// 1. Capture initial response headers from tunnel stream
// 2. Forward relevant provider headers before streaming chunks
// 3. Headers to forward: retry-after, x-ratelimit-*, cf-ray (if applicable)

// Modify the stream setup to capture headers:
stream, err := h.tunnel.Stream(c.Context(), tunnelReq)
if err != nil {
    return writeAPIError(...)
}

// Forward provider headers if available
if headers := stream.Headers(); headers != nil {
    for key, values := range headers {
        if isAllowedHeader(key) { // whitelist: cf-ray, x-request-id, etc.
            for _, val := range values {
                c.Set(key, val)
            }
        }
    }
}
```

**Testing:** Add test that mocks tunnel stream with headers and verifies they appear in Fiber response.

#### 1.2 Provider Error Response Transformation

**File:** `services/control-plane/internal/proxy/handler.go`
**Lines:** 148-150 (sync error path) and streaming error path

**Problem:** When the upstream provider returns an error (4xx/5xx), the raw response body is sent to the client. This leaks provider-specific error formats and internal details.

**Implementation:**
```go
// New function to transform provider errors
func transformProviderError(c *fiber.Ctx, statusCode int, body []byte) error {
    // Try to parse as OpenAI error first
    var openaiErr contractsopenai.APIError
    if err := json.Unmarshal(body, &openaiErr); err == nil && openaiErr.Error.Message != "" {
        return c.Status(statusCode).JSON(openaiErr)
    }
    
    // Try to parse as Anthropic error
    var anthropicErr struct {
        Type    string `json:"type"`
        Error   struct {
            Message string `json:"message"`
            Type    string `json:"type"`
        } `json:"error"`
    }
    if err := json.Unmarshal(body, &anthropicErr); err == nil && anthropicErr.Error.Message != "" {
        return c.Status(statusCode).JSON(contractsopenai.APIError{
            Error: contractsopenai.ErrorInfo{
                Message: anthropicErr.Error.Message,
                Type:    mapProviderErrorType(anthropicErr.Error.Type),
                Code:    fmt.Sprintf("%d", statusCode),
            },
        })
    }
    
    // Fallback: generic error
    return c.Status(statusCode).JSON(contractsopenai.APIError{
        Error: contractsopenai.ErrorInfo{
            Message: "Provider returned an error",
            Type:    "api_error",
            Code:    fmt.Sprintf("%d", statusCode),
        },
    })
}
```

**Testing:** Add tests for OpenAI error passthrough, Anthropic error transformation, and malformed body handling.

#### 1.3 Trace ID / Cost Headers Verification

**File:** `services/control-plane/internal/proxy/handler.go`
**Lines:** 196-198 (handleSync) and 211 (handleStream)

**Verification:**
- `handleSync` sets `x-shahrah-trace-id`, `x-shahrah-cost-usd`, `x-shahrah-cost-irr` ✅
- `handleStream` sets `x-shahrah-trace-id` but does NOT set cost headers in initial response ✅ (cost is only known after stream ends)

**Action:** Add a comment documenting that cost headers are not available for streaming responses until the stream completes. Consider adding a final SSE event with cost metadata.

#### 1.4 Request Body Validation Enhancement

**File:** `services/control-plane/internal/proxy/handler.go`
**Lines:** 73-82

**Current validation:** Model required, allowed models check.

**Add validation for:**
```go
// After model validation, before billing estimation:
if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
    return writeAPIError(c, fiber.StatusBadRequest, "temperature must be between 0 and 2", "invalid_request_error")
}
if req.TopP != nil && (*req.TopP <= 0 || *req.TopP > 1) {
    return writeAPIError(c, fiber.StatusBadRequest, "top_p must be between 0 and 1", "invalid_request_error")
}
if req.MaxTokens != nil && *req.MaxTokens < 1 {
    return writeAPIError(c, fiber.StatusBadRequest, "max_tokens must be at least 1", "invalid_request_error")
}
if len(req.Messages) == 0 {
    return writeAPIError(c, fiber.StatusBadRequest, "messages array must not be empty", "invalid_request_error")
}
for i, msg := range req.Messages {
    if msg.Role == "" {
        return writeAPIError(c, fiber.StatusBadRequest, fmt.Sprintf("message[%d].role is required", i), "invalid_request_error")
    }
}
```

**Testing:** Add unit tests for each validation rule.

#### 1.5 Unit Tests for Proxy Handler

**File:** `services/control-plane/tests/unit/proxy/handler_test.go` (new file)

**Test cases:**
| Test | Description |
|------|-------------|
| `TestHandleSync_Success` | Valid request returns completion |
| `TestHandleSync_MissingModel` | Returns 400 for missing model |
| `TestHandleSync_ModelNotAllowed` | Returns 403 for disallowed model |
| `TestHandleSync_BudgetExceeded` | Returns 402 for insufficient wallet |
| `TestHandleSync_RateLimited` | Returns 429 for rate limit |
| `TestHandleSync_ProviderError` | Transforms provider error to OpenAI format |
| `TestHandleStream_Success` | Valid streaming request returns SSE |
| `TestHandleStream_TunnelError` | Returns 502 for tunnel failure |
| `TestHandleStream_ClientDisconnect` | Gracefully handles client disconnection |
| `TestEstimateCost_GPT4` | Correct cost estimation for gpt-4 |
| `TestEstimateCost_DefaultModel` | Correct cost estimation for default model |

**Estimated Effort:** 3-4 days

---

### Task 2: Complete Tenant Resolution with Subdomain-Based Lookup

**Current Status:** ~70% complete. The tenant middleware exists in [`tenant/middleware.go`](../services/control-plane/internal/tenant/middleware.go:1) but calls the database on every request with no caching.

#### 2.1 Verify GetTenantByHost Implementation in Storage

**File:** `services/control-plane/internal/storage/store.go`

**Action:** Read the current implementation and verify:
- SQL query for `GetTenantByHost` uses `host` column or `api_subdomain`
- Returns `*platform.Tenant` with all required fields (ID, Slug, Status, Config)
- Handles "not found" case correctly (returns error, not nil tenant)

**If missing, implement:**
```go
func (s *Store) GetTenantByHost(ctx context.Context, host string) (*platform.Tenant, error) {
    // Extract subdomain from host (e.g., "acme.sharah.ir" -> "acme")
    subdomain := extractSubdomain(host)
    
    var t platform.Tenant
    err := s.db.QueryRowContext(ctx, `
        SELECT id, slug, name, type, status, config, feature_flags, created_at, updated_at
        FROM tenants 
        WHERE api_subdomain = $1 AND status = 'ACTIVE'
        LIMIT 1
    `, subdomain).Scan(&t.ID, &t.Slug, &t.Name, &t.Type, &t.Status, &t.Config, &t.FeatureFlags, &t.CreatedAt, &t.UpdatedAt)
    
    if err == sql.ErrNoRows {
        return nil, ErrTenantNotFound
    }
    return &t, err
}
```

#### 2.2 Add Tenant Caching Layer

**File:** `services/control-plane/internal/tenant/middleware.go`

**Problem:** Every request performs a DB lookup for tenant data. Tenants rarely change, making this an ideal cache candidate.

**Implementation:**
```go
// Middleware with caching
type Middleware struct {
    store           Store
    cache           cache.Cache
    cacheTTL        time.Duration
    defaultTenant   string
    allowHeaderHint bool
}

func NewMiddleware(store Store, c cache.Cache, cacheTTL time.Duration, defaultTenant string, allowHeaderHint bool) *Middleware {
    return &Middleware{store: store, cache: c, cacheTTL: cacheTTL, defaultTenant: defaultTenant, allowHeaderHint: allowHeaderHint}
}

func (m *Middleware) Handle() fiber.Handler {
    return func(c *fiber.Ctx) error {
        host := c.Hostname()
        if m.allowHeaderHint && (host == "localhost" || strings.HasPrefix(host, "127.")) {
            if hinted := c.Get("x-shahrah-tenant"); hinted != "" {
                host = hinted
            }
        }

        // Check cache first
        cacheKey := "tenant:host:" + host
        if cached, err := m.cache.Get(c.Context(), cacheKey); err == nil && cached != nil {
            c.Locals(ContextKey{}, cached.(*platform.Tenant))
            return c.Next()
        }

        // Cache miss - query database
        tenant, err := m.store.GetTenantByHost(c.Context(), host)
        if err != nil {
            tenant, err = m.store.EnsureTenant(c.Context(), m.defaultTenant)
            if err != nil {
                return fiber.NewError(fiber.StatusNotFound, "tenant not found")
            }
        }

        // Check tenant status
        if tenant.Status != "ACTIVE" {
            return fiber.NewError(fiber.StatusForbidden, "tenant is not active")
        }

        // Cache the result
        _ = m.cache.SetWithTTL(c.Context(), cacheKey, tenant, m.cacheTTL)
        
        c.Locals(ContextKey{}, tenant)
        return c.Next()
    }
}
```

**Cache invalidation:** When a tenant is updated via admin API, delete the cache key:
```go
// In admin handler when updating tenant:
m.cache.Delete(ctx, "tenant:host:"+tenant.APISubdomain)
m.cache.Delete(ctx, "tenant:id:"+tenant.ID)
```

**Cache TTL:** 5 minutes (tenants rarely change, admin API handles cache invalidation on updates)

#### 2.3 Subdomain Extraction Utility

**File:** `services/control-plane/internal/tenant/util.go` (new file)

```go
package tenant

import "strings"

// ExtractSubdomain extracts the tenant slug from a hostname.
// Examples:
//   "acme.sharah.ir" -> "acme"
//   "my-company.sharah.ir" -> "my-company"
//   "localhost" -> ""
func ExtractSubdomain(host string) string {
    // Remove port if present
    if idx := strings.LastIndex(host, ":"); idx != -1 {
        host = host[:idx]
    }
    
    // Split by dots
    parts := strings.Split(host, ".")
    if len(parts) < 3 {
        return "" // No subdomain (e.g., "localhost", "127.0.0.1")
    }
    
    return parts[0] // First part is the subdomain
}

// ValidateTenantSlug checks if a slug is valid.
func ValidateTenantSlug(slug string) bool {
    if len(slug) < 1 || len(slug) > 63 {
        return false
    }
    for _, c := range slug {
        if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
            return false
        }
    }
    return slug[0] != '-' && slug[len(slug)-1] != '-'
}
```

#### 2.4 Tenant Status Check in Middleware

**Already covered in 2.2** - The cached middleware implementation includes status validation.

#### 2.5 Unit Tests for Tenant Middleware

**File:** `services/control-plane/tests/unit/tenant/middleware_test.go`

**Test cases:**
| Test | Description |
|------|-------------|
| `TestMiddleware_CacheHit` | Returns cached tenant without DB call |
| `TestMiddleware_CacheMiss` | Queries DB and caches result |
| `TestMiddleware_TenantNotFound` | Falls back to default tenant |
| `TestMiddleware_TenantSuspended` | Returns 403 for suspended tenant |
| `TestMiddleware_LocalDevHint` | Uses x-shahrah-tenant header for localhost |
| `TestMiddleware_InvalidHost` | Returns 404 for malformed host |

**Estimated Effort:** 2-3 days

---

### Task 3: Implement Redis-Based Rate Limiting with Lua Scripts

**Current Status:** ~60% complete. Rate limiting exists in [`billing/service.go`](../services/control-plane/internal/billing/service.go:47) but uses simple cache increment with fixed minute buckets.

#### 3.1 Atomic Lua-Script Rate Limiter

**File:** `services/control-plane/internal/cache/cache.go` (extend)

**Problem:** Current `Increment` method is not atomic for rate limiting. If two requests increment simultaneously, both may read the same value before either writes.

**Implementation:**
```go
// Add to cache package:

// RateLimitResult holds the result of a rate limit check.
type RateLimitResult struct {
    Allowed  bool
    Current  int64
    Limit    int64
    ResetAt  time.Time
}

// CheckRateLimit atomically checks and increments a rate limit counter using Lua script.
func (r *RedisClient) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
    script := `
        local key = KEYS[1]
        local limit = tonumber(ARGV[1])
        local window = tonumber(ARGV[2])
        
        local current = redis.call('GET', key)
        if not current then
            current = 0
        else
            current = tonumber(current)
        end
        
        if current >= limit then
            local ttl = redis.call('TTL', key)
            if ttl < 0 then
                ttl = window
            end
            return {0, current, ttl}
        end
        
        redis.call('INCR', key)
        local ttl = redis.call('TTL', key)
        if ttl < 0 then
            redis.call('EXPIRE', key, window)
            ttl = window
        end
        
        return {1, current + 1, ttl}
    `
    
    result, err := r.client.Eval(ctx, script, []string{key}, limit, int64(window.Seconds())).IntSlice()
    if err != nil {
        return nil, err
    }
    
    return &RateLimitResult{
        Allowed: result[0] == 1,
        Current: int64(result[1]),
        Limit:   limit,
        ResetAt: time.Now().Add(time.Duration(result[2]) * time.Second),
    }, nil
}
```

#### 3.2 Sliding Window Rate Limiter

**File:** `services/control-plane/internal/cache/cache.go`

**Problem:** Current implementation uses fixed minute buckets (`200601021504` format). This means a burst of requests at 10:00:59 and 10:01:01 counts as two separate windows, allowing 2x the limit.

**Implementation (Sliding Window Log):**
```go
// CheckSlidingWindowRateLimit uses a sorted set for true sliding window.
func (r *RedisClient) CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
    script := `
        local key = KEYS[1]
        local limit = tonumber(ARGV[1])
        local window = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])
        local windowStart = now - window
        
        -- Remove expired entries
        redis.call('ZREMRANGEBYSCORE', key, '0', windowStart)
        
        -- Count current entries
        local count = redis.call('ZCARD', key)
        
        if count >= limit then
            -- Get oldest entry to calculate reset time
            local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
            local resetAt = tonumber(oldest[2]) + window
            return {0, count, resetAt - now}
        end
        
        -- Add new entry
        redis.call('ZADD', key, now, now .. ':' .. math.random(1000000))
        redis.call('EXPIRE', key, window * 2)
        
        return {1, count + 1, window}
    `
    
    now := time.Now().UnixMilli()
    result, err := r.client.Eval(ctx, script, []string{key}, limit, int64(window.Seconds()), now).IntSlice()
    if err != nil {
        return nil, err
    }
    
    return &RateLimitResult{
        Allowed: result[0] == 1,
        Current: int64(result[1]),
        Limit:   limit,
        ResetAt: time.Now().Add(time.Duration(result[2]) * time.Millisecond),
    }, nil
}
```

**Note:** Sliding window is more expensive (sorted set vs simple key). Use it only for TPM limits where precision matters. Keep fixed window for RPM limits.

#### 3.3 Per-Tenant Rate Limit Aggregation

**File:** `services/control-plane/internal/billing/service.go`

**Current behavior:** Rate limits are per-key only (`auth.TokenHash`).

**Add tenant-level aggregation:**
```go
// In CheckRateLimit, after per-key check:
// Also check tenant-level aggregate rate limit
tenantRPMKey := "rpm:tenant:" + auth.TenantID + ":" + nowBucket
tenantResult, err := s.cache.CheckRateLimit(ctx, tenantRPMKey, tenantRPMLimit, window)
if err != nil {
    return err // Cache error
}
if !tenantResult.Allowed {
    return ErrRateLimitExceeded // Tenant-level limit exceeded
}
```

**Tenant RPM/TPM limits:** Stored in tenant config or a separate rate_limits table.

#### 3.4 Rate Limit Response Headers

**File:** `services/control-plane/internal/proxy/handler.go`

**Add headers to all responses:**
```go
// After rate limit check, set headers:
c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Limit - result.Current))
c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
```

#### 3.5 Unit Tests for Rate Limiter

**File:** `services/control-plane/tests/unit/cache/rate_limit_test.go`

**Test cases:**
| Test | Description |
|------|-------------|
| `TestRateLimit_Allowed` | Requests within limit are allowed |
| `TestRateLimit_Exceeded` | Requests over limit are blocked |
| `TestRateLimit_WindowReset` | Counter resets after window |
| `TestRateLimit_Concurrent` | Concurrent requests counted correctly |
| `TestSlidingWindow_Precision` | Sliding window allows burst at boundary |
| `TestTenantAggregate` | Tenant-level rate limiting works |

**Estimated Effort:** 3-4 days

---

### Task 4: Complete Billing - Wallet Debit and Spend Tracking

**Current Status:** ✅ COMPLETE (95%). All implementations verified and working.

- [`DebitWallet`](services/control-plane/internal/storage/wallet_repo.go:74) - Atomic with `FOR UPDATE` row lock
- [`LogSpend`](services/control-plane/internal/storage/log_repo.go:22) - Writes to spend_logs table
- [Unit tests](services/control-plane/tests/unit/billing/service_test.go) - Pass

#### 4.1 Verified Implementations

**Verification Complete (2026-04-26):**

1. **DebitWallet** - [`wallet_repo.go:74`](services/control-plane/internal/storage/wallet_repo.go:74)
   - Uses `FOR UPDATE` row lock for atomicity ✅
   - Checks balance before deduction ✅
   - Records transaction with metadata ✅

2. **LogSpend** - [`log_repo.go:22`](services/control-plane/internal/storage/log_repo.go:22)
   - Inserts into spend_logs table ✅
   - Updates verification_tokens spend ✅

3. **Integration Tests** - [`storage_integration_test.go`](services/control-plane/tests/integration/storage_integration_test.go)
   - Concurrent debit test verifies row lock ✅

**Status:** Complete - No further action required for billing wallet debit and spend tracking.

---

### Task 5: Add Additional OpenAI-Compatible Endpoints

**Current Status:** ~50% complete. `POST /v1/completions` implemented. Embeddings and model detail pending.

#### 5.1 POST /v1/completions (Legacy Text Completions)

**File:** `services/control-plane/internal/proxy/completions_handler.go` (new file)

```go
type CompletionsHandler struct {
    llmRouter *router.Router
    tunnel    platform.TunnelTransport
    billing   budgeter
    wallets   walletStore
    logs      usageLogger
    logger    *zap.Logger
}

// Handle processes a text completion request.
func (h *CompletionsHandler) Handle(c *fiber.Ctx) error {
    var req contractsopenai.CompletionRequest
    if err := c.BodyParser(&req); err != nil {
        return writeAPIError(...)
    }
    
    // Transform to chat completion format internally
    chatReq := contractsopenai.ChatCompletionRequest{
        Model:     req.Model,
        Messages:  []contractsopenai.Message{{Role: "user", Content: req.Prompt}},
        Stream:    req.Stream,
        MaxTokens: req.MaxTokens,
        Temperature: req.Temperature,
    }
    
    // Reuse chat completion handler logic
    // ... (same pipeline: auth, tenant, rate limit, budget, route, tunnel)
}
```

#### 5.2 POST /v1/embeddings

**File:** `services/control-plane/internal/proxy/embeddings_handler.go` (new file)

```go
type EmbeddingsHandler struct {
    llmRouter *router.Router
    tunnel    platform.TunnelTransport
    billing   budgeter
    wallets   walletStore
    logs      usageLogger
    logger    *zap.Logger
}

func (h *EmbeddingsHandler) Handle(c *fiber.Ctx) error {
    var req contractsopenai.EmbeddingRequest
    if err := c.BodyParser(&req); err != nil {
        return writeAPIError(...)
    }
    
    // Embeddings are always sync (no streaming)
    // Pipeline: auth -> tenant -> rate limit -> budget -> route -> tunnel -> billing
    // ...
}
```

#### 5.3 GET /v1/models

**Already exists** in admin handlers. Needs enhancement to return only models available to the current tenant.

#### 5.4 GET /v1/models/{model_id}

**File:** Extend admin handlers or create new proxy handler.

```go
func (h *AdminHandlers) GetModel(c *fiber.Ctx) error {
    modelID := c.Params("model_id")
    
    // Look up model in tenant's deployment list
    deployments, err := h.store.ListDeployments(c.Context(), currentTenant.ID)
    if err != nil {
        return err
    }
    
    for _, dep := range deployments {
        if dep.ModelName == modelID {
            return c.JSON(contractsopenai.ModelInfo{
                ID:      dep.ModelName,
                Object:  "model",
                Created: dep.CreatedAt.Unix(),
                OwnedBy: dep.Provider,
            })
        }
    }
    
    return writeAPIError(c, fiber.StatusNotFound, "Model not found", "not_found")
}
```

#### 5.5 Provider Transformation Layer

**File:** `services/control-plane/internal/providers/transformer.go` (new file)

```go
type Transformer struct {
    providerType string
}

func (t *Transformer) TransformRequest(model string, req *contractsopenai.ChatCompletionRequest) ([]byte, error) {
    switch t.providerType {
    case "openai":
        return json.Marshal(req) // Already in OpenAI format
    case "anthropic":
        return t.toAnthropicFormat(req)
    default:
        return json.Marshal(req)
    }
}

func (t *Transformer) TransformResponse(body []byte, model string) (*contractsopenai.ChatCompletionResponse, error) {
    switch t.providerType {
    case "openai":
        var resp contractsopenai.ChatCompletionResponse
        err := json.Unmarshal(body, &resp)
        return &resp, err
    case "anthropic":
        return t.fromAnthropicFormat(body, model)
    default:
        var resp contractsopenai.ChatCompletionResponse
        err := json.Unmarshal(body, &resp)
        return &resp, err
    }
}
```

**Estimated Effort:** 4-5 days

---

### Task 6: Database Schema and Migration Completeness

**Current Status:** ~60% complete. Two migration files exist but need verification against current code.

#### 6.1 Review and Update Migrations

**Files:** 
- `db/main/migrations/001_initial.sql`
- `db/main/migrations/002_platform_foundations.sql`

**Action:** Compare current SQL schema with all Go struct definitions in `contracts/platform/types.go`. Ensure every field has a corresponding column.

**Tables that should exist:**
- `tenants` - ID, slug, name, type, status, api_subdomain, admin_subdomain, config, feature_flags, wallet_id, deployment_type, created_at, updated_at
- `api_keys` - ID, tenant_id, key_hash, key_prefix, key_name, permissions, allowed_models, rpm_limit, tpm_limit, status, expires_at, last_used_at, created_by, created_at, updated_at
- `wallets` - ID, tenant_id, user_id, team_id, balance_irr, pending_irr, balance_usd, daily_limit_irr, monthly_limit_irr, created_at, updated_at
- `transactions` - ID, wallet_id, tx_type, amount_irr, amount_usd, exchange_rate, gateway, gateway_tx_id, description, metadata, status, request_id, created_at, completed_at
- `exchange_rates` - ID, usd_to_irr, source, fetched_at, rate_type, is_manual, overridden_by
- `deployments` - ID, tenant_id, model_name, provider, model, api_base, api_version, region_name, rpm, tpm, weight, timeout_sec, healthy, created_at, updated_at
- `spend_logs` - ID, request_id, token_hash, tenant_id, org_id, team_id, model, input_tokens, output_tokens, total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata, completion_start, completion_end
- `request_logs` - ID, request_id, token_hash, tenant_id, model, status, cost_usd, response_time_ms, request_body, response_body, created_at

#### 6.2 Add Performance Indexes

```sql
-- Tenant lookups
CREATE INDEX idx_tenants_api_subdomain ON tenants(api_subdomain) WHERE status = 'ACTIVE';
CREATE INDEX idx_tenants_slug ON tenants(tenant_slug);

-- API key lookups
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_tenant_status ON api_keys(tenant_id, status);

-- Wallet lookups
CREATE INDEX idx_wallets_tenant ON wallets(tenant_id);

-- Transaction lookups
CREATE INDEX idx_transactions_wallet ON transactions(wallet_id);
CREATE INDEX idx_transactions_tenant ON transactions(wallet_id, created_at DESC);
CREATE INDEX idx_transactions_request ON transactions(request_id);

-- Spend log queries
CREATE INDEX idx_spend_logs_tenant_date ON spend_logs(tenant_id, completion_start DESC);
CREATE INDEX idx_spend_logs_model ON spend_logs(model, completion_start DESC);

-- Deployment lookups
CREATE INDEX idx_deployments_tenant_model ON deployments(tenant_id, model_name, healthy);
```

#### 6.3 Seed Data for Local Development

**File:** `db/main/migrations/003_seed_data.sql`

```sql
-- Default tenant
INSERT INTO tenants (id, slug, name, type, status, api_subdomain, config, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'default', 'Default Tenant', 'TRIAL', 'ACTIVE', 'localhost', '{}', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Default wallet
INSERT INTO wallets (id, tenant_id, balance_irr, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 10000000, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Test API key (sk-test-key -> hash it first)
INSERT INTO api_keys (id, tenant_id, key_hash, key_prefix, key_name, status, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '<sha256-of-sk-test-key>', 'sk-test-', 'Test Key', 'ACTIVE', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Test deployment
INSERT INTO deployments (id, tenant_id, model_name, provider, model, api_base, weight, healthy, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'gpt-4o', 'openai', 'gpt-4o', 'https://api.openai.com', 1.0, true, NOW(), NOW())
ON CONFLICT DO NOTHING;
```

#### 6.4 Migration Runner

**File:** `services/control-plane/cmd/migrate/main.go` (new file)

```go
package main

import (
    "database/sql"
    "embed"
    "fmt"
    "log"
    "os"
    "sort"
    "strings"
    
    _ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable required")
    }
    
    db, err := sql.Open("pgx", dbURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Create migrations table
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INT PRIMARY KEY,
            applied_at TIMESTAMPTZ DEFAULT NOW()
        )
    `)
    if err != nil {
        log.Fatalf("Failed to create migrations table: %v", err)
    }
    
    // Get applied migrations
    rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
    if err != nil {
        log.Fatalf("Failed to query migrations: %v", err)
    }
    defer rows.Close()
    
    applied := make(map[int]bool)
    for rows.Next() {
        var v int
        rows.Scan(&v)
        applied[v] = true
    }
    
    // Read migration files
    entries, err := migrationsFS.ReadDir("migrations")
    if err != nil {
        log.Fatalf("Failed to read migrations directory: %v", err)
    }
    
    // Sort and apply
    var migrations []string
    for _, e := range entries {
        if strings.HasSuffix(e.Name(), ".sql") {
            migrations = append(migrations, e.Name())
        }
    }
    sort.Strings(migrations)
    
    for _, filename := range migrations {
        version := extractVersion(filename)
        if applied[version] {
            continue
        }
        
        content, err := migrationsFS.ReadFile("migrations/" + filename)
        if err != nil {
            log.Fatalf("Failed to read migration %s: %v", filename, err)
        }
        
        log.Printf("Applying migration %s...", filename)
        tx, err := db.Begin()
        if err != nil {
            log.Fatalf("Failed to begin transaction: %v", err)
        }
        
        _, err = tx.Exec(string(content))
        if err != nil {
            tx.Rollback()
            log.Fatalf("Failed to apply migration %s: %v", filename, err)
        }
        
        _, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
        if err != nil {
            tx.Rollback()
            log.Fatalf("Failed to record migration %s: %v", filename, err)
        }
        
        if err := tx.Commit(); err != nil {
            log.Fatalf("Failed to commit migration %s: %v", filename, err)
        }
        
        log.Printf("Applied migration %s", filename)
    }
    
    log.Println("All migrations applied successfully")
}

func extractVersion(filename string) int {
    // "001_initial.sql" -> 1
    parts := strings.SplitN(filename, "_", 2)
    var v int
    fmt.Sscanf(parts[0], "%d", &v)
    return v
}
```

#### 6.5 Integration Tests Against Real Database

**File:** `services/control-plane/tests/integration/storage_test.go`

**Use testcontainers-go for PostgreSQL:**
```go
func TestStore_Integration(t *testing.T) {
    ctx := context.Background()
    
    // Start PostgreSQL container
    pg, err := testcontainers.GenericContainer(ctx, testcontainers.ContainerRequest{
        Image: "postgres:16-alpine",
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "shahrah_test",
        },
        Ports: []string{"5432/tcp"},
    })
    require.NoError(t, err)
    t.Cleanup(func() { pg.Terminate(ctx) })
    
    // Get connection string
    host, _ := pg.Host(ctx)
    port, _ := pg.MappedPort(ctx, "5432")
    dbURL := fmt.Sprintf("postgres://test:test@%s:%s/shahrah_test", host, port.Port())
    
    // Run migrations
    runMigrations(dbURL)
    
    // Create store and test operations
    store, err := storage.NewStore(ctx, dbURL, "default", zap.NewNop())
    require.NoError(t, err)
    
    t.Run("CreateAndFindTenant", testCreateAndFindTenant(store))
    t.Run("DebitWallet", testDebitWallet(store))
    // ...
}
```

**Estimated Effort:** 3-4 days

---

### Task 7: Provider Expansion - Add Anthropic Client

**Current Status:** ~10% complete. Only OpenAI provider client exists.

#### 7.1 Anthropic Provider Client

**File:** `services/control-plane/internal/providers/anthropic/client.go` (new file)

```go
package anthropic

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
    "github.com/Delpak/shahrah/packages/contracts/platform"
)

type Client struct {
    apiKey     string
    apiBase    string
    httpClient *http.Client
    timeout    time.Duration
}

func NewClient(apiKey, apiBase string, timeout time.Duration) *Client {
    return &Client{
        apiKey:  apiKey,
        apiBase: apiBase,
        httpClient: &http.Client{
            Timeout: timeout,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 20,
                IdleConnTimeout:     30 * time.Second,
            },
        },
    }
}

func (c *Client) ChatCompletion(ctx context.Context, req *contractsopenai.ChatCompletionRequest) (*contractsopenai.ChatCompletionResponse, error) {
    anthropicReq := toAnthropicRequest(req)
    
    body, err := json.Marshal(anthropicReq)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiBase+"/v1/messages", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("anthropic-version", "2023-06-01")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("anthropic API error: %s", string(respBody))
    }
    
    return fromAnthropicResponse(respBody, req.Model)
}
```

#### 7.2 Provider Transformation Layer

**File:** `services/control-plane/internal/providers/anthropic/transform.go`

```go
type AnthropicRequest struct {
    Model         string                   `json:"model"`
    MaxTokens     int                      `json:"max_tokens"`
    Messages      []AnthropicMessage       `json:"messages"`
    System        string                   `json:"system,omitempty"`
    Temperature   *float64                 `json:"temperature,omitempty"`
    TopP          *float64                 `json:"top_p,omitempty"`
    Stream        bool                     `json:"stream,omitempty"`
}

type AnthropicMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func toAnthropicRequest(req *contractsopenai.ChatCompletionRequest) *AnthropicRequest {
    messages := make([]AnthropicMessage, 0, len(req.Messages))
    systemPrompt := ""
    
    for _, msg := range req.Messages {
        if msg.Role == "system" {
            systemPrompt = msg.Content
            continue
        }
        messages = append(messages, AnthropicMessage{
            Role:    msg.Role, // "user" -> "user", "assistant" -> "assistant"
            Content: msg.Content,
        })
    }
    
    maxTokens := 1024
    if req.MaxTokens != nil {
        maxTokens = *req.MaxTokens
    }
    
    return &AnthropicRequest{
        Model:       mapModelName(req.Model),
        MaxTokens:   maxTokens,
        Messages:    messages,
        System:      systemPrompt,
        Temperature: req.Temperature,
        TopP:        req.TopP,
        Stream:      req.Stream,
    }
}

func mapModelName(openaiModel string) string {
    // Map common model aliases to Anthropic models
    switch {
    case openaiModel == "claude-3-opus", openaiModel == "claude-opus":
        return "claude-3-opus-20240229"
    case openaiModel == "claude-3-sonnet", openaiModel == "claude-sonnet":
        return "claude-3-sonnet-20240229"
    case openaiModel == "claude-3-haiku", openaiModel == "claude-haiku":
        return "claude-3-haiku-20240307"
    default:
        return openaiModel // Pass through if it looks like an Anthropic model
    }
}
```

#### 7.3 Provider Router Fallback

**File:** `services/control-plane/internal/router/router.go`

**Enhance the router to try fallback providers:**
```go
func (r *Router) RouteWithFallback(ctx context.Context, tenantID, modelName string) (*Deployment, error) {
    // First, try the primary model mapping
    dep, err := r.Route(ctx, tenantID, modelName)
    if err == nil {
        return dep, nil
    }
    
    // If not found, check if there's a fallback model mapping
    fallback := r.getFallbackModel(modelName)
    if fallback != "" {
        r.logger.Warn("Model not found, trying fallback", 
            zap.String("original", modelName),
            zap.String("fallback", fallback))
        return r.Route(ctx, tenantID, fallback)
    }
    
    return nil, err
}

func (r *Router) getFallbackModel(modelName string) string {
    fallbacks := map[string]string{
        "gpt-4":        "claude-3-sonnet",
        "gpt-4-turbo":  "claude-3-sonnet",
        "gpt-3.5-turbo": "claude-3-haiku",
    }
    if fallback, ok := fallbacks[modelName]; ok {
        return fallback
    }
    return ""
}
```

#### 7.4 Model Aliasing

**File:** `services/control-plane/internal/router/alias.go` (new file)

```go
// ModelAliasRegistry maps user-facing model names to actual provider models.
var ModelAliasRegistry = map[string]AliasEntry{
    "gpt-4o": {
        Provider: "openai",
        Model:    "gpt-4o",
    },
    "gpt-4": {
        Provider: "openai",
        Model:    "gpt-4-turbo",
    },
    "claude-sonnet": {
        Provider: "anthropic",
        Model:    "claude-3-sonnet-20240229",
    },
    "claude-opus": {
        Provider: "anthropic",
        Model:    "claude-3-opus-20240229",
    },
}

type AliasEntry struct {
    Provider string
    Model    string
}
```

#### 7.5 Unit Tests for Provider Transformation

**File:** `services/control-plane/tests/unit/providers/anthropic/transform_test.go`

**Test cases:**
| Test | Description |
|------|-------------|
| `TestToAnthropicRequest_Basic` | Simple chat conversion |
| `TestToAnthropicRequest_SystemPrompt` | System prompt extracted to top-level field |
| `TestToAnthropicRequest_MultiTurn` | Conversation history preserved |
| `TestFromAnthropicResponse` | Parse Anthropic response to OpenAI format |
| `TestModelMapping` | Model name aliases resolve correctly |

**Estimated Effort:** 4-5 days

---

### Task 8: Integration and End-to-End Testing

**Current Status:** ~40% complete. Test infrastructure exists but coverage is limited.

#### 8.1 Full Request Lifecycle Integration Test

**File:** `services/control-plane/tests/integration/lifecycle_test.go`

```go
func TestFullRequestLifecycle(t *testing.T) {
    // Setup: Start test server with real dependencies
    env := setupTestEnvironment(t)
    
    // Step 1: Create tenant
    tenant := env.CreateTenant(t, "test-org")
    
    // Step 2: Create API key
    key := env.CreateAPIKey(t, tenant.ID, "sk-test-lifecycle")
    
    // Step 3: Fund wallet
    env.FundWallet(t, tenant.ID, 100.00) // $100 USD equivalent
    
    // Step 4: Make chat completion request
    req := map[string]interface{}{
        "model": "gpt-4o",
        "messages": []map[string]string{{"role": "user", "content": "Say hello"}},
    }
    
    resp := env.MakeRequest(t, "POST", "/v1/chat/completions", key, req)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Step 5: Verify response format
    var completion contractsopenai.ChatCompletionResponse
    json.NewDecoder(resp.Body).Decode(&completion)
    assert.NotEmpty(t, completion.ID)
    assert.Len(t, completion.Choices, 1)
    
    // Step 6: Verify response headers
    assert.NotEmpty(t, resp.Header.Get("x-shahrah-trace-id"))
    assert.NotEmpty(t, resp.Header.Get("x-shahrah-cost-usd"))
    assert.NotEmpty(t, resp.Header.Get("x-shahrah-cost-irr"))
    
    // Step 7: Verify wallet was debited
    wallet := env.GetWallet(t, tenant.ID)
    assert.True(t, wallet.BalanceIRR < 100.00*env.ExchangeRate) // Balance decreased
    
    // Step 8: Verify spend log was recorded
    logs := env.GetSpendLogs(t, tenant.ID)
    assert.Len(t, logs, 1)
    assert.Equal(t, completion.Usage.TotalTokens, logs[0].TotalTokens)
}
```

#### 8.2 Streaming Integration Test

```go
func TestStreamingLifecycle(t *testing.T) {
    env := setupTestEnvironment(t)
    tenant := env.CreateTenant(t, "test-stream")
    key := env.CreateAPIKey(t, tenant.ID, "sk-test-stream")
    env.FundWallet(t, tenant.ID, 100.00)
    
    req := map[string]interface{}{
        "model": "gpt-4o",
        "messages": []map[string]string{{"role": "user", "content": "Count 1 to 5"}},
        "stream": true,
    }
    
    resp := env.MakeStreamingRequest(t, "POST", "/v1/chat/completions", key, req)
    assert.Equal(t, 200, resp.StatusCode)
    assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
    
    // Read SSE chunks
    reader := bufio.NewReader(resp.Body)
    chunkCount := 0
    for {
        line, err := reader.ReadString('\n')
        if err == io.EOF {
            break
        }
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                break
            }
            var chunk contractsopenai.ChatCompletionChunk
            json.Unmarshal([]byte(data), &chunk)
            chunkCount++
        }
    }
    
    assert.Greater(t, chunkCount, 0)
    
    // Verify billing after stream
    logs := env.GetSpendLogs(t, tenant.ID)
    assert.Len(t, logs, 1)
}
```

#### 8.3 Error Handling Integration Tests

```go
func TestErrorHandling(t *testing.T) {
    env := setupTestEnvironment(t)
    tenant := env.CreateTenant(t, "test-errors")
    
    t.Run("Invalid API Key", func(t *testing.T) {
        resp := env.MakeRequest(t, "POST", "/v1/chat/completions", "sk-invalid", nil)
        assert.Equal(t, 401, resp.StatusCode)
        assertAPIError(t, resp, "authentication_error")
    })
    
    t.Run("Tenant Not Found", func(t *testing.T) {
        key := env.CreateAPIKey(t, tenant.ID, "sk-test")
        resp := env.MakeRequestWithHost(t, "POST", "/v1/chat/completions", "unknown.sharah.ir", key, nil)
        assert.Equal(t, 404, resp.StatusCode)
    })
    
    t.Run("Budget Exceeded", func(t *testing.T) {
        key := env.CreateAPIKey(t, tenant.ID, "sk-test-budget")
        env.FundWallet(t, tenant.ID, 0.001) // Tiny balance
        req := map[string]interface{}{
            "model": "gpt-4o",
            "messages": []map[string]string{{"role": "user", "content": "Hello"}},
        }
        resp := env.MakeRequest(t, "POST", "/v1/chat/completions", key, req)
        assert.Equal(t, 402, resp.StatusCode)
        assertAPIError(t, resp, "budget_exceeded")
    })
    
    t.Run("Rate Limited", func(t *testing.T) {
        key := env.CreateAPIKey(t, tenant.ID, "sk-test-ratelimit", WithRPMLimit(1))
        req := map[string]interface{}{
            "model": "gpt-4o",
            "messages": []map[string]string{{"role": "user", "content": "Hello"}},
        }
        // First request succeeds
        resp := env.MakeRequest(t, "POST", "/v1/chat/completions", key, req)
        assert.Equal(t, 200, resp.StatusCode)
        
        // Second request blocked
        resp = env.MakeRequest(t, "POST", "/v1/chat/completions", key, req)
        assert.Equal(t, 429, resp.StatusCode)
        assertAPIError(t, resp, "rate_limit_error")
    })
}
```

#### 8.4 Load Testing

**File:** `services/control-plane/tests/benchmarks/load_test.go`

```go
func BenchmarkConcurrentRequests(b *testing.B) {
    env := setupBenchmarkEnvironment(b)
    key := env.CreateAPIKey(b, env.TenantID, "sk-bench")
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        client := &http.Client{Timeout: 30 * time.Second}
        for pb.Next() {
            req := createBenchmarkRequest(key)
            resp, err := client.Do(req)
            if err != nil {
                b.Errorf("Request failed: %v", err)
                continue
            }
            resp.Body.Close()
            if resp.StatusCode != 200 && resp.StatusCode != 502 {
                b.Errorf("Unexpected status: %d", resp.StatusCode)
            }
        }
    })
}
```

**Target:** 100+ req/sec on a single instance with mocked tunnel responses.

#### 8.5 Test Coverage Reporting

**Add to Makefile:**
```makefile
test-coverage:
    go test ./... -coverprofile=coverage.out -covermode=atomic
    go tool cover -html=coverage.out -o coverage.html
    go tool cover -func=coverage.out | grep total
```

**Target:** 80%+ code coverage for all core packages.

**Estimated Effort:** 3-4 days

---

## Phase 1 Summary

| Task | Estimated Effort | Status | Dependencies |
|------|-----------------|--------|--------------|
| Task 1: Streaming SSE Proxy | 3-4 days | ~85% - needs error transformation | None (extends existing) |
| Task 2: Tenant Resolution | 2-3 days | ~70% - needs caching | Task 1 (shared cache) |
| Task 3: Rate Limiting | 3-4 days | ~60% - sliding window exists | Task 2 (tenant aggregation) |
| Task 4: Billing Completion | ~~2-3 days~~ | ✅ **COMPLETE** (95%) | None (extends existing) |
| Task 5: Additional Endpoints | 4-5 days | ~20% - chat only | Tasks 1-4 |
| Task 6: Database Schema | 3-4 days | ~60% - needs verification | Tasks 1-5 (must match code) |
| Task 7: Provider Expansion | 4-5 days | ~10% - OpenAI only | Task 5 (transformation) |
| Task 8: Integration Testing | 3-4 days | ~40% - partial coverage | Tasks 1-7 |
| **Remaining** | **22-29 days** | **~5-6 weeks** | |

### Updated Assessment (2026-04-26)
- **Task 4 Billing** verified complete: `DebitWallet` + `LogSpend` implementations confirmed with `FOR UPDATE` row locking
- Total estimated effort reduced by 2-3 days
- Overall stabilization progress: **~50% complete** (up from ~40%)

---

## Risk Mitigation

| Risk | Impact | Mitigation | Status |
|------|--------|------------|--------|
| Database schema mismatch with Go structs | High | Run migration tests early, use code generation | Addressed with `TestIntegration_StorageSchemaContract` to validate required table columns and critical index coverage before release |
| Race condition in wallet debit | High | Use `FOR UPDATE` row locks, test with concurrent debits | Addressed with `FOR UPDATE` row lock + sentinel error (`ErrInsufficientWalletBalance`) and `TestIntegration_ConcurrentWalletDebitUsesRowLock` |
| Redis not available in production | Medium | Dual cache fallback (memory-only) with degraded rate limiting |
| Tunnel errors not properly handled | Medium | Add comprehensive error classification in tunnel transport |
| Provider API format changes | Low | Abstract provider interface, test with real API calls |

---

## Acceptance Criteria

Phase 1 is complete when:

- [ ] All 8 tasks completed with passing tests
- [ ] 80%+ code coverage on core packages
- [ ] Load test achieves 100+ req/sec on single instance
- [ ] Integration test covers full request lifecycle
- [ ] Error handling test validates all 4xx/5xx scenarios
- [ ] Database migrations apply cleanly to fresh database
- [ ] Seed data enables local development without manual setup
- [ ] Documentation updated with new endpoints and configurations

# GodarAI Multi-Tenant Architecture

## Overview

This document describes the multi-tenant architecture for Shahrah, designed to support both cloud-hosted SaaS and future on-premise deployments. The architecture ensures complete data isolation between tenants while maintaining operational efficiency for the cloud deployment model.

## 1. Tenant Isolation Model

### 1.1 Isolation Strategy: Row-Level (Schema-per-Tenant via Tenant ID)

For the cloud-hosted SaaS model, we use a **shared database with row-level isolation**:

```mermaid
graph TB
    subgraph "Shared Infrastructure"
        A[PostgreSQL Database]
        B[Redis Cache]
        C[Object Storage]
    end
    
    subgraph "Tenant Data Isolation"
        A --> D[Tenant 1 Data]
        A --> E[Tenant 2 Data]
        A --> F[Tenant N Data]
        
        D -. tenant_id=1 .-> D
        E -. tenant_id=2 .-> E
        F -. tenant_id=N .-> F
    end
    
    subgraph "Application Layer"
        G[Tenant Context Middleware]
        G --> H[Query Filter: WHERE tenant_id = ?]
    end
    
    G --> A
```

### 1.2 Why Row-Level Isolation?

| Factor | Decision |
|--------|----------|
| Cost Efficiency | Shared infrastructure reduces costs |
| Operational Simplicity | Single schema to manage, migrate, backup |
| Scalability | Easy to add new tenants |
| Data Size | Suitable for expected data volumes |
| Future On-Premise | Clean separation enables extraction |

### 1.3 Future On-Premise Readiness

The architecture is designed to support on-premise deployment in the future:

```mermaid
graph LR
    subgraph "Current: SaaS Model"
        A1[Shared App] --> B1[Shared DB with tenant_id]
    end
    
    subgraph "Future: On-Premise Model"
        A2[Tenant App Instance] --> B2[Tenant Dedicated DB]
    end
    
    C[Common Codebase] -. enables .-> A1
    C -. enables .-> A2
    
    style C fill:#e1f5fe
```

**Key Design Principles for On-Premise Readiness:**

1. **Tenant ID as First-Class Citizen**: Every database query includes `tenant_id` filtering
2. **Configuration Externalization**: All tenant-specific config stored in DB, not hardcoded
3. **Connection Abstraction**: Database access through a tenant-aware connection manager
4. **Feature Flags**: Per-tenant feature flags to enable/disable cloud-specific features
5. **External Service Abstraction**: Payment, email, etc. through interfaces that can be swapped

## 2. Tenant Context Middleware

### 2.1 Request Processing Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant T as Tenant Resolver
    participant A as Application
    participant D as Database
    
    C->>G: Request to tenant1.sharah.ir
    G->>T: Extract subdomain
    T->>T: Lookup tenant by api_subdomain
    T->>T: Verify tenant status
    T->>A: Inject tenant_id into request context
    A->>D: Query with tenant_id filter
    D->>A: Return tenant-scoped data
    A->>C: Response
```

### 2.2 Implementation Pattern

```go
// TenantContext holds tenant information for the current request.
type TenantContext struct {
    TenantID     string
    TenantSlug   string
    Config       map[string]interface{}
    FeatureFlags map[string]interface{}
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const tenantContextKey contextKey = "tenant"

// TenantMiddleware extracts tenant from subdomain and injects into request context.
func TenantMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            host := r.Host
            subdomain := extractSubdomain(host)

            if subdomain == "" {
                http.Error(w, "Invalid tenant subdomain", http.StatusBadRequest)
                return
            }

            var tenant ShahrahTenant
            if err := db.Where("api_subdomain = ?", subdomain).First(&tenant).Error; err != nil {
                http.Error(w, "Tenant not found", http.StatusNotFound)
                return
            }

            if tenant.Status != "ACTIVE" {
                http.Error(w, "Tenant is not active", http.StatusForbidden)
                return
            }

            // Set tenant context for this request
            ctx := &TenantContext{
                TenantID:     tenant.TenantID,
                TenantSlug:   tenant.TenantSlug,
                Config:       tenant.Config,
                FeatureFlags: tenant.FeatureFlags,
            }

            ctx_with_tenant := context.WithValue(r.Context(), tenantContextKey, ctx)
            next.ServeHTTP(w, r.WithContext(ctx_with_tenant))
        })
    }
}

// GetCurrentTenant retrieves the tenant context from the request context.
func GetCurrentTenant(ctx context.Context) (*TenantContext, error) {
    tenant, ok := ctx.Value(tenantContextKey).(*TenantContext)
    if !ok || tenant == nil {
        return nil, fmt.Errorf("no tenant context available")
    }
    return tenant, nil
}

// RequireTenantPermission checks if current tenant has specific permission.
func RequireTenantPermission(ctx context.Context, permission string) bool {
    tenant, err := GetCurrentTenant(ctx)
    if err != nil {
        return false
    }
    val, ok := tenant.FeatureFlags[permission]
    if !ok {
        return false
    }
    hasPermission, _ := val.(bool)
    return hasPermission
}
```

### 2.3 Database Query Filtering

```go
// TenantScopedQuery ensures all database queries include tenant_id filtering.
type TenantScopedQuery struct {
    DB       *gorm.DB
    TenantID string
}

// NewTenantScopedQuery creates a new TenantScopedQuery with the current tenant.
func NewTenantScopedQuery(ctx context.Context, db *gorm.DB) (*TenantScopedQuery, error) {
    tenant, err := GetCurrentTenant(ctx)
    if err != nil {
        return nil, err
    }
    return &TenantScopedQuery{
        DB:       db,
        TenantID: tenant.TenantID,
    }, nil
}

// FilterByCurrentTenant applies tenant filter to a GORM query.
func (q *TenantScopedQuery) FilterByCurrentTenant(query *gorm.DB) *gorm.DB {
    return query.Where("tenant_id = ?", q.TenantID)
}

// Find performs a tenant-scoped find operation.
func (q *TenantScopedQuery) Find(dest interface{}, conds ...interface{}) error {
    return q.FilterByCurrentTenant(q.DB).Find(dest, conds...).Error
}

// First performs a tenant-scoped first operation.
func (q *TenantScopedQuery) First(dest interface{}, conds ...interface{}) error {
    return q.FilterByCurrentTenant(q.DB).First(dest, conds...).Error
}
```

## 6. Security Considerations

### 6.1 Tenant Data Isolation

- All queries MUST include `tenant_id` filter
- API keys are tenant-scoped
- Cache keys include tenant prefix
- Logs include tenant_id for audit

### 6.2 Cross-Tenant Access Prevention

```go
// VerifyTenantOwnership verifies that a resource belongs to the specified tenant.
func VerifyTenantOwnership(db *gorm.DB, tenantID string, resourceID string) (bool, error) {
    var resource struct {
        TenantID string `gorm:"column:tenant_id"`
    }
    if err := db.Select("tenant_id").Where("id = ?", resourceID).First(&resource).Error; err != nil {
        return false, err
    }
    return resource.TenantID == tenantID, nil
}

// SanitizeTenantID validates tenant_id format to prevent injection.
func SanitizeTenantID(tenantID string) (string, error) {
    matched, err := regexp.MatchString(`^[a-f0-9-]{36}$`, tenantID)
    if err != nil {
        return "", err
    }
    if !matched {
        return "", fmt.Errorf("invalid tenant_id format")
    }
    return tenantID, nil
}
```
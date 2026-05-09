# Tenant-Admin Auth Decoupling Plan

## Current State

The tenant-admin application at [`apps/tenant-admin/lib/auth/store.ts`](apps/tenant-admin/lib/auth/store.ts) directly uses Prisma for:
- CAPTCHA challenge lifecycle (create, verify, delete)
- OTP request/verification storage
- Session creation, lookup, and revocation
- Tenant user find/create/update operations
- Runtime schema migrations

This tight coupling prevents:
- Testing without a real database
- Migrating auth to a separate service
- Swapping backends (e.g., Redis for sessions, separate auth service)

## Target Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                    tenant-admin (Next.js)                               │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  API Routes (app/api/auth/*/route.ts)                        │   │
│  │  - request-otp/route.ts                                     │   │
│  │  - verify-otp/route.ts                                      │   │
│  │  - captcha/route.ts                                         │   │
│  │  - logout/route.ts                                         │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  AuthService (lib/auth/service.ts)                           │   │
│  │  Orchestrates: captcha, otp, session, user operations     │   │
│  │  Depends on: AuthRepository port, NotificationPort       │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                     │
│         ┌────────────────────┴────────────────────┐             │
│         ▼                                         ▼             │
│  ┌─────────────────────┐               ┌─────────────────┐ │
│  │  AuthRepository     │               │  NotifyPort     │ │
│  │  (PORT/interface)   │               │  (email/sms)    │ │
│  └─────────────────────┘               └─────────────────┘ │
│         │                                                 │
│         ▼                                                 │
│  ┌─────────────────────────────────────────────────────┐            │
│  │  ApiAuthRepository (lib/auth/api-auth-repository.ts) │            │
│  │  Calls Go control-plane via HTTP                   │            │
│  └────────────��────────────────────────────────────────┘            │
│                              │                                     │
│                              ▼ HTTP/REST                           │
│  ┌─────────────────────────────────────────────────────────────┐      │
│  │           services/control-plane (Go)                    │      │
│  │  - POST /api/v1/auth/captcha      → create/verify captcha  │      │
│  │  - POST /api/v1/auth/otp        → create OTP            │      │
│  │  - POST /api/v1/auth/otp/verify→ verify OTP         │      │
│  │  - POST /api/v1/auth/session   → create session     │      │
│  │  - GET  /api/v1/auth/session/:id → get session    │      │
│  │  - DELETE /api/v1/auth/session/:id → revoke       │      │
│  │  - POST /api/v1/auth/user/find-or-create           │      │
│  │  - GET /api/v1/auth/user/:id                   │      │
│  │  - PATCH /api/v1/auth/user/:id                → update│      │
│  └─────────────────────────────────────────────────────────────┘      │
│                              │                                     │
│                              ▼                                     │
│  ┌─────────────────────────────────────────────────────────────┐      │
│  │  internal/storage/auth_repo.go  (PostgreSQL)               │      │
│  └─────────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

## Implementation Steps

### Phase 1: Define Repository Interface

Create [`apps/tenant-admin/lib/auth/repository.ts`](apps/tenant-admin/lib/auth/repository.ts) defining the port:

```typescript
export interface AuthRepository {
  // CAPTCHA
  createCaptchaChallenge(): Promise<CaptchaChallenge>;
  consumeCaptcha(id: string, answer: string): Promise<ConsumeResult>;

  // OTP
  createOTPRequest(params: CreateOTPParams): Promise<OTPRequest>;
  verifyOTPRequest(params: VerifyOTPParams): Promise<VerifyResult>;

  // Session
  createSession(params: CreateSessionParams): Promise<Session>;
  getSession(id: string): Promise<Session | null>;
  revokeSession(id: string): Promise<void>;

  // User
  findOrCreateUser(params: FindOrCreateUserParams): Promise<TenantUser>;
  getUser(tenantSlug: string, userID: string): Promise<TenantUser | null>;
  updateUserContact(params: UpdateUserContactParams): Promise<TenantUser>;
  touchLastLogin(tenantSlug: string, userID: string): Promise<TenantUser | null>;
}
```

### Phase 2: Create API Contract (Go backend)

Add new HTTP handlers in control-plane:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/captcha/create` | Create CAPTCHA challenge |
| POST | `/api/v1/auth/captcha/consume` | Verify and consume CAPTCHA |
| POST | `/api/v1/auth/otp/request` | Create OTP request |
| POST | `/api/v1/auth/otp/verify` | Verify OTP code |
| POST | `/api/v1/auth/session` | Create session |
| GET | `/api/v1/auth/session/{id}` | Get session |
| DELETE | `/api/v1/auth/session/{id}` | Revoke session |
| POST | `/api/v1/auth/user/find-or-create` | Find or create user |
| GET | `/api/v1/auth/user/{id}` | Get user by ID |
| PATCH | `/api/v1/auth/user/{id}` | Update user contact |
| POST | `/api/v1/auth/user/{id}/touch-last-login` | Update last login |

### Phase 3: Implement HTTP Adapter

Create [`apps/tenant-admin/lib/auth/api-auth-repository.ts`](apps/tenant-admin/lib/auth/api-auth-repository.ts):

```typescript
export class ApiAuthRepository implements AuthRepository {
  constructor(private baseURL: string, private apiKey: string) {}

  async createCaptchaChallenge(): Promise<CaptchaChallenge> {
    const res = await fetch(`${this.baseURL}/api/v1/auth/captcha/create`, {
      method: 'POST',
      headers: this.headers(),
    });
    return res.json();
  }
  // ... similar implementations for all methods
}
```

### Phase 4: Refactor API Routes

Update [`apps/tenant-admin/app/api/auth/*/route.ts`](apps/tenant-admin/app/api/auth) to inject repository:

```typescript
// In route handlers, inject via middleware or factory
const authRepo = getAuthRepository();
const result = await authRepo.createOTPRequest(params);
```

### Phase 5: Migration Strategy

1. Keep existing Prisma implementation as `PrismaAuthRepository` for backward compatibility during migration
2. Add feature flag `USE_API_AUTH=true` to switch between implementations
3. Run both in parallel for one release cycle, then remove Prisma dependency

### Phase 6: Schema Migration

Move DDL from [`ensureTenantAdminAuthSchema()`](apps/tenant-admin/lib/auth/store.ts:683) to [`db/main/migrations/`](db/main/migrations).

## Types to Extract

From store.ts, move these types to shared location ([`apps/tenant-admin/lib/auth/types.ts`](apps/tenant-admin/lib/auth/types.ts)):

- `AuthChannel`
- `OTPPurpose`
- `TenantUser`
- `TenantSession`
- `OTPRecord`
- Input/output DTOs for each operation

## Files to Create/Modify

| File | Action |
|------|--------|
| `apps/tenant-admin/lib/auth/repository.ts` | Create (port interface) |
| `apps/tenant-admin/lib/auth/types.ts` | Create (shared types) |
| `apps/tenant-admin/lib/auth/api-auth-repository.ts` | Create (HTTP adapter) |
| `apps/tenant-admin/lib/auth/service.ts` | Create (orchestration) |
| `apps/tenant-admin/lib/auth/prisma-auth-repository.ts` | Create (wrap existing as adapter) |
| `services/control-plane/internal/auth/handlers.go` | Extend with new endpoints |
| `services/control-plane/internal/storage/auth_repo.go` | Extend with new queries |

## Testing Strategy

1. Create [`apps/tenant-admin/lib/auth/inmemory-repository.ts`](apps/tenant-admin/lib/auth/inmemory-repository.ts) for unit tests
2. All service tests use in-memory implementation
3. Integration tests use API adapter pointing to Go backend
4. Contract tests validate API responses match TypeScript interfaces

## Implementation Status

### Phase 1: Repository Interface ✅
- Created `apps/tenant-admin/lib/auth/repository.ts` - AuthRepository port interface
- Created `apps/tenant-admin/lib/auth/types.ts` - Shared types

### Phase 2: Go Backend API Contract ✅
- Created `services/control-plane/internal/admin/auth_handlers.go` - HTTP handlers
- Created `services/control-plane/internal/storage/tenant_admin_auth_repo.go` - PostgreSQL storage
- Registered routes in `services/control-plane/internal/server/server.go`

### Phase 3: HTTP Adapter ✅
- Created `apps/tenant-admin/lib/auth/api-auth-repository.ts` - HTTP adapter

### Phase 4: Refactor API Routes ✅
- Updated all `apps/tenant-admin/app/api/auth/*/route.ts` files
- Updated `apps/tenant-admin/lib/auth/session.ts` and `request-auth.ts`
- Created `apps/tenant-admin/lib/auth/service.ts` - Orchestrator
- Created `apps/tenant-admin/lib/auth/factory.ts` - Factory with feature flag

### Phase 5: Migration Strategy ✅
- `PrismaAuthRepository` wraps existing logic
- Feature flag `USE_API_AUTH=true` switches implementations
- `InMemoryAuthRepository` available for testing

### Phase 6: Schema Migration
- DDL ready in `services/control-plane/internal/storage/tenant_admin_auth_repo.go`
- Should be moved to `db/main/migrations/`

## Security Considerations

- All auth API endpoints require authentication (master key or signed JWT)
- Rate limiting on OTP creation endpoints
- CAPTCHA before OTP to prevent brute force
- Session tokens are cryptographically secure (randomBytes(24))
- Tenant context isolation enforced at both API and service layers

## Security Considerations

- All auth API endpoints require authentication (e.g., master key or signed JWT)
- Rate limiting on OTP creation endpoints
- CAPTCHA before OTP to prevent brute force
- Session tokens should be cryptographically secure (already implemented via randomBytes(24))
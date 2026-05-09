# Tenant Admin Auth Decoupling - Breaking Changes & Migration Guide

## Overview

The tenant-admin authentication layer has been decoupled from direct Prisma database access into a repository pattern with two implementations:
1. **PrismaAuthRepository** - Local Prisma database (default, backward-compatible)
2. **ApiAuthRepository** - HTTP calls to Go control-plane service

## Breaking Changes

### 1. Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `USE_API_AUTH` | No | Set to `true` to switch to HTTP-based auth. Default: `false` (Prisma) |
| `CONTROL_PLANE_URL` | Yes (if USE_API_AUTH=true) | URL of the Go control-plane service. Default: `http://localhost:8080` |
| `SHAHRAH_MASTER_KEY` | Yes (if USE_API_AUTH=true) | Master API key for control-plane authentication |

### 2. API Response Format Changes

The Go backend auth endpoints use snake_case JSON keys instead of camelCase:

| Field | Old (Prisma) | New (Go API) |
|-------|-------------|--------------|
| request_id | `request_id` | `request_id` (same) |
| tenant_slug | `tenantSlug` | `tenant_slug` |
| user_id | `userID` | `user_id` |
| created_at_iso | `createdAtISO` | `created_at_iso` |
| expires_at_iso | `expiresAtISO` | `expires_at_iso` |
| pending_value | `pendingValue` | `pending_value` |

**Note**: The TypeScript API layer handles this normalization internally, so tenant-admin UI code is unaffected.

### 3. Removed Direct Store Imports

The following direct imports from `@/lib/auth/store` are deprecated:

```typescript
// OLD - Deprecated
import { createOTPRequest, verifyOTPRequest, createTenantSession } from "@/lib/auth/store";

// NEW - Use the service/factory
import { getAuthService } from "@/lib/auth/factory";
// or
import { getAuthRepository } from "@/lib/auth/factory";
```

### 4. Captcha Image Generation

Captcha SVG generation has been moved from the Go backend to the Next.js frontend. The `/api/v1/auth/captcha/create` endpoint now returns an empty `image_data_url`. The frontend should generate captcha images locally.

### 5. Database Schema Migration

The DDL previously created at runtime by `ensureTenantAdminAuthSchema()` should be moved to `db/main/migrations/`:

```sql
-- Run this migration before deploying the new code
-- Tables: tenant_admin_captchas, tenant_admin_sessions, tenant_admin_otp_requests
```

## Backward Compatibility

### Default Behavior (USE_API_AUTH not set or false)
- **Fully backward compatible** - uses Prisma directly like before
- No changes to existing API behavior
- All existing session cookies, OTP codes, captcha challenges continue to work

### API Mode (USE_API_AUTH=true)
- Requires Go control-plane to be running and configured
- Network latency may add ~50-200ms to auth operations
- Session state is centralized in Go service PostgreSQL

## Migration Steps

### Phase 1: Deploy Code with Feature Flag Off (Recommended)
```bash
# Deploy without any env changes - uses Prisma directly
USE_API_AUTH=false
```

### Phase 2: Deploy Go Auth Endpoints
```bash
# Ensure control-plane has the new auth routes registered
# Test manually: curl -X POST http://localhost:8080/api/v1/auth/captcha/create
```

### Phase 3: Enable API Mode (After Verification)
```bash
USE_API_AUTH=true
CONTROL_PLANE_URL=https://control-plane.internal:8443
SHAHRAH_MASTER_KEY=<your-master-key>
```

### Phase 4: Remove Prisma Dependency (Future)
After one full release cycle with API mode running stably, remove `PrismaAuthRepository` and the `tenant_admin_*` tables from the tenant-admin Prisma schema.

## Rollback

If issues arise after enabling API mode:
```bash
USE_API_AUTH=false
# Restart tenant-admin - it will fall back to Prisma
```

Sessions created via API mode will NOT be accessible in Prisma mode and vice versa, so users may need to re-login during rollback.

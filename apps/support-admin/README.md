# Support Admin

Internal Next.js application for Shahrah operations, support, and privileged catalog administration. It is deployed at the root of its own admin origin and has no Next.js `basePath`.

## Ownership

This app owns:

- Human authentication, opaque sessions, and RBAC
- Model and provider administration
- Import, indexing, and audit operations
- Display-setting mutations
- Tenant support and commercial operations through control-plane

Public model browsing, provider discovery, and comparison belong to `apps/model-discovery-web`.

## Run

From the repository root:

```powershell
pnpm --filter @shahrah/support-admin dev
```

The local app listens on `http://localhost:3001`. Running `pnpm dev` at the repository root starts `tenant-admin`, not this application. Use `pnpm dev:support-admin` from the root to avoid routing `/api/admin/*` to the wrong Next.js app.

Environment variables:

- `DATABASE_URL`: support-admin-owned database
- `MODEL_DISCOVERY_API_BASE_INTERNAL`: model-discovery API base; default `http://localhost:3002/api/v1`
- `MODEL_DISCOVERY_SERVICE_TOKEN`: scoped credential for protected model-discovery operations
- `CONTROL_PLANE_SERVICE_TOKEN`: scoped credential for platform and tenant support APIs
- `SHAHRAH_CONTROL_PLANE_URL`: internal control-plane origin
- `SUPPORT_ADMIN_ORIGIN`: exact externally visible origin used for CSRF validation
- `SUPPORT_ADMIN_COOKIE_PATH`: optional cookie path override; default `/`

The former `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` setting is obsolete and should be removed from local and deployment secrets.

## Authentication

Passwords are scrypt-hashed and browsers receive only opaque, revocable session cookies. Human credentials and browser sessions are never forwarded to model-discovery.

After applying migrations, provision or rotate an operator account without placing the password on the command line:

```powershell
$env:DATABASE_URL = "postgresql://.../support_admin?schema=public"
$env:SUPPORT_ADMIN_INITIAL_PASSWORD = "use-a-secret-manager-value"
pnpm --filter @shahrah/support-admin admin:user:create -- --email=operator@example.com --role=super_admin
```

Supported roles are `super_admin`, `catalog_admin`, and read-only `support_agent`. Updating a user with this command revokes that user's sessions.

Support-admin must not connect directly to tenant databases. Tenant source-of-truth reads and mutations go through control-plane.

## Routes

- `/` redirects to `/admin`
- `/login`
- `/admin`
- `/admin/providers`
- `/admin/models`
- `/admin/model-series`
- `/admin/pricing`
- `/admin/capabilities`
- `/admin/import`
- `/admin/indexing`
- `/admin/audit-log`
- `/admin/settings/display`

Privileged model-discovery requests pass through authenticated Next route handlers using `MODEL_DISCOVERY_SERVICE_TOKEN`. Provider assets are also proxied through an authenticated support-admin route so the internal service origin is not exposed to the browser.

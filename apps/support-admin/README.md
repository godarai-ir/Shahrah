# Support Admin / Catalog Frontend

Next.js UI for model discovery + admin workflows.

## Run
- `pnpm --filter @shahrah/support-admin dev`
- open `http://localhost:3000`

Configure API base with:
- `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` (default: `http://localhost:3002/api/v1`)
- `MODEL_CATALOG_API_BASE_INTERNAL` for server-to-server requests
- `DATABASE_URL` for the separate support-admin-owned database (never the tenant/control-plane database)
- `MODEL_DISCOVERY_SERVICE_TOKEN` for protected model-discovery requests
- `CONTROL_PLANE_SERVICE_TOKEN` for scoped platform and tenant support APIs
- `SUPPORT_ADMIN_ORIGIN` as the exact production origin used for CSRF validation

## Authentication

Human authentication is owned by `support-admin`. Passwords are scrypt-hashed and the browser receives only an opaque, revocable session cookie. `model-discovery` never receives a user's password or session.

After applying the database migration, provision or rotate an operator account without placing the password on the command line:

```powershell
$env:DATABASE_URL = "postgresql://.../support_admin?schema=public"
$env:SUPPORT_ADMIN_INITIAL_PASSWORD = "use-a-secret-manager-value"
pnpm --filter @shahrah/support-admin admin:user:create -- --email=emailtord@gmail.com --role=super_admin
```

Supported roles are `super_admin`, `catalog_admin`, and read-only `support_agent`. Updating an existing user with this command revokes all of that user's sessions.

`support-admin` must not connect directly to tenant databases. Tenant source-of-truth reads and mutations go through control-plane; local cross-tenant data is a minimal, eventually-consistent projection only.

## Model Discovery Contract

`support-admin` is a direct client of `model-discovery`.

- Live Swagger UI: `http://localhost:3002/docs`
- Live OpenAPI JSON: `http://localhost:3002/docs/openapi.json`
- Exported repo spec: [docs/model-discovery-api-contract.md](/D:/source/Shahrah/docs/model-discovery-api-contract.md)

Current client entrypoints:

- Public catalog client: [apps/support-admin/lib/api.ts](/D:/source/Shahrah/apps/support-admin/lib/api.ts)
- Local login/session handler: [apps/support-admin/app/api/auth/login/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/auth/login/route.ts)
- Admin providers proxy: [apps/support-admin/app/api/admin/providers/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/providers/route.ts)
- Admin models proxy: [apps/support-admin/app/api/admin/models/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/models/route.ts)

Implementation rule:

- Before changing any `model-discovery` request or response handling in `support-admin`, read the spec first and verify the route shape against the exported artifact.

## Routes
Public:
- `/`
- `/models`
- `/models/[providerSlug]/[modelSlug]`
- `/providers`
- `/providers/[providerSlug]`
- `/compare`
- `/categories/[categorySlug]`

Admin:
- `/admin`
- `/admin/providers`
- `/admin/models`
- `/admin/models/new`
- `/admin/models/[id]/edit`
- `/admin/pricing`
- `/admin/capabilities`
- `/admin/import`
- `/admin/indexing`
- `/admin/audit-log`

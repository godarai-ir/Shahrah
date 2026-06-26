# Support Admin AGENTS

Internal Next.js operations console for Shahrah.

## Ownership

This app owns:

- human authentication and opaque sessions
- RBAC for operators
- model and provider administration
- import, indexing, and audit workflows
- display-setting mutations
- tenant support and commercial operations through control-plane

Public model browsing belongs to `apps/model-discovery-web`.

## Working Rules

1. This app uses the Next.js App Router and has no `basePath`.
2. Run commands from the repository root with `pnpm --filter @shahrah/support-admin ...`.
3. Do not connect directly to tenant databases.
4. Use control-plane and service tokens for cross-system access.
5. Keep browser sessions opaque and revocable.

## Runtime Notes

- Local dev port: `3001`
- Root route redirects to `/admin`
- Admin routes live under `/admin/*`
- Auth routes live under `/api/auth/*`
- Admin API routes live under `/api/admin/*`

Environment variables commonly required:

- `DATABASE_URL`
- `MODEL_DISCOVERY_API_BASE_INTERNAL`
- `MODEL_DISCOVERY_SERVICE_TOKEN`
- `CONTROL_PLANE_SERVICE_TOKEN`
- `SHAHRAH_CONTROL_PLANE_URL`
- `SUPPORT_ADMIN_ORIGIN`
- `SUPPORT_ADMIN_COOKIE_PATH`

The older `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` setting is obsolete.

## Commands

```bash
pnpm --filter @shahrah/support-admin dev
pnpm --filter @shahrah/support-admin build
pnpm --filter @shahrah/support-admin lint
pnpm --filter @shahrah/support-admin admin:user:create -- --email=operator@example.com --role=super_admin
```

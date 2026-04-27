# Support Admin

Internal Next.js application for Shahrah support and operations teams.

## Purpose

- support ticket handling
- customer 360 views
- billing and operational review
- synced tenant state from the main platform

## Environment

Create `apps/support-admin/.env.local`:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sharah?sslmode=disable
```

## Commands

Run from the repository root:

```bash
pnpm --filter @shahrah/support-admin prisma:generate
pnpm --filter @shahrah/support-admin dev
pnpm --filter @shahrah/support-admin build
pnpm --filter @shahrah/support-admin lint
```

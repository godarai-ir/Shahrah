# Tenant Admin

## Purpose

- API key management
- model management
- wallet and top-up views
- Persian UX (`fa-IR`) with full RTL layout

Current implementation is wired to existing control-plane management APIs:

- `GET/POST /api/v1/keys`
- `GET/POST /api/v1/models`
- `GET /api/v1/wallet`
- `POST /api/v1/wallet/topup`

Authentication flow implemented in app routes:

- `GET /login` with image captcha + OTP sign-in
- `POST /api/auth/request-otp`
- `POST /api/auth/verify-otp`
- `POST /logout` and `POST /api/auth/logout`
- `GET /settings` for profile contact updates with OTP:
  - `POST /api/auth/profile/request-contact-change`
  - `POST /api/auth/profile/verify-contact-change`

Notes:

- OTP sender is currently stubbed (server logs + `debug_otp` in non-production).
- Auth data persistence is database-backed via Prisma models:
  - users
  - sessions
  - otp requests
  - captcha challenges
- Database schema is auto-bootstrapped on first auth request.

## Environment

Create `apps/tenant-admin/.env.local`:

```bash
SHAHRAH_CONTROL_PLANE_URL=http://127.0.0.1:4000
SHAHRAH_MASTER_KEY=sk-master
SHAHRAH_DEFAULT_TENANT=default
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sharah_db?sslmode=disable
```

For hostname-based tenant resolution in local development, map a tenant host:

```text
127.0.0.1 tenant1.localhost
```

Then run the app and open:

`http://tenant1.localhost:3000`

## Commands

Run from the repository root:

```bash
pnpm --filter @shahrah/tenant-admin prisma:generate
pnpm --filter @shahrah/tenant-admin dev
pnpm --filter @shahrah/tenant-admin build
pnpm --filter @shahrah/tenant-admin lint
```

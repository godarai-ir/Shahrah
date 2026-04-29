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
- `POST /api/v1/wallet/topup/intents`
- `GET|POST /api/v1/wallet/topup/callback/:provider`
- `GET /api/v1/wallet/transactions`
- `GET /api/v1/wallet/topup/settings`

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

Wallet top-up UX flow in tenant-admin:

1. User opens `/billing/topup`, enters top-up amount in USD and selects provider (`mock`, `zarinpal`, `idpay`).
2. Frontend calls `POST /api/wallet/topup` (app route proxy to control-plane intent endpoint).
3. On success, app redirects browser to returned `payment_url`.
4. After provider callback/verification, user lands on `/billing/topup/result` and can return to dashboard.
5. Dashboard also reads dynamic top-up constraints (`min/max/tax/fee`) and shows recent wallet transactions.

## Environment

Create `apps/tenant-admin/.env.local`:

```bash
SHAHRAH_CONTROL_PLANE_URL=http://127.0.0.1:4000
SHAHRAH_MASTER_KEY=-sk-dev-master-key-12345
SHAHRAH_DEFAULT_TENANT=default
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sharah_db?sslmode=disable
EMAIL_SENDING_ENABLED=true
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_SECURE=false
SMTP_FROM=no-reply@shahrah.local
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

## Dev Email Mock (Mailpit)

Run local compose stack (includes Mailpit):

```bash
docker compose -f infra/compose/local.yml up -d mailpit
```

Mailpit endpoints:

- SMTP: `127.0.0.1:1025`
- Web UI: `http://127.0.0.1:8025`

# Tenant Admin

## Purpose

- API key management
- model management
- wallet and top-up views
- Persian UX (`fa-IR`) with full RTL layout

Current implementation is wired to existing control-plane management APIs:

- `GET/POST /api/v1/models`
- `GET /api/v1/wallet`
- `POST /api/v1/wallet/topup/intents`
- `GET|POST /api/v1/wallet/topup/callback/:provider`
- `GET /api/v1/wallet/transactions`
- `GET /api/v1/wallet/topup/settings`
- `GET/POST /api/v1/personal-access-tokens`
- `GET /api/v1/personal-access-tokens/exists`
- `DELETE /api/v1/personal-access-tokens/:id`
- `DELETE /api/v1/personal-access-tokens/revoke/all`
- `GET /api/v1/traces`
- `GET /api/v1/traces/:trace_id`
- `GET /api/v1/traces/:trace_id/spans`

Authentication flow implemented in app routes:

- `GET /login` with image captcha + OTP sign-in
- `GET /login` with Google, GitHub, and Microsoft sign-in through the central auth domain
- `GET /auth/handoff` for one-time tenant session handoff redemption
- `POST /api/auth/social/link-ticket` for explicit account linking after recent authentication
- `POST /api/auth/request-otp`
- `POST /api/auth/verify-otp`
- `POST /logout` and `POST /api/auth/logout`
- `GET /settings` for profile contact updates with OTP:
  - `POST /api/auth/profile/request-contact-change`
  - `POST /api/auth/profile/verify-contact-change`

Notes:

- tenant-admin authentication always uses control-plane auth APIs.
- tenant-admin has no direct database access and no Prisma dependency.
- Local/dev email delivery uses SMTP (Mailpit in local setup); production uses Resend when `NODE_ENV=production`.

Wallet top-up UX flow in tenant-admin:

1. User opens `/billing/topup`, enters top-up amount in USD and selects provider (`mock`, `zarinpal`, `idpay`).
2. Frontend calls `POST /api/wallet/topup` (app route proxy to control-plane intent endpoint).
3. On success, app redirects browser to returned `payment_url`.
4. After provider callback/verification, user lands on `/billing/topup/result` and can return to dashboard.
5. Dashboard also reads dynamic top-up constraints (`min/max/tax/fee`) and shows recent wallet transactions.

PAT UX flow in tenant-admin:

1. User opens `/personal-access-tokens`.
2. Frontend lists user-owned PATs and allows search/refresh.
3. User creates a PAT with token name and optional expiry.
4. Raw token is displayed once in success panel; subsequent list responses never include raw token.
5. User can revoke a single PAT or revoke all own PATs.

Trace UX flow in tenant-admin:

1. User opens `/traces`.
2. Frontend fetches trace list via tenant-scoped control-plane API with filters/pagination.
3. User opens `/traces/:trace_id` for detailed request summary and span timeline.

## Environment

Configure the tenant-admin connection in `apps/tenant-admin/.env.local`:

```bash
SHAHRAH_CONTROL_PLANE_URL=http://127.0.0.1:4000
SHAHRAH_MASTER_KEY=-sk-dev-master-key-12345
GODARAI_AUTH_BASE_URL=http://auth.localhost:3000
MODEL_CATALOG_V2_ENABLED=true
```

`GODARAI_AUTH_BASE_URL` is server-only. Tenant login buttons first call a
same-origin redirect route, which resolves the tenant from the request host and
then redirects to the central authentication origin.

`MODEL_CATALOG_V2_ENABLED` is a server-side rollout flag. It defaults to `false`; enable it explicitly in local development or staging to use the Providers/Models catalog experience.

OTP delivery is owned by `control-plane`. Configure local email delivery on that process:

```bash
EMAIL_SENDING_ENABLED=true
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_SECURE=false
SMTP_FROM=no-reply@shahrah.local
```

For mobile OTP delivery through Bale Safir, configure these server-only variables on `control-plane`:

```bash
BALE_OTP_ENABLED=true
BALE_SAFIR_API_ACCESS_KEY=replace-with-safir-api-access-key
BALE_SAFIR_BOT_ID=123456789
```

Production email env on `control-plane`:

```bash
NODE_ENV=production
RESEND_API_KEY=replace-in-production-secret-store
RESEND_FROM_EMAIL=auth@notifications.godarai.ir
RESEND_FROM_NAME=گدارAI
RESEND_REPLY_TO=support@godarai.ir
```

`RESEND_FROM_EMAIL` must use the verified `notifications.godarai.ir`
subdomain. Missing values and legacy root-domain senders fall back to
`auth@notifications.godarai.ir`.

Tenants are not seeded and there is no default tenant. Create a tenant through
the website onboarding flow first. For local hostname-based tenant resolution,
map the tenant host after provisioning:

```text
127.0.0.1 mytenant.localhost
```

Then run the app and open the provisioned tenant login URL:

`http://mytenant.localhost:3001/login`

## Commands

Run from the repository root:

```bash
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

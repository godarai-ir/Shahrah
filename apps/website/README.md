This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Environment

Create `apps/website/.env` or `apps/website/.env.local`:

```bash
SHAHRAH_CONTROL_PLANE_URL=http://127.0.0.1:4000
SHAHRAH_MASTER_KEY=replace-with-control-plane-master-key
DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:5432/shahrah
BETTER_AUTH_SECRET=replace-with-at-least-32-random-bytes
GODARAI_AUTH_BASE_URL=http://auth.localhost:3000
GODARAI_TENANT_ORIGIN_TEMPLATE=http://{tenant}.localhost:3001
SHAHRAH_TENANT_BASE_DOMAIN=localhost
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_TENANT_ID=common
EMAIL_SENDING_ENABLED=true
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_SECURE=false
SMTP_FROM=no-reply@shahrah.local
```

Social sign-in is implemented with pinned `better-auth@1.6.25`. Register these
exact OAuth/OIDC callback URLs with the providers (use HTTPS in production):

- `https://auth.godarai.ir/api/auth/callback/google`
- `https://auth.godarai.ir/api/auth/callback/github`
- `https://auth.godarai.ir/api/auth/callback/microsoft`

The central auth cookie is host-only for `auth.godarai.ir`; tenant applications
receive only a hashed, single-use, 90-second handoff code. Run migration
`056_social_auth.sql` before enabling any provider. Do not expose provider
secrets, `BETTER_AUTH_SECRET`, `DATABASE_URL`, or the control-plane master key
through `NEXT_PUBLIC_*` variables.

The GitHub Actions secret names used by the production sync workflow are
`OAUTH_GITHUB_CLIENT_ID` and `OAUTH_GITHUB_CLIENT_SECRET`. The workflow maps
them to the runtime Kubernetes keys `GITHUB_CLIENT_ID` and
`GITHUB_CLIENT_SECRET`; no GitHub Actions secret name starts with `GITHUB_`.

Production email env:

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

The website contact form sends submissions to `support@godarai.ir` through this
server-side email configuration. The Resend API key is never exposed to the browser.

For local email capture, run Mailpit:

```bash
docker compose -f infra/compose/local.yml up -d mailpit
```

Mailpit endpoints:

- SMTP: `127.0.0.1:1025`
- Web UI: `http://127.0.0.1:8025`

## Getting Started

First, run the development server:

```bash
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.

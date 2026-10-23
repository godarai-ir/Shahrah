# Website AGENTS

Public Shahrah marketing site and onboarding funnel.

## Ownership

This app owns:

- landing pages and feature pages
- pricing, about, contact, privacy, and terms pages
- login, signup, and onboarding flows
- sitemap and robots routes
- public marketing APIs such as signup, contact, login discovery, and onboarding

`apps/blog` owns the blog content and `apps/docs` owns product documentation.

## Working Rules

1. This is a Next.js App Router app.
2. The site is RTL and uses Estedad (Vazirmatn as fallback) for Persian-facing UI; follow `DESIGN.md`.
3. Keep browser-visible content free of server secrets.
4. Keep Persian copy UTF-8 and avoid mojibake or corrupted strings.
5. Run commands from the repository root with `pnpm --filter @shahrah/website ...`.

## Runtime Notes

- Public site routes live under `app/(site)`.
- Login, signup, and onboarding live in dedicated route groups outside the marketing shell.
- The site uses server-side email delivery through SMTP in local dev and Resend in production.

## Commands

```bash
pnpm --filter @shahrah/website dev
pnpm --filter @shahrah/website build
pnpm --filter @shahrah/website lint
```

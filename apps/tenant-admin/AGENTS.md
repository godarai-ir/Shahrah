# tenant-admin AGENTS

## Purpose
Private tenant-facing admin console for Shahrah. This app owns:

- virtual models and provider accounts
- virtual users and access tokens
- teams and org units
- billing top-up and plan views
- trace inspection and privacy settings
- account settings and profile updates

This is a Next.js App Router app with `next-intl` for UI copy.

## Non-Negotiable Defaults

1. Never ship hardcoded user-facing strings in components.
2. Use `next-intl` keys (`useTranslations(...)`) for labels, helper text, buttons, empty states, and messages.
3. Keep the product language precise for LLM models, providers, accounts, and access policy flows.
4. Treat copy quality as product behavior, not decoration.

## i18n Source of Truth

1. Primary locale is `fa` (`i18n/request.ts`).
2. Source of truth is `messages/modules/<namespace>/<locale>.json`.
3. Keep `fa` and `en` structurally aligned per namespace.
4. Add both locales in the same change whenever a key changes.

## Current UI Flows

- Provider account setup uses a 3-step flow: provider selection, provider-specific connection fields, then models plus access policy.
- Billing top-up lives under `/billing/topup` and the result page is `/billing/topup/result`.
- PAT management lives under `/personal-access-tokens`.
- Traces live under `/traces` and `/traces/[traceID]`.
- Core admin areas include `/virtual-models`, `/virtual-users`, `/teams`, `/settings`, and `/billing/plan`.

## Integration Rules

1. Use control-plane APIs through the app routes in this app.
2. Do not add direct database or Prisma access here.
3. Keep local email delivery on SMTP/Mailpit and production email on Resend.

## Encoding and Corruption Guardrails

1. Files must stay UTF-8.
2. Do not introduce mojibake, replacement characters, or `????` strings.
3. Before finalizing i18n changes, scan locale files for corrupted strings.

## Implementation Checklist for i18n Edits

1. Update `messages/modules/<namespace>/fa.json`.
2. Mirror the same structure in `messages/modules/<namespace>/en.json`.
3. Replace any new hardcoded UI text with translation keys.
4. Run `pnpm --filter @shahrah/tenant-admin lint`.
5. Run `pnpm --filter @shahrah/tenant-admin validate:i18n`.
6. Verify no corrupted locale values remain.

## Commands

Run from the repository root:

```bash
pnpm --filter @shahrah/tenant-admin dev
pnpm --filter @shahrah/tenant-admin build
pnpm --filter @shahrah/tenant-admin lint
pnpm --filter @shahrah/tenant-admin validate:i18n
```

## UI / Frontend Work

Before making UI changes:

1. Read `../../docs/design-system/DESIGN.md`.
2. Read `../../docs/design-system/COMPONENT_USAGE.md`.
3. Inspect existing `components/ui` and `components/app-ui`.
4. Reuse existing product-level components before creating new ones.
5. Use semantic Tailwind tokens.
6. Avoid raw color utilities and arbitrary values.
7. Preserve RTL behavior.
8. Update `/design-system` when adding reusable UI patterns.
9. Run available lint/typecheck/build commands before finishing.
10. Summarize any UI debt that remains.

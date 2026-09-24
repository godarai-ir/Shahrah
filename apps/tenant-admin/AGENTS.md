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

Read `ARCHITECTURE.md` before changing application structure, feature ownership, data fetching, API contracts, or client/server boundaries. Its dependency direction is enforced by `pnpm run validate:architecture`.

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
5. A new namespace must also be listed in `MESSAGE_NAMESPACES` in `i18n/request.ts`
   and in the namespace list of every layout that needs it.

## Language Switching and Direction

1. The active locale is resolved per request from the `NEXT_LOCALE` cookie, then
   from `Accept-Language`, then from the `fa` default. `lib/locale.ts` owns the
   supported list and the narrowing helpers.
2. `components/locale-toggle.tsx` writes the cookie and refreshes the route.
   `LocaleToggle` is the standalone control used on `/login`; `LocaleToggleMenuItem`
   is the row in the admin profile menu.
3. Direction follows the locale: `fa` is RTL and `en` is LTR. Derive it with
   `getLocaleDirection` and never hardcode `dir="rtl"` or physical utilities such
   as `text-left`, `text-right`, `ml-*`, or `pr-*`. The design system guardrail
   test enforces this.
4. Format numbers and dates through `hooks/use-locale-format.ts`, or through the
   `getIntlLocale` and `getCalendarLocale` helpers on the server. Do not construct
   an `Intl` formatter with a hardcoded locale tag.
5. Billing keeps the Jalali calendar in both locales because billing periods are
   anchored to it; only the script and month names change.

## Current UI Flows

- Provider account setup uses a 3-step flow: provider selection, provider-specific connection fields, then models plus access policy.
- Guardrails are created at `/guardrails/new` and edited at `/guardrails/[slug]/edit` through a 4-step wizard: type, settings, behavior, then name and review. `features/guardrails/catalog.ts` is the single source of type metadata.
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
4. Lint the changed locale-related source files using the fast validation workflow below.
5. Run `pnpm --filter @shahrah/tenant-admin validate:i18n`.
6. Verify no corrupted locale values remain.

## Fast Validation Workflow

Lint and production builds for this app can be slow on Windows. Use progressive
validation and do not repeatedly run repository-wide commands.

1. During implementation, lint only changed JavaScript and TypeScript files.
   Run the command from `apps/tenant-admin` and pass paths relative to that
   directory:

   ```bash
   pnpm exec eslint <changed-file-1> <changed-file-2> --cache --cache-location .next/cache/eslint/.eslintcache
   ```

2. When a broader type check is useful, prefer the incremental compiler before
   running a production build:

   ```bash
   pnpm exec tsc --noEmit --incremental
   ```

3. Before finishing a source-code change, run the app-scoped Turbo lint from the
   repository root. Turbo can reuse the result when the app inputs have not
   changed:

   ```bash
   pnpm turbo run lint --filter=@shahrah/tenant-admin
   ```

4. Run the app-scoped production build only when the change can affect
   compilation, routing, rendering, dependencies, environment handling, or
   runtime behavior:

   ```bash
   pnpm turbo run build --filter=@shahrah/tenant-admin
   ```

5. Documentation-only, test-only, planning, or other non-production changes do
   not require a production build. State in the final summary why it was skipped.
6. Do not use root `pnpm lint` or root `pnpm build` to validate tenant-admin-only
   work; those commands fan out across the monorepo.
7. Do not delete `.next` or `.turbo` as routine cleanup because doing so removes
   reusable caches. Clear them only when diagnosing suspected stale output.
8. If a full command is unexpectedly slow, do not immediately rerun the same
   command. Report the elapsed time, preserve any output, and use the narrower
   checks above to identify whether linting, type checking, or bundling is slow.
9. Run `pnpm --filter @shahrah/tenant-admin validate:source-encoding` for source or locale changes. App lint runs it automatically.

## Commands

Run from the repository root:

```bash
pnpm --filter @shahrah/tenant-admin dev
pnpm turbo run build --filter=@shahrah/tenant-admin
pnpm turbo run lint --filter=@shahrah/tenant-admin
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
9. Follow the fast validation workflow above; run a production build when the
   change affects production code or configuration.
10. Summarize any UI debt that remains.

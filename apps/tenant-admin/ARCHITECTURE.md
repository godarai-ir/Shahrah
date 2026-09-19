# Tenant Admin Architecture

This document is the source of truth for code ownership and dependency direction in `tenant-admin`. Feature implementation under `app/components` is forbidden; `app` contains only Next.js entry points and their smallest route-local composition.

## Dependency Direction

```text
app ───────────────► features ───────────────► components + lib
 │                       │                           ▲
 └───────────────────────┴───────────────────────────┘
```

- `app`: Next.js routes, layouts, server page composition, and Route Handler adapters. It must not own feature implementation.
- `features/<feature>`: one business capability, including its UI orchestration, API client, schemas, domain rules, and tests.
- `components/ui`: shadcn primitives. Feature code must consume these wrappers instead of importing primitive libraries directly.
- `components/app-ui`: reusable product-level UI with no feature-specific data fetching or permissions.
- `lib`: shared runtime infrastructure and domain-neutral helpers. It must not import `app`, `features`, or React UI.
- `hooks`: shared domain-neutral hooks. A feature-specific hook stays in its feature.

`components` and `lib` together are the shared layer. Do not create another shared UI hierarchy.

The observe area (`/logs`, `/traces`, `/metrics`) is three features sharing one filter state. Because features must not import each other, the shared filter model lives in `lib/observe-filters.ts` and its query bar in `components/app-ui/ObserveQueryBar.tsx`. Each feature maps the shared filters onto its own and keeps view-specific filters to itself.

## Feature Shape

Use only the files and folders a feature needs. Small features may stay flat; introduce `api`, `domain`, `server`, `ui`, or `tests` subfolders only when each folder has more than one cohesive module:

```text
features/virtual-models/
  api/
    client.ts
    schemas.ts
  domain/
    reducer.ts
    types.ts
  server/
    service.ts
  ui/
    VirtualModelList.tsx
    VirtualModelForm.tsx
  tests/
```

Keep page entry points thin. A page reads server-owned session, permissions, translations, and initial data, then composes a feature entry point. Next.js Route Handlers validate transport input and delegate to feature or control-plane infrastructure.

Broad barrel files are discouraged because they can merge unrelated route graphs. Import an explicit feature entry point. Cross-feature imports are forbidden; composition across capabilities belongs in `app`, while genuinely shared behavior moves down to `components`, `hooks`, or `lib`.

## Client and Server Boundaries

- Files are Server Components by default.
- Put `"use client"` on the smallest interactive leaf, not a whole route tree.
- Browser modules must never import a `server` entry point or a module marked with `server-only`.
- Initial route data, session, and permissions belong on the server when possible.
- Use React Query for client-owned server state that is refreshed or mutated repeatedly.
- Put shareable filters and pagination in the URL.
- Use reducers for multi-step workflows with coupled transitions; avoid unrelated `useState` collections.

## HTTP and Contracts

- UI components do not define generic request helpers.
- Browser transport lives in `lib/http`; feature endpoint functions live in `features/<feature>/api`.
- Direct `fetch` calls in a feature are restricted to its API adapter; components call named feature operations.
- Every external payload is untrusted at runtime. Parse it with a shared Zod schema or a generated control-plane contract before exposing a typed domain value.
- Route Handlers use the shared authentication, error, body parsing, and control-plane proxy helpers.
- Error responses use stable error codes. Display copy is localized at the UI boundary.

## Permissions

Permission identifiers come from the canonical permission registry in `lib/permissions`. Do not add ad-hoc permission strings to components. UI visibility is a convenience; authorization remains enforced by the control plane.

## UI Ownership

Follow `docs/design-system/DESIGN.md`, `docs/design-system/COMPONENT_USAGE.md`, and the local `AGENTS.md`.

- Reuse `components/app-ui` before adding a feature-local visual pattern.
- A reusable product component must not fetch feature data.
- Feature forms own their domain schema and submission; shared form components own presentation and accessibility.
- Preserve RTL, semantic tokens, and localized copy.

See `docs/adr/0001-frontend-library-policy.md` for the approved primitive, chart, date, and server-state libraries.

## Executable Guardrails

`pnpm run validate:architecture` checks layer direction, direct cross-feature dependencies, server imports from client modules, feature transport placement, forbidden `app/components` ownership, and local dependency cycles.

`pnpm run validate:source-encoding` rejects invalid UTF-8, replacement characters, question-mark corruption, and common mojibake. Both run as part of app lint.

The performance boundary test is a ratchet, not an acceptable final size. Lower a component's budget whenever orchestration is extracted, and never increase a budget to make a test pass.

## Agent Checklist

Before editing:

1. Identify the owning feature and its existing API, domain, UI, translations, and tests.
2. Read the route entry point and follow dependencies downward only.
3. Check `components/app-ui` before creating reusable UI.
4. Confirm whether the module is server-only, client-only, or shared.

Before finishing:

1. Run architecture and encoding validation.
2. Run lint on changed files, incremental typecheck, related tests, and the app-scoped Turbo gates required by `AGENTS.md`.
3. Update this document or an ADR when introducing a new architectural decision.

# UI Audit

## Current Stack

- Next.js router type: App Router under `apps/tenant-admin/app`.
- Tailwind version/config: Tailwind CSS v4 using `@theme inline` in `apps/tenant-admin/app/globals.css`.
- shadcn/ui setup: `apps/tenant-admin/components.json`, `new-york`, RSC enabled, CSS variables enabled, lucide icons, RTL enabled.
- Global CSS location: `apps/tenant-admin/app/globals.css`, importing `apps/tenant-admin/design-system/tokens.css`.
- Component directories: `apps/tenant-admin/components/ui` for shadcn primitives, `apps/tenant-admin/app/components` and `apps/tenant-admin/features` for product and feature UI, new `apps/tenant-admin/components/app-ui` for product-level reusable UI.
- RTL status: `html dir="rtl"`, body direction RTL, `components.json` has `rtl: true`. Some components still use physical left/right spacing that requires review.

## Observed Design Issues

- Repeated arbitrary text sizes such as `text-[11px]`, `text-[13px]`, and route-specific clamp sizing.
- Some raw slate classes remain in code display components where a semantic code surface should be introduced later.
- Physical positioning utilities such as `left-*`, `right-*`, `ml-*`, and `mr-*` appear in shadcn primitives and feature components; some are intentional, some require RTL review.
- Several pages use hand-composed page headers, cards, stats, empty states, status chips, and destructive dialogs.
- Dense table/list pages use different toolbar and empty-state layouts.
- Settings and forms use feature-local grouping instead of a shared `FormSection` pattern.
- Status colors are partly semantic and partly direct CSS variable expressions.

## Existing Useful Patterns

- The app already has semantic tokens in `design-system/tokens.css`.
- shadcn primitives are established under `components/ui`.
- `AdminShell` centralizes the authenticated admin layout.
- Existing `Card`, `Badge`, `Button`, `Dialog`, `Skeleton`, and data-table primitives provide a good base.
- RTL is already a first-class app setting.

## Recommended Migration Strategy

Standardize page headers, section headers, cards, metric cards, empty states, status badges, and form sections first. Next, migrate repeated list/table pages to common toolbar and empty-state patterns. Leave dense feature-specific layouts, code blocks, charts, and trace timelines for later passes because they have higher regression risk.

## Representative Migrations Completed

- `app/components/stat-card.tsx` now composes `MetricCard` and `StatusBadge` while preserving its public props.
- `app/components/status-option-card.tsx` now uses `StatusBadge` for repeated badge styling.
- `app/components/profile-settings.tsx` now uses `PageHeader`, `SettingsCard`, and `FormSection` for account settings layout while preserving form behavior.

## Guardrail Search Findings

Searched `apps/tenant-admin/app`, `apps/tenant-admin/components`, and `apps/tenant-admin/features` for raw gray/slate/zinc colors, arbitrary values, and physical RTL utilities. Safe replacements were limited in this pass to new components and representative migrations. Remaining debt includes shadcn primitive physical offsets, feature search inputs using `left-*`/`pl-*`, dense model cards with arbitrary text sizes, trace timeline physical positioning, and code blocks with raw slate colors.

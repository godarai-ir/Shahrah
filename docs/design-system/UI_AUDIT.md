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
- `app/components/personal-access-token-management.tsx` now uses `AppCard`, `SectionHeader`, `EmptyState`, and `StatusBadge` for a representative list/table page.

## Guardrail Search Findings

Searched `apps/tenant-admin/app`, `apps/tenant-admin/components`, and `apps/tenant-admin/features` for raw gray/slate/zinc colors, arbitrary values, and physical RTL utilities. Safe replacements were limited in this pass to new components and representative migrations. Remaining debt includes shadcn primitive physical offsets, feature search inputs using `left-*`/`pl-*`, dense model cards with arbitrary text sizes, trace timeline physical positioning, and code blocks with raw slate colors.

## Review Classification

### Safe to fix now

- Product-level components should keep mobile layouts flexible and avoid forcing horizontal card headers.
- Migrated option cards should use logical text alignment and mark decorative icons as hidden from assistive technology.

### Intentional or acceptable for now

- Some shadcn primitive files contain physical offsets and arbitrary values inherited from the installed component implementation.
- Code display components use slate tones for syntax/code surfaces; migrate them only when a semantic code-surface token is introduced.
- Dense trace and model cards use compact arbitrary text sizes for highly constrained data displays.

### Should be migrated later

- Feature search inputs using physical `left-*`/`pl-*` should move to logical `start-*`/`ps-*` after visual verification.
- Table/list pages should converge on shared toolbar, empty, error, and pagination patterns.
- Form-heavy pages should move gradually toward `FormSection` or shadcn field primitives without changing validation behavior.

## Review Follow-up Completed

- Added semantic code-surface tokens and migrated `app/components/code-block.tsx` away from raw slate surface classes.
- Converted obvious search and date-input icon offsets from physical `left-*`/`pl-*` to logical `start-*`/`ps-*` in representative feature components.
- Kept ambiguous diff panes, side-specific sheets, and dense trace/model layouts untouched for later visual review.

## Dropdown/List Follow-up Completed

- Added `RowActionsMenu`, `AppSelectField`, and `FilterCombobox` under `apps/tenant-admin/components/app-ui`.
- Migrated representative row action menus in `team-management.tsx` and `users-access-management.tsx`.
- Migrated team filter `<select>` controls to `AppSelectField`.
- Routed trace filter combobox rendering through `FilterCombobox`.
- Updated `/design-system` with static examples for select, filter combobox, and row actions.

## Remaining Dropdown/List Debt

- `components/ui/dropdown-menu.tsx` and `components/ui/select.tsx` still contain physical RTL classes such as `right-*`, `pr-*`, `pl-*`, and `mr-auto`; update primitives carefully in a separate pass with visual checks.
- Several feature forms still use raw `Select` directly. Migrate them gradually when touching those forms.
- Destructive row actions now have a common visual pattern, but not every destructive action is wrapped in a confirmation dialog yet.

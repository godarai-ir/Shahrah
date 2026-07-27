# Design System

## Product Context

This application is a B2B SaaS / developer tool application. The UI should feel:
- professional
- calm
- trustworthy
- clean
- efficient
- suitable for technical users
- suitable for Persian/RTL usage

## Visual Principles

- Consistency over novelty
- Reuse before invention
- Semantic tokens over raw colors
- Clear hierarchy
- Compact but readable interfaces
- Strong form and table usability
- Accessible interactive states
- RTL-first layout awareness

## Tenant-admin Theme

Tenant-admin uses a clean shadcn theme with neutral application surfaces,
indigo actions, restrained cyan data-visualization colors, and independent
semantic operational states. The result should feel precise, calm, and
purpose-built for infrastructure work rather than like a generic SaaS template.

- Use the 4px spacing grid and prefer 8, 12, 16, 24, and 32px intervals.
- Default cards use a 10px radius, a one-pixel semantic border, and a very
  restrained shadow. Elevation must communicate hierarchy, not decoration.
- The sidebar stays visually distinct through its own neutral surface, border,
  active-state accent, and restrained texture in both themes.
- Geometric grid motifs are allowed only on large shell or authentication
  canvases. Keep them low-contrast and never place them behind dense text.
- Indigo identifies primary actions and selection in both themes. Green, amber,
  and red remain reserved for meaningful operational state.
- Hover motion is limited to interactive surfaces and should not exceed a
  one-pixel lift. Static information cards do not float on hover.

## Brand Mark

`web-assets` is the source of truth for the GodarAI mark. Product applications
consume synchronized files under `/brand`; do not copy, recolor, or redraw the
mark in feature code.

| Context | Asset |
| --- | --- |
| Light UI surface | `godarai-mark-light.svg` |
| Dark UI surface | `godarai-mark-dark.svg` |
| Theme-aware product UI | `GodarAILogo` with `variant="auto"` |
| Single-color light surface | `godarai-mark-mono-navy.svg` |
| Single-color dark surface | `godarai-mark-mono-white.svg` |
| Email | `godarai-mark-light-256x188.png` |
| Browser/PWA/social metadata | Files under `/brand/icons`, `/brand/web`, and `/brand/social` |

Always preserve the intrinsic `336:247` ratio. Consumers set a responsive
height and use `w-auto max-w-full object-contain`; they must not force the mark
into a stretched square. Use an empty alt only when adjacent text already names
GodarAI.

## Core Rules for AI Coding Agents

1. Before creating or changing UI, inspect existing `components/ui`, `components/app-ui`, and this design system documentation.
2. Reuse existing product-level components before creating new ones.
3. Do not invent new visual patterns unless there is a clear product need.
4. Prefer semantic Tailwind classes such as `bg-background`, `text-foreground`, `border-border`, `text-muted-foreground`, `bg-card`, `text-card-foreground`, `bg-primary`, `text-primary-foreground`.
5. Avoid raw color utilities like `text-gray-*`, `bg-slate-*`, `border-zinc-*` unless they are already part of a documented component pattern.
6. Avoid arbitrary values like `p-[13px]`, `rounded-[7px]`, `text-[15px]` unless absolutely necessary and documented.
7. Use product-level components for common patterns such as page headers, cards, empty states, forms, tables, badges, and dialogs.
8. Preserve RTL behavior.
9. Preserve accessibility behavior.
10. When adding a reusable visual pattern, document it and add it to the design system playground.

## Layout System

Use full-page containers with `mx-auto`, `w-full`, `max-w-7xl`, and page padding from the standard scale. Page-level vertical rhythm should use `flex flex-col gap-6` or `gap-8`. Route-level pages inside `AdminShell` pass localized page metadata to the shell instead of repeating a `PageHeader` in content. List content should start with its toolbar/filter content, then table/list content, then pagination or state feedback.

Dashboard pages use compact metric grids with `MetricCard`, then wider content sections. Settings and detail pages should use a narrow readable column or `SettingsCard`/`FormSection` groups rather than free-form cards.

## Typography

- Page title: `text-2xl font-bold tracking-tight md:text-3xl`
- Section title: `text-base font-semibold md:text-lg`
- Card title: shadcn `CardTitle`
- Description text: `text-sm leading-6 text-muted-foreground`
- Labels: `text-sm font-medium`
- Helper text: `text-xs leading-5 text-muted-foreground`
- Muted metadata: `text-xs text-muted-foreground`
- Table text: compact `text-sm`; identifiers may use `font-mono`

## Colors

Colors must come from semantic tokens. Use `app/globals.css` and `design-system/tokens.css` as the source of truth. Do not add one-off raw palette classes for normal product UI.

## Spacing

Preferred spacing: `gap-2`, `gap-3`, `gap-4`, `gap-6`, `gap-8`, `p-4`, `p-6`, `px-4`, `py-6`. Avoid one-off spacing values unless a shadcn primitive requires them.

## Radius and Borders

Cards use the shadcn card radius and `border-card-border`. Controls use existing shadcn button/input radius. Borders should use `border-border`, `border-card-border`, or documented tone classes. Focus styles must remain visible through `ring-ring` or the global focus style.

## Components

- `components/ui`: shadcn primitives and low-level UI only.
- `components/app-ui`: product-specific reusable patterns.
- Route/page components: business logic, data fetching, and page-specific composition.

## Forms

Use consistent grouping: title, description, fields, validation, and actions. Prefer `FormSection` for settings and larger forms. Keep labels and helper text aligned in RTL.

## Tables

Tables should have a toolbar/search/filter area, readable columns, row actions, empty state, and pagination area. Preserve existing `components/ui/data-table` behavior; create additional wrappers only after repeated usage is clear.

## States

Loading states should use `Skeleton` or an established loading component. Empty states should use `EmptyState`. Error states should be visible, concise, and actionable. Destructive actions should use `ConfirmDialog`.

## RTL Rules

See `RTL_GUIDELINES.md`.

## Review Checklist

See `UI_REVIEW_CHECKLIST.md`.

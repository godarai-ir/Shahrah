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

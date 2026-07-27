# Design Tokens

## Theme Foundation

Tenant-admin uses the shadcn CSS-variable model with OKLCH values. Light mode
uses white neutral surfaces and indigo actions; dark mode uses warm charcoal
surfaces and a deeper indigo action color. The sidebar has its own neutral
surface and active-state tokens. Feature code must consume semantic tokens
rather than copying theme values directly.

## Color Tokens

- `background`
- `foreground`
- `card`
- `card-foreground`
- `popover`
- `popover-foreground`
- `primary`
- `primary-foreground`
- `secondary`
- `secondary-foreground`
- `muted`
- `muted-foreground`
- `accent`
- `accent-foreground`
- `destructive`
- `destructive-foreground`
- `border`
- `input`
- `ring`

The tenant admin also defines documented product tones: `success`, `warning`, `info`, soft tone tokens, sidebar tokens, and chart tokens in `apps/tenant-admin/design-system/tokens.css`.

The generated `ring` value remains part of the theme contract. Interactive
components map their visible focus treatment to the derived `focus-ring` token
so keyboard focus reaches the project’s non-text contrast target in both modes.

The theme maps indigo to primary actions and selection, neutral OKLCH values to
secondary, accent, sidebar, and code surfaces, and a restrained cyan scale to
charts. Success, warning, and destructive tokens stay semantically independent
from the product palette.

Code display surfaces use:

- `code`
- `code-foreground`
- `code-border`
- `code-muted-foreground`

## Tailwind Usage

Use:

```tsx
<div className="bg-background text-foreground" />
```

Avoid:

```tsx
<div className="bg-white text-gray-900" />
```

## Spacing Tokens

Prefer Tailwind spacing values already used by the system: `gap-2`, `gap-3`, `gap-4`, `gap-6`, `gap-8`, `p-4`, `p-6`, `px-4`, and `py-6`.

## Radius Tokens

Use shadcn radius classes and the CSS variables in `tokens.css`: `--radius-sm`, `--radius-md`, `--radius-lg`, and `--radius-xl`.

## Shadow / Elevation

Use shadows sparingly. Default cards already include `shadow-sm`; higher elevation should be reserved for overlays or intentionally elevated surfaces.

## Code Surfaces

Use `bg-code`, `text-code-foreground`, `border-code-border`, and `text-code-muted-foreground` for code blocks and line numbers. Do not use raw slate palettes for new code surfaces.

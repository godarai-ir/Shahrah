# Design Tokens

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

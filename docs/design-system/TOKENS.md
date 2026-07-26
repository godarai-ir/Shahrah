# Design Tokens

## GodarAI Brand Primitives

These primitives come from `web-assets/colors.json` and are exposed by the
tenant-admin token sheet. Feature code must consume semantic tokens rather than
using these hex values directly.

| Primitive | Value | Intended role |
| --- | --- | --- |
| `brand-cyan` | `#00C9EB` | Dark-mode actions and branded selection |
| `brand-cyan-bright` | `#05D1EE` | Cyan hover |
| `brand-cyan-deep` | `#00BCE1` | Focus borders and derived rings |
| `brand-navy` | `#012651` | Light-mode actions and navigation |
| `brand-navy-top` | `#022A57` | Navy hover |
| `brand-navy-bottom` | `#011D40` | Dark elevated surfaces |
| `brand-dark-background` | `#06182D` | Dark canvas and sidebar |
| `brand-off-white` | `#F6FBFF` | Light canvas and dark-mode text |
| `brand-ice` | `#D9EEFF` | Soft surfaces and secondary dark text |

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

GodarAI Atlas maps navy to light-mode primary actions, cyan to dark-mode
primary actions and focus, and brand-derived `color-mix()` values to secondary,
accent, sidebar, code, and chart surfaces. Success, warning, and destructive
tokens stay semantically independent from the brand palette.

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

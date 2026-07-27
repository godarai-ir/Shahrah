# GodarAI Tenant Admin Design System Guide

This guide is the implementation contract for GodarAI tenant-admin UI work in this repository. It is written for coding agents and frontend engineers who need to create consistent pages, components, dashboards, and configuration workflows without re-deciding the product direction each time.

## Repository Baseline

| Area | Current state |
| --- | --- |
| Framework | Next.js `16.2.9` App Router with React `19.2.4` and TypeScript. |
| Routing | Root `app/` directory with Next App Router pages for login, dashboard, models, providers, traces, billing, settings, and gateway policy workflows. |
| Styling | Tailwind CSS v4 through `@tailwindcss/postcss`; global CSS imported from `app/globals.css`. |
| shadcn/ui | `components.json` uses style `new-york`, `rsc: true`, `tsx: true`, `baseColor: slate`, CSS variables, lucide icons, and `rtl: true`. |
| Components | Multiple shadcn primitives exist under `components/ui`, including button, card, input, label, alert, badge, table, tabs, sidebar, dialog, sheet, select, popover, tooltip, and chart. |
| Utilities | `lib/utils.ts` exports `cn()` with `clsx` and `tailwind-merge`. |
| Tokens | `app/globals.css` imports `design-system/tokens.css`, which defines shadcn semantic tokens for neutral core surfaces, indigo primary actions, cyan chart colors, success, warning, info, destructive, borders, inputs, sidebar colors, radius, shadows, and typography. |
| Dark mode | Token values exist under `.dark`, Tailwind has `@custom-variant dark (&:is(.dark *));`, and `ThemeProvider` plus `ThemeToggle` are available. |
| Typography | `app/layout.tsx` loads `Vazirmatn` and `JetBrains_Mono` through `next/font/google`; `--font-sans` maps to Vazirmatn and `--font-mono` maps to JetBrains Mono. |
| RTL | shadcn config enables RTL, the app root renders `dir="rtl"`, and `lang` comes from `next-intl`. |
| Page patterns | Product pages now use the tenant-admin shell, operational dashboard pages, and form/list/detail workflows; new UI should extend those patterns rather than the original starter. |
| Naming | Current conventions are root-level `app`, `components`, `components/ui`, `lib`, and `@/*` imports. |

Next.js 16 note: follow the local docs in `node_modules/next/dist/docs/` before changing routing, layouts, fonts, CSS, or Server/Client Component boundaries. Pages and layouts are Server Components by default; use `"use client"` only for interactive components that need state, event handlers, effects, or browser APIs.

## 1. Product Design Principles

| Principle | Practical rule |
| --- | --- |
| Infrastructure-first | Prioritize operational clarity over marketing polish. The first screen of app routes should show system state, not promotional content. |
| Developer-friendly | Make model IDs, endpoints, API keys, traces, JSON, and errors easy to scan, copy, filter, and inspect. |
| Clarity over decoration | Use neutral surfaces, strong hierarchy, and purposeful contrast. Avoid decoration that competes with data. |
| Trust and operational confidence | Dangerous actions, degraded providers, billing risks, and missing credentials must be visible and specific. |
| Data-dense but readable | Favor compact tables, metrics, badges, and split panels, but maintain consistent spacing and line length. |
| Fast scanning | Put status, owner, provider, model, cost, latency, and last activity in predictable positions across pages. |
| Consistent interactions | Reuse the same table, filter, dialog, sheet, toast, and copy patterns across domains. |
| RTL-first and bilingual-ready | Treat Persian and English as first-class. Use direction-aware layout utilities and preserve LTR technical strings. |
| Accessibility by default | Semantic HTML, keyboard flows, visible focus, labels, and contrast are required, not cleanup tasks. |

## 2. Visual Identity

GodarAI tenant-admin uses Vazirmatn typography, neutral application surfaces, crisp data cards, indigo primary actions and selection in both themes, a restrained cyan chart scale, and independent semantic operational states. Tenant-admin should feel like a focused control plane for teams running LLM-powered products—precise, reliable, modern, and enterprise-ready, with enough density to compare providers, models, traces, cost, and limits quickly.

Use restrained 10px cards, thin semantic borders, compact controls, meaningful icons, readable tables, and status colors that serve operations. The sidebar remains distinct through a neutral surface, border, active indigo marker, and quiet texture; the main canvas stays intentionally quieter. Green, amber, and red communicate operational meaning. Low-contrast geometric grids may appear on the shell and authentication canvases, but never compete with dense content.

Avoid overly playful illustrations, excessive gradients, heavy animation, consumer landing-page layouts inside authenticated app screens, low-contrast text, oversized cards for dense operational content, and decorative elements that reduce readability. Login/onboarding screens may carry more brand atmosphere than authenticated pages, but they should still show trust, security, and gateway context rather than generic marketing art.

## 3. Design Tokens

Use the shadcn CSS-variable token model already defined in `app/globals.css`. Prefer Tailwind classes that resolve to tokens, such as `bg-background`, `text-foreground`, `border-border`, `bg-card`, `text-muted-foreground`, `bg-primary`, and `ring-ring`.

| Token | Use |
| --- | --- |
| `background` / `foreground` | Page-level base surface and default text. |
| `card` / `card-foreground` | Cards, panels, table containers, and grouped sections. |
| `popover` / `popover-foreground` | Menus, command palettes, popovers, selects, tooltips where applicable. |
| `primary` / `primary-foreground` | Main action, active navigation, selected high-emphasis state. |
| `secondary` / `secondary-foreground` | Secondary buttons, low-emphasis selected states. |
| `muted` / `muted-foreground` | Subtle panels, helper text, metadata, skeleton backgrounds. |
| `accent` / `accent-foreground` | Hover rows, focused navigation items, lightweight highlights. |
| `destructive` | Delete, revoke, expired, failed, blocked, or irreversible actions. |
| `border` | Dividers, card outlines, table grids, control borders. |
| `input` | Inputs, selects, textareas, searchable controls. |
| `ring` | Focus indicators and keyboard-visible outlines. |
| `chart-1` to `chart-5` | Analytics and breakdown charts. Current palette is narrow; expand later for provider/model charts. |
| `sidebar-*` | App sidebar surfaces, active state, borders, and focus ring. |
| `radius` | Base radius is `0.625rem`; keep app cards at `rounded-lg` or below unless a shadcn primitive requires more. |

Semantic tokens already available for status-heavy UI:

| Token | Purpose |
| --- | --- |
| `success` / `success-foreground` | Connected, healthy, paid, completed, active. |
| `warning` / `warning-foreground` | Low balance, nearing limit, degraded, pending rotation. |
| `info` / `info-foreground` | Informational notices, neutral progress, documentation hints. |
| `surface-elevated` / `surface-demo` | Branded elevated panels and dark console/demo surfaces inspired by `apps/website`. |
| `neutral-soft` | Large dashboard bands or filter bars distinct from `muted`. |
| `code` / `code-foreground` | Add later if JSON viewers, API snippets, request IDs, and endpoints need a dedicated surface beyond `font-mono` plus existing tokens. |

Do not hardcode one-off colors in application UI. Current starter code includes hardcoded `zinc`, `black`, and hex hover colors; future Shahrah pages should replace that style with tokens.

```tsx
// Good: semantic tokens and logical spacing
<section className="rounded-lg border border-border bg-card p-4 text-card-foreground">
  <p className="text-sm text-muted-foreground">Provider latency</p>
</section>

// Avoid: visual color decisions embedded in the component
<section className="rounded-xl border border-zinc-200 bg-white p-[19px] text-zinc-950">
  ...
</section>
```

Z-index should be standardized when overlays are introduced:

| Layer | Suggested range |
| --- | --- |
| Base content | `z-0` |
| Sticky table headers / page header | `z-10` |
| Sidebar / top nav | `z-20` |
| Dropdowns / popovers | `z-40` |
| Dialogs / sheets | `z-50` |
| Toasts / global command menu | `z-60` if configured as a project utility |

### 3.1 Color Accessibility (WCAG 2.2)

Tenant-admin UI must meet [WCAG 2.2 Level AA](https://www.w3.org/TR/WCAG22/) in both light and dark themes. Treat contrast as a property of a complete semantic token pair in its rendered context, not of an individual color. A token change is incomplete until its foreground/background, borders, focus state, hover state, and disabled state have been checked on every surface where the token is used.

#### Contrast requirements

| Content or indicator | Minimum contrast | Application rule |
| --- | --- | --- |
| Normal text and images of text | `4.5:1` against the effective background | Applies to body text, labels, helper text, placeholders, badge text, table content, and text inside controls. |
| Large text | `3:1` against the effective background | Large means at least `18pt` (`24px`) regular or `14pt` (about `18.67px`) bold. Do not treat `text-lg` as large text unless its computed size and weight qualify. |
| Essential control boundaries, states, and meaningful graphics | `3:1` against adjacent colors | Applies to input borders needed to locate a field, focus indicators, selected states, checkmarks, icons without a visible text equivalent, and meaningful chart elements. See [WCAG 1.4.11 Non-text Contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html). |
| Enhanced contrast target | `7:1` for normal text and `4.5:1` for large text | Prefer this WCAG AAA target for critical error, security, billing, and destructive-action copy when it can be achieved without weakening hierarchy. |

Contrast ratios are thresholds and must not be rounded up: `4.49:1` does not satisfy `4.5:1`. Measure the computed foreground against the final composited background. Re-test text or controls placed on translucent colors, gradients, images, hover fills, selected rows, sticky regions, overlays, and nested surfaces because the effective contrast may differ from the base token value.

#### Semantic color-pair contract

Use the foreground token designed for the surface; do not assume that `foreground` or white is readable on every semantic color.

| Surface | Required foreground | Typical use |
| --- | --- | --- |
| `background` | `foreground` | Page content. |
| `card` | `card-foreground` | Cards, tables, grouped panels. |
| `popover` | `popover-foreground` | Menus, selects, popovers. |
| `primary` | `primary-foreground` | Primary actions and high-emphasis selections. |
| `secondary` | `secondary-foreground` | Secondary actions and selections. |
| `muted` | `muted-foreground` | Helper text and low-emphasis metadata; never use it for critical instructions or errors. |
| `accent` | `accent-foreground` | Hovered and highlighted items. |
| `success`, `warning`, `info`, `destructive` | Matching `*-foreground` | Solid semantic badges, alerts, and controls. |
| `*-soft` | Matching `*-soft-foreground` and `*-soft-border` | Low-emphasis statuses and notices. |
| `surface-demo`, `code` | Matching `*-foreground` | Console, code, and technical preview surfaces. |

Do not reduce text opacity or apply Tailwind opacity modifiers to a passing pair without measuring the composited result. When a new color role is needed, add a semantic surface/foreground/border set to `design-system/tokens.css`, provide light and dark values, document its intended use, and test the pair before using it in a component. Do not fix a contrast failure locally with a raw color utility.

#### Color must not carry meaning alone

[WCAG 1.4.1 Use of Color](https://www.w3.org/WAI/WCAG22/Understanding/use-of-color) requires a visible cue in addition to color wherever color communicates meaning or action.

- Statuses must include a persistent text label such as “فعال”, “در انتظار”, or “ناموفق”; an icon or shape may reinforce the label.
- Form errors must include an error message and invalid-state semantics, not only a red border. Required fields need a text or symbol cue explained to users.
- Links inside prose must be underlined or otherwise visibly distinguishable without relying only on hue. Preserve a non-color cue on hover and focus.
- Selected rows, active navigation, toggles, and comparison results need a checkmark, border, icon, weight, or text change in addition to a fill-color change.
- Charts must have direct labels or an accessible legend and distinguish series with shapes, dash patterns, markers, or labels as well as `chart-*` colors. Never use only green/red to encode success/failure or positive/negative values.

Color alternatives must remain visible to sighted users with color-vision deficiency; an `aria-label` alone does not satisfy the color-only requirement. Programmatic names and states are still required for assistive technology.

#### Interaction and focus states

- Preserve the shared `ring-ring` focus indicator. It must remain visible on `background`, `card`, `popover`, sidebar, solid semantic fills, and dark demo surfaces. Aim for a focus indicator equivalent to a `2px` perimeter with at least a `3:1` change from the unfocused pixels, following the WCAG 2.2 [Focus Appearance](https://www.w3.org/WAI/WCAG22/Understanding/focus-appearance.html) enhanced target.
- Hover, active, selected, validation, and focus states must remain identifiable and keep their text and essential graphics above the applicable contrast threshold. Do not make hover the only state that reveals an action.
- Disabled controls are exempt from the WCAG minimum contrast thresholds, but they must still be recognizable. Communicate disabled state through native semantics and more than a subtle color change; do not make important explanatory text disabled or unreadable.

#### Verification before merge

For every new or changed color token or component:

1. Check every documented foreground/background pair in light and dark themes with a contrast analyzer; record exact ratios for token changes in the pull request.
2. Exercise default, hover, active, focus-visible, selected, invalid, loading, and disabled states on each surface used by the component.
3. Verify the page in grayscale and with common red-green and blue-yellow color-vision simulations; all status, selection, chart, and validation meanings must remain understandable.
4. Run the available automated accessibility checks, then manually inspect text over translucent or variable backgrounds and keyboard focus. Automated checks do not prove that color is not the only cue.
5. Test representative Persian and English strings because font weight, size, and truncation can change whether text qualifies as large and whether its non-color cue remains visible.

## 4. Typography

Current fonts are loaded through `next/font/google` in `app/layout.tsx`. Keep using `next/font` for app-wide fonts because it self-hosts optimized font assets and avoids layout shift. Do not add external font files or network CSS unless explicitly approved.

| Text type | Rule |
| --- | --- |
| Product UI | Use `font-sans`; currently Vazirmatn via `--font-sans`, matching the website and Persian-first product tone. |
| Persian/Farsi | Vazirmatn is installed through `next/font/google` with Arabic and Latin subsets. Test digits, diacritics, line-height, and mixed Persian/English strings. |
| English | Vazirmatn is acceptable for bilingual product UI; technical Latin strings should use `font-mono` where scanability matters. |
| Code and IDs | Use `font-mono`; current token maps to JetBrains Mono. |
| Headings | Keep compact: page titles `text-2xl font-semibold`, section titles `text-base` to `text-lg font-semibold`, card titles `text-sm` to `text-base font-medium`. |
| Body | Default `text-sm` for app content; `text-base` only for dense explanatory pages or major empty states. |
| Tables | `text-sm`, compact row height, `font-mono` for IDs, endpoints, model names, and request IDs. |
| Metrics | Use `text-2xl font-semibold` for primary dashboard values; pair with `text-xs` or `text-sm text-muted-foreground` labels. |
| Helper text | `text-xs` or `text-sm text-muted-foreground`; keep it short and actionable. |
| Errors | `text-sm text-destructive`; include the field or resource name when possible. |

For Persian text, avoid tight line-height. Use `leading-6` for body copy and `leading-7` where strings are longer. For mixed Persian and technical English, wrap code-like values in LTR spans.

```tsx
<span dir="ltr" className="inline-block font-mono text-xs">
  gpt-4.1-mini
</span>
```

## 5. Layout System

Shahrah is primarily a desktop infrastructure console. Build app pages with a stable shell, predictable page headers, dense content sections, and responsive fallbacks.

| Layout element | Standard |
| --- | --- |
| App shell | Sidebar plus main content. Keep navigation persistent on desktop. Collapse to sheet/drawer on mobile. |
| Sidebar | Use `bg-sidebar`, `text-sidebar-foreground`, `border-sidebar-border`, and logical border placement. |
| Header | Use for breadcrumbs, search, environment/tenant switcher, alerts, and account menu. Keep it compact. |
| Main content | `flex-1`, `min-w-0`, and a constrained inner width for forms; full-width for data tables and charts. |
| Page container | `px-4 py-4 sm:px-6 lg:px-8`; use `gap-4` or `gap-6` between major sections. |
| Page header | Title, short description, metadata, and actions in a responsive flex layout. Actions align to inline end. |
| Breadcrumbs | Present on nested pages and details. Mirror separator icons in RTL if needed. |
| Cards | Use for individual repeated items or framed tools. Do not nest cards inside cards. |
| Tables | Preferred for models, traces, providers, keys, billing events, users, and policies. |
| Detail pages | Header summary, metadata strip, tabs, activity, configuration, and danger zone. |
| Settings pages | Sectioned forms with sticky or repeated save/cancel actions for long pages. |
| Empty states | Small, practical, and action-oriented. Avoid large illustrations inside the app. |
| Loading states | Skeletons matching final layout; avoid spinners for whole pages when skeletons are possible. |

Recommended spacing:

| Context | Spacing |
| --- | --- |
| Page padding | `px-4 py-4 sm:px-6 lg:px-8` |
| Major vertical sections | `gap-6` |
| Card padding | `p-4` for dense cards, `p-6` for forms/details |
| Table cells | `px-3 py-2` or `px-4 py-3` depending density |
| Form fields | `space-y-2` inside a field, `gap-4` between fields |
| Dialog body | `p-6`, footer with `gap-2` and logical alignment |

```tsx
export default function ModelsPage() {
  return (
    <main className="flex-1 px-4 py-4 sm:px-6 lg:px-8">
      <div className="flex flex-col gap-6">
        <header className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <h1 className="text-2xl font-semibold tracking-normal">Models</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Compare providers, pricing, latency, and capabilities.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2 sm:justify-end">...</div>
        </header>
      </div>
    </main>
  )
}
```

## 6. RTL and Bilingual Design Rules

RTL support is mandatory for Shahrah. Build direction-aware UI from the start.

Strict rules:

| Topic | Rule |
| --- | --- |
| Root direction | Set `dir` according to locale. Persian routes should render `dir="rtl"` and `lang="fa-IR"` or `fa`. English should render `dir="ltr"`. |
| Locale routing | When localization is added, prefer `app/[lang]/layout.tsx` and typed `LayoutProps<'/[lang]'>` as described in local Next docs. |
| Spacing | Prefer logical Tailwind utilities: `ps-*`, `pe-*`, `ms-*`, `me-*`, `start-*`, `end-*`, `text-start`, `text-end`. |
| Avoid | Do not use `left-*`, `right-*`, `ml-*`, `mr-*`, `pl-*`, `pr-*`, `text-left`, or `text-right` unless the behavior is intentionally physical. |
| Icons | Directional icons must mirror in RTL: chevrons, arrows, panel open/close, breadcrumb separators, send/forward/back. |
| Sidebar | Sidebar position should follow product direction unless a route explicitly uses a fixed operational layout. Use logical borders. |
| Forms | Labels, helper text, errors, and action rows align to text start/end. |
| Tables | Header and cells align by data type: text to start, numbers often to end, technical identifiers LTR. |
| Technical strings | API keys, URLs, JSON, request IDs, model IDs, and code remain LTR even inside Persian UI. |
| Mixed text | Wrap code-like substrings in `dir="ltr"` and `font-mono`; use `unicode-bidi: isolate` via class if needed. |
| Search | Inputs must accept Persian and English. Placeholder copy should localize without breaking layout. |
| Truncation | Test long Persian organization names and long English IDs. Use `min-w-0`, `truncate`, and tooltips/copy actions. |

Correct:

```tsx
<div dir={locale === "fa" ? "rtl" : "ltr"} className="flex items-center gap-2">
  <label className="text-start text-sm font-medium">کلید API</label>
  <span dir="ltr" className="ms-auto truncate font-mono text-xs">
    sk_live_••••9Gf2
  </span>
</div>
```

Incorrect:

```tsx
<div className="flex items-center gap-2 text-left">
  <label className="mr-2">کلید API</label>
  <span className="ml-auto">sk_live_••••9Gf2</span>
</div>
```

Icon mirroring:

```tsx
<ChevronRight className="size-4 rtl:rotate-180" aria-hidden="true" />
```

If `rtl:` utilities are not available in the current Tailwind setup, add a small project utility later rather than scattering conditional class strings.

## 7. Component Standards

Use shadcn/ui conventions before creating custom primitives. Current installed primitive is `Button`; add more shadcn components as the app needs them.

| Component | Use and behavior | RTL and accessibility |
| --- | --- | --- |
| Buttons | Use for explicit commands. Variants should match current `Button`: `default`, `outline`, `secondary`, `ghost`, `destructive`, `link`; sizes include text and icon sizes. | Text starts with icon on logical inline start. Use `aria-label` for icon-only buttons. |
| Inputs | Use for single-line values, search, filters, API names, limits. Include labels and helper/error text. | Use `text-start`; technical inputs may use `dir="ltr"` and `font-mono`. |
| Textareas | Use for JSON, prompts, descriptions, policy notes. Use monospace for JSON or request bodies. | Preserve LTR for JSON/code even in RTL forms. |
| Selects | Use for small bounded option sets like provider, region, status, role. | Menu alignment should follow trigger inline start. |
| Comboboxes | Use for searchable model/provider/team selectors. | Search must match Persian and English labels. |
| Checkboxes | Use for independent binary options in forms and table selection. | Label spacing uses logical gap, not physical margins. |
| Switches | Use for immediate on/off settings only when change can be safely saved. Otherwise use checkbox plus Save. | Announce checked state; label at text start. |
| Radio groups | Use for mutually exclusive modes like routing strategy or billing owner. | Preserve keyboard arrow behavior. |
| Tabs | Use for detail pages: Overview, Configuration, Usage, Traces, Activity, Danger. | Tab order follows reading direction. |
| Cards | Use for repeated summaries, provider cards, model cards, metric cards. | Keep card actions at inline end and content text-start. |
| Dialogs | Use for confirmations and short create/edit flows. | Focus trap, labelled title, clear cancel action. |
| Sheets | Use for side panels, filters, trace metadata, create/edit flows that benefit from context. | Open from inline end for normal details; mirror in RTL. |
| Dropdown menus | Use for row actions and secondary command groups. | Align to inline end for action menus. |
| Tooltips | Use for short explanations and icon labels. | Never hide critical information only in a tooltip. |
| Badges | Use for status, capabilities, regions, modalities, roles. | Keep labels short and localizable. |
| Alerts | Use for page-level or section-level status: missing credentials, low balance, degraded provider. | Include icon plus text; do not rely on color only. |
| Toasts | Use for completion feedback, copy success, async failures. | Messages should be specific and screen-reader friendly. |
| Tables | Use for operational lists. Include sorting, filtering, pagination, loading, empty, and error states. | Text start alignment; numeric and cost columns can align end. |
| Data grids | Use only when tables need column resizing, pinning, or complex bulk operations. | Keyboard navigation must be tested. |
| Pagination | Use for server-backed lists. Include result count when possible. | Previous/next icons mirror in RTL. |
| Filters | Use compact filter bars and sheets for advanced filters. | Chips should remove with accessible close buttons. |
| Search bars | Use for model, provider, key, trace, user, and tenant search. | Support Persian and English input. |
| Command menus | Use for global navigation and quick actions. | Search aliases should include localized names. |
| Empty states | Explain what is absent and the next action. | No oversized decorative content. |
| Skeletons | Match the final layout dimensions. | Keep motion subtle and respect reduced motion. |
| Error states | State what failed, why if known, and what the user can do. | Use semantic alert regions for blocking errors. |

## 8. Domain-Specific UI Patterns for LLM Gateway

### Model Cards

Model cards should be compact comparison units. Use them when users browse model catalog entries; use tables when they compare many models.

Required fields: model name, provider, capabilities, context window, input/output pricing, supported modalities, latency indicator, availability, region, tags, and recommended use cases.

Pattern:

```tsx
<article className="rounded-lg border border-border bg-card p-4">
  <div className="flex items-start justify-between gap-3">
    <div className="min-w-0">
      <h3 dir="ltr" className="truncate font-mono text-sm font-semibold">
        openai/gpt-4.1-mini
      </h3>
      <p className="mt-1 text-sm text-muted-foreground">OpenAI · Text, vision</p>
    </div>
    <span className="rounded-md bg-success/10 px-2 py-1 text-xs text-success">
      Available
    </span>
  </div>
</article>
```

### Provider Cards

Provider cards show provider name, status, configured accounts, supported models, credential status, and BYOK/shared-key status. Put credential health in a visible badge. If a provider is misconfigured, include a direct configuration action.

### Request Trace Views

Trace detail pages should support a timeline, metadata, provider/model used, latency, token usage, cost, errors, raw request/response inspection, collapsible technical sections, and copyable IDs. Use split panels on desktop: timeline and summary on one side, raw request/response tabs on the other.

### Usage Analytics

Analytics pages should include metric cards, time-series charts, provider breakdown, model breakdown, tenant/team breakdown, cost and token usage, latency percentiles, and error rates. Use `chart-*` tokens initially, but add broader chart tokens before showing many providers at once.

### API Key Management

API key pages should show masked keys, copy actions, rotation, expiration, scopes, last used date, and security warnings. Never display full secrets after creation. Use confirmation dialogs for revoke, rotate, and scope changes.

### Rate Limit and Policy UI

Policy pages should support rule lists, priority ordering, scope selection, condition builders, quota display, and warnings before destructive changes. Use a readable rule summary so users do not need to parse JSON for common policies.

### Tenant and Organization UI

Organization pages should support members, roles, teams, invitations, access control, usage limits, and billing ownership. Put role changes, transfer ownership, and user removal behind explicit confirmations.

## 9. Data Display Guidelines

| Data type | Display rule |
| --- | --- |
| Tables | Scannable, compact, sortable where useful, filterable for operational pages, with persistent column meanings. |
| Metrics | Include value, label, comparison period, and trend if available. Avoid unlabelled large numbers. |
| Charts | Use clear legends, accessible colors, and visible units. Do not rely on hover-only details. |
| Logs | Use monospace, timestamps, severity, source, and expandable details. |
| JSON viewers | Use monospace, collapsible sections, copy action, and line wrapping controls. |
| Copyable values | Add copy buttons for API keys, IDs, URLs, endpoints, trace IDs, and model IDs. |
| IDs | Truncate middle or end intelligently, preserve full value in copy. |
| URLs/endpoints | Use `dir="ltr"`, `font-mono`, and truncation. |
| Timestamps | Show localized date/time plus tooltip or detail for exact ISO when needed. |
| Currency | Label currency explicitly. For Iran market, be precise about rial/toman if billing uses either. |
| Token counts | Use compact formatting for large values, but exact values in details. |
| Percentages | Include denominator or context for error rates and cache hit rates. |
| Latency | Prefer `ms`; show percentiles as `p50`, `p95`, `p99`. |
| Status | Use badge plus text; do not rely on color alone. |

Long technical strings should use this pattern:

```tsx
<div className="flex min-w-0 items-center gap-2">
  <code dir="ltr" className="min-w-0 truncate font-mono text-xs">
    req_01JZ8A6W4EH9R6Q4R9C7M2
  </code>
  <Button variant="ghost" size="icon-xs" aria-label="Copy request ID">...</Button>
</div>
```

## 10. Status, Feedback, and System States

Use status language consistently. Add missing semantic tokens before building many status-heavy screens.

| State | Visual treatment | Shahrah examples |
| --- | --- | --- |
| Success | Calm green token when added | Provider connected, request completed, key rotated. |
| Warning | Amber/yellow token when added | Billing balance low, near rate limit, provider degraded. |
| Error | `destructive` token | Request failed, API key expired, invalid credential. |
| Info | Blue/cyan token when added | New model available, usage sync delayed. |
| Loading | Skeletons matching content | Loading traces, models, analytics. |
| Empty | Muted panel with one action | No API keys, no traces for selected period. |
| Disabled | Lower opacity and no pointer events | Action unavailable due to role or missing config. |
| Pending | Neutral badge or warning if delayed | Invitation pending, rotation pending. |
| Processing | Inline progress and disabled action | Policy applying, export generating. |
| Degraded | Warning badge and explanation | Provider latency elevated. |
| Offline | Destructive or warning depending severity | Provider unreachable, ClickHouse offline. |
| Connected | Success badge | Provider credential valid. |
| Misconfigured | Warning/destructive with CTA | Provider credential missing, ClickHouse not configured. |
| Unauthorized | Destructive text and recovery | Tenant user inactive, role lacks permission. |

Error microcopy examples:

| Weak | Better |
| --- | --- |
| `Failed` | `Request trace could not be loaded. Retry or check ClickHouse configuration.` |
| `Invalid input` | `Context window must be between 1,000 and 1,000,000 tokens.` |
| `Something went wrong` | `Provider credential is missing. Add an API key before enabling routing.` |

## 11. Navigation Information Architecture

Preferred primary navigation:

| Section | Purpose | Suggested icon |
| --- | --- | --- |
| Dashboard | Operational summary, alerts, recent activity | `LayoutDashboard` |
| Models | Model catalog, capabilities, pricing | `Bot` or `Brain` |
| Providers | Provider accounts and health | `Cloud` or `Server` |
| Gateway | Routes, endpoints, routing strategy | `Network` |
| Traces | Request logs and trace details | `ListTree` or `Activity` |
| Usage | Analytics, tokens, cost, latency | `ChartNoAxesCombined` |
| API Keys | Keys, scopes, rotation | `KeyRound` |
| Policies | Rate limits, routing policies, quotas | `ShieldCheck` |
| Organizations | Tenants, members, roles | `Building2` |
| Teams | Team membership and access | `Users` |
| Billing | Wallet, packages, invoices | `WalletCards` |
| Settings | Environment and account settings | `Settings` |

Sidebar behavior: show icon plus label on desktop, support collapsed icon-only state later, and use active states based on route segment. Use breadcrumbs for detail pages such as a model, provider account, trace, policy, tenant, or invoice. On mobile, use a sheet opened by an icon button with a clear label.

In RTL, place navigation according to direction and mirror directional icons. Page titles should use resource nouns: `Models`, `Provider accounts`, `Trace req_...`, `API keys`, `Rate limits`.

## 12. Forms and Configuration UX

Configuration changes can affect production infrastructure. Design forms for clarity and safety.

| Topic | Rule |
| --- | --- |
| Layout | Use single-column forms for sensitive configuration; use two columns only for short independent fields. |
| Grouping | Group by operational meaning: Credentials, Routing, Limits, Observability, Billing. |
| Required fields | Mark required fields clearly. Do not rely only on `*` if localized copy can be clearer. |
| Validation | Validate inline and at submit. Show specific range, format, or permission requirements. |
| Helper text | Explain operational consequences, not obvious UI behavior. |
| Advanced settings | Hide under disclosure only when defaults are safe. |
| Secrets | Mask existing values. Show newly created secrets only once. Provide copy and rotate actions. |
| Confirmation | Require confirmation for revoke, delete, rotate, disable provider, transfer ownership, and billing changes. |
| Save/cancel | Use explicit Save/Cancel for multi-field forms. Inline toggles only for safe immediate changes. |
| Dirty state | Warn before navigation if unsaved infrastructure changes exist. |
| Optimistic updates | Use only when rollback is clear and low risk. |
| Read-only | Explain why a field is read-only: role, inherited policy, billing state, or system-managed value. |

## 13. Interaction Patterns

| Interaction | Standard |
| --- | --- |
| Hover | Subtle `bg-muted` or `bg-accent`; never shift layout. |
| Focus | Visible `ring-ring` focus states. Do not remove outlines without replacement. |
| Keyboard | All controls, menus, dialogs, tabs, tables, and command menus must be usable by keyboard. |
| Copy | Copyable technical values get an icon button, tooltip/label, and success toast. |
| Inline editing | Use for low-risk labels/descriptions; use forms/dialogs for credentials, limits, billing, roles. |
| Row actions | Use dropdowns for secondary actions and visible buttons for primary row action where space allows. |
| Bulk actions | Require selected count and confirmation for destructive operations. |
| Filtering | Server-backed for large trace/model/key lists; show active filters as removable chips. |
| Sorting | Use clear active sort state; preserve sort in URL for list pages. |
| Search | Debounce client input; support Persian and English. |
| Refresh | Provide manual refresh on operational lists; show last updated time. |
| Real-time updates | Avoid disruptive layout jumps. Highlight new/degraded rows subtly. |
| Confirmation | State resource name, consequence, and recovery path if any. |
| Toasts | Use for short feedback. Blocking or critical issues belong in alerts, not only toasts. |
| Dialogs | Use for focused decisions. |
| Sheets | Use for contextual side tasks, filters, and trace inspection. |

## 14. Accessibility Guidelines

Use semantic HTML first. Buttons are buttons; links are links; headings follow hierarchy; tables use table elements when representing tabular data.

Requirements:

| Area | Requirement |
| --- | --- |
| Keyboard | No mouse-only workflow. Test tab order in dialogs, menus, forms, and data tables. |
| Focus | Use visible focus rings from tokens. Focus should move into dialogs/sheets and return to trigger. |
| ARIA | Add ARIA only when semantic HTML is insufficient. Every icon-only button needs an accessible name. |
| Contrast | Meet WCAG AA for text and controls in light and dark modes. |
| Forms | Labels are always associated. Errors identify the field and are announced where appropriate. |
| Tables | Header cells use `scope`; row actions have resource-specific labels. |
| Dialogs | Must have title, description if needed, focus trap, escape handling, and close/cancel action. |
| Alerts | Critical alerts should use semantic roles and visible text. |
| RTL | Screen reader labels should be localized; LTR technical strings should preserve reading order. |
| Motion | Respect reduced motion preferences. |

## 15. Responsive Design

Shahrah is desktop-first but must not break on tablets or phones.

| Viewport | Behavior |
| --- | --- |
| Desktop | Persistent sidebar, dense tables, split detail views, visible filter bars. |
| Laptop | Keep sidebar and tables, reduce padding if needed. |
| Tablet | Sidebar may collapse; detail views stack; filters can move into a sheet. |
| Mobile | Use sheet navigation, stacked cards, horizontal table scroll only when card layout is worse. |

Responsive rules:

| Pattern | Rule |
| --- | --- |
| Sidebar | Persistent on large screens, collapsible or sheet on small screens. |
| Tables | Keep important columns visible; hide secondary metadata behind row details on small screens. |
| Cards | Stack with `grid-cols-1`, then `sm:grid-cols-2`, `xl:grid-cols-3` when useful. |
| Filters | Compact filters in a sheet on mobile. Active chips remain visible. |
| Dialogs | Use full-width or near-full-width on mobile; avoid overflow beyond viewport. |
| Charts | Resize with parent; keep legends readable and avoid hover-only access. |
| Long text | Use `min-w-0`, `truncate`, wrapping, or details panels. Persian and English must both be tested. |

## 16. Dark Mode

Dark tokens already exist in `app/globals.css` under `.dark`, but there is no theme provider or user toggle. Until a provider is added, implement UI with token classes so dark mode works when `.dark` is applied.

Rules:

| Topic | Rule |
| --- | --- |
| Token usage | Use semantic token utilities, not hardcoded light/dark colors. |
| Contrast | Verify text, borders, badges, charts, and disabled states in both modes. |
| Charts | Current chart tokens include indigo, green, purple, amber, and red. Use them for provider/model analytics before adding one-off chart colors. |
| Shadows | Prefer borders in dark mode; shadows can disappear on dark surfaces. |
| Status | `success`, `warning`, and `info` tokens exist in light and dark modes; use them through shared Badge/Alert variants instead of raw Tailwind colors. |
| Images/logos | Avoid relying on `dark:invert` for product assets unless visually verified. |

## 17. Iconography

Use `lucide-react` as the default icon set. Do not mix unrelated icon styles.

| Icon use | Standard |
| --- | --- |
| Navigation | `size-4` or `size-5`, consistent stroke, paired with label. |
| Buttons | `size-4` by default; shadcn `Button` already normalizes SVG size. |
| Icon-only buttons | Use `size="icon"` variants and `aria-label`. |
| Status | Pair icon with text for errors, warnings, and success states. |
| Placement | Icon before label at inline start; trailing icons only for dropdowns or next actions. |
| Direction-aware | Mirror arrows, chevrons, panel controls, external flow icons, and breadcrumbs in RTL. |

Avoid icons as decoration when text would be clearer. Use tooltips for unfamiliar icon-only actions.

## 18. Motion and Animation

Motion should be subtle and functional. The repo includes `tw-animate-css`, so use it through shadcn primitives where appropriate.

| Motion type | Rule |
| --- | --- |
| Page transitions | Avoid broad animated route transitions unless they improve orientation. |
| Dialogs/sheets | Use standard shadcn entrance/exit motion. |
| Loading | Prefer skeletons; spinners only for small inline actions. |
| Hover | Use short color transitions, not movement except the existing button active press. |
| Expand/collapse | Animate height/opacity subtly and keep content accessible. |
| Reduced motion | Respect `prefers-reduced-motion`. |

Avoid attention-grabbing loops, heavy parallax, decorative background motion, and animation that delays operations.

## 19. Content and Microcopy

Tone: clear, technical, calm, helpful, and precise. Avoid overly casual copy and vague failure messages.

| UI text | Rule |
| --- | --- |
| Button labels | Use verbs: `Create API key`, `Rotate key`, `Save policy`, `Retry request`. |
| Empty states | State what is absent and give the next useful action. |
| Errors | Name the resource and the likely fix. |
| Confirmations | State consequence clearly: `This revokes access for all services using this key.` |
| Tooltips | Short explanation, not critical-only content. |
| Helper text | Explain constraints or operational impact. |
| Security warnings | Be direct and specific. |
| Billing warnings | Include amount, threshold, and consequence. |

Bilingual readiness:

| Area | Rule |
| --- | --- |
| Labels | Keep English and Persian labels concise but not cryptic. |
| Numbers | Support Persian numbers where product requirements demand them; technical counters may stay Latin. |
| Technical strings | Keep LTR and monospace. |
| Translation length | Allow longer Persian labels and avoid fixed-width text buttons unless necessary. |

Avoid vague messages like `Something went wrong`, `Invalid input`, and `Failed`. Prefer actionable messages.

## 20. Page Templates

### Dashboard Page

Structure: page title, system alerts, summary metrics, charts, recent activity, and operational links. Put urgent configuration issues above charts.

### List Page

Structure: page header, primary action, search, filters, table, pagination, empty state, loading skeleton, and error state. Preserve filters/sort in URL when server-backed.

### Detail Page

Structure: header with resource name and status, metadata strip, tabs, activity, configuration, usage, and danger zone. Place dangerous actions at the bottom or behind a clearly labelled tab/section.

### Settings Page

Structure: sectioned forms, helper text for consequences, inline validation, and save/cancel actions. Use read-only states where policy or role prevents editing.

### Create/Edit Page

Structure: clear form sections, preview/review step for sensitive changes, submit actions, and cancellation path. Use dialogs for short create flows and pages/sheets for long forms.

### Trace Detail Page

Structure: timeline, metadata, provider/model, cost/token breakdown, latency, error details, request/response tabs, JSON/code viewer, and copyable IDs. Keep raw payloads LTR.

## 21. Implementation Guidelines for Coding Agents

Follow these rules for UI work in this repository:

| Rule | Detail |
| --- | --- |
| Read local Next docs | Before changing Next.js behavior, read the relevant file in `node_modules/next/dist/docs/`. This project uses Next `16.2.9`. |
| Reuse first | Use existing components before creating new primitives. Current shared primitive is `components/ui/button.tsx`. |
| Prefer shadcn | Add shadcn/ui primitives for common controls instead of inventing custom behavior. |
| Use tokens | Use Tailwind token utilities from `app/globals.css`. Do not hardcode colors or random spacing. |
| Use `cn()` | Compose conditional classes with `cn()` from `lib/utils.ts`. |
| Keep RSC defaults | In App Router, pages/layouts are Server Components by default. Add `"use client"` only where state, effects, handlers, browser APIs, or client hooks are required. |
| Keep boundaries small | Do not mark full pages/client shells as client components unless needed. |
| Support RTL | Use logical utilities and preserve LTR technical strings. |
| Include states | Every data-driven component needs loading, empty, and error states. |
| Use types | Define TypeScript props for components and domain models. |
| Keep composition small | Build domain components from shadcn primitives and small presentational parts. |
| Avoid new UI libraries | Do not add UI/chart/table libraries without strong justification. |
| Use semantic HTML | Tables for tabular data, forms for forms, buttons for actions, links for navigation. |
| Verify responsive behavior | Check long Persian labels, long English IDs, and narrow screens. |

Suggested future structure as Shahrah grows:

```txt
app/
  [lang]/
    (app)/
      dashboard/page.tsx
      models/page.tsx
      providers/page.tsx
      traces/[traceId]/page.tsx
components/
  ui/
  shell/
  data-display/
  forms/
  domain/
    models/
    providers/
    traces/
    billing/
lib/
  utils.ts
  i18n/
  formatting/
  constants/
```

## 22. File and Component Naming Conventions

Use existing root folders unless a larger reorganization is intentional.

| Type | Convention | Example |
| --- | --- | --- |
| UI primitives | `components/ui/{name}.tsx` | `components/ui/button.tsx` |
| App shell | `components/shell/{name}.tsx` | `components/shell/app-sidebar.tsx` |
| Domain components | `components/domain/{domain}/{name}.tsx` | `components/domain/models/model-card.tsx` |
| Page components | Colocate under route `_components` for route-specific UI, or domain folder if reused. | `app/[lang]/(app)/models/_components/model-filters.tsx` |
| Hooks | `hooks/use-{thing}.ts` | `hooks/use-copy-to-clipboard.ts` |
| Utilities | `lib/{domain}/{name}.ts` or `lib/{name}.ts` | `lib/formatting/currency.ts` |
| Types | `types/{domain}.ts` or colocated when local | `types/provider.ts` |
| Constants | `lib/constants/{domain}.ts` | `lib/constants/navigation.ts` |
| Components | PascalCase exports | `ModelCard` |
| Files | kebab-case for component filenames | `model-card.tsx` |
| Variants | Use `cva` for meaningful reusable variants | `badgeVariants` |

Route folders follow Next App Router conventions: route groups `(app)`, dynamic segments `[traceId]`, private folders `_components`, and special files `page.tsx`, `layout.tsx`, `loading.tsx`, `error.tsx`, `not-found.tsx`.

## 23. Do and Don't Examples

| Case | Do | Don't |
| --- | --- | --- |
| Card design | Use one bordered `bg-card` surface with clear title, status, and key metrics. | Nest cards inside cards or use oversized marketing panels for operational data. |
| Table design | Use compact rows, sortable headers, visible status badges, and copy actions for IDs. | Hide important state in tooltips or force users to parse long raw strings. |
| RTL | Use `ps-*`, `pe-*`, `ms-*`, `me-*`, `text-start`, `text-end`, and `dir` on technical strings. | Use `ml-*`, `mr-*`, `left-*`, `right-*`, `text-left`, or `text-right` by default. |
| Status badges | Use semantic labels like `Connected`, `Degraded`, `Expired`, `Misconfigured`. | Use color-only dots or vague labels like `OK` and `Bad`. |
| Form errors | `Max monthly tokens must be greater than the current usage: 8.2M.` | `Invalid input.` |
| Empty states | `No traces found for this period. Change filters or send a test request.` | `Nothing here.` |
| Technical IDs | Show truncated monospace ID with copy action and full value available. | Wrap a long ID across the whole table and break row scanning. |

Status badge example:

```tsx
const statusClassName = {
  connected: "bg-success/10 text-success",
  degraded: "bg-warning/10 text-warning",
  expired: "bg-destructive/10 text-destructive",
  pending: "bg-muted text-muted-foreground",
}
```

Do not use this exact object until `success` and `warning` tokens are added; it defines the intended pattern.

## 24. Checklist Before Merging UI Changes

- Uses existing design tokens from `app/globals.css`.
- Avoids hardcoded colors, random spacing, and one-off visual styles.
- Supports RTL with logical spacing, alignment, placement, and mirrored directional icons.
- Preserves LTR for API keys, URLs, request IDs, model IDs, JSON, and code.
- Works in dark mode if `.dark` is applied.
- Has loading, empty, and error states for data-driven UI.
- Is keyboard accessible.
- Uses semantic HTML.
- Has visible focus states.
- Uses consistent typography and sensible text hierarchy.
- Handles long Persian labels and long English technical strings.
- Has responsive behavior for desktop, tablet, and mobile.
- Uses copy actions for important technical values.
- Uses clear, actionable microcopy.
- Does not introduce a new UI library without justification.
- Keeps Server/Client Component boundaries appropriate for Next.js 16 App Router.
- Reuses shadcn/ui primitives and existing shared components where possible.
- Includes confirmations for destructive or infrastructure-impacting actions.
- Does not modify unrelated files while implementing UI work.


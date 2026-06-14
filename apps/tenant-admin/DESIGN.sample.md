---
version: alpha
name: GodarAi Linear
website: "https://linear.app"
description: "A GodarAi-ready adaptation of the Linear design system: a near-black product-focused canvas built around \"#010102\", light gray text (\"#f7f8f8\"), and Linear lavender-blue (\"#5e6ad2\") as the single chromatic accent. Typography is replaced with Pelak for Persian UI, direction is RTL by default, and LTR is reserved only for API keys, URLs, emails, code, and model names. Cards live as charcoal panels (\"#0f1011\") with hairline borders. The accent lavender appears on focus rings and intentional CTAs — never decoratively."

seo:
  title: "GodarAi Linear Design System — Pelak Persian UI, RTL default, Lavender (#5e6ad2), 22 components"
  metaDescription: "GodarAi-ready Linear DESIGN.md. Lavender-blue #5e6ad2, Pelak Persian typography, RTL default, LTR only for technical identifiers, 22 colors, 22 components. For React, Next.js, shadcn/ui, and AI tools."
  highlights:
    - "Deepest dark canvas in the directory — #010102 with a faint blue tint, never pure #000000 true black"
    - "Single chromatic accent — lavender-blue #5e6ad2 reserved for brand mark, focus ring, and primary CTA only"
    - "Four-step surface ladder carries hierarchy without drop shadows — canvas to surface-4 with hairline borders"
    - "Pelak Persian typography — no Latin-style negative tracking on Persian UI text"
    - "RTL by default — LTR only for API keys, URLs, emails, code, and model names"
    - "Product UI screenshots are the protagonist — marketing chrome is a dark frame for high-fidelity app captures"
  tags:
    - "Project Management"
    - "Productivity & SaaS"
  lastUpdated: "2026-05-12"
  author:
    name: "Dov Azencot"
    url: "https://x.com/dovazencot"
  opening: |
    Linear's marketing site is the strictest dark-canvas system in this directory. The page floor is "#010102" — essentially pure black with a faint blue cast, never the unfortunate "#000000" true black that most dark sites default to. On top sits a four-step surface ladder: charcoal panels at "#0f1011", "#141516", "#18191a", and "#191a1b" carrying every card, hovered tile, and lifted dropdown. Hairlines run from "#23252a" up through "#34343a" and "#3e3e44". The single chromatic accent is Linear's lavender-blue "#5e6ad2", deployed scarcely on the brand mark, focus rings, and one primary CTA per section. There is no second chromatic color, no atmospheric gradient, no spotlight card. Linear trusts surface lift and hairline borders to carry every bit of hierarchy.

    This page packages the entire spec into one DESIGN.md file written to the Google Labs spec and adapted for GodarAi. Inside: 22 color tokens grouped into brand, surface, hairline, text, and semantic families; 14 type tokens using Pelak for Persian UI; an RTL-first direction system with strict LTR exceptions for technical identifiers; an 8-step rounded scale (4px through 24px plus pill and full); a 9-step spacing scale on a 4px base unit ending at 96px section padding; and 22 component definitions covering buttons, pricing cards, product screenshot panels, testimonial cards, changelog rows, and the dark top-nav and footer.

    Feed the file to Claude, Cursor, or GitHub Copilot and the agent reproduces Linear's discipline — dark canvas, single lavender accent, surface ladder over shadows — while using Pelak for Persian UI and RTL as the default layout direction. Reference the tokens directly inside Tailwind config, CSS variables, or a shadcn/ui component library.
  related:
    - href: "/design"
      title: "Browse all design systems"
      description: "The full directory of DESIGN.md files on shadcn.io, with live mockups for each."
    - href: "https://linear.app/brand"
      title: "Linear brand guidelines"
      description: "Linear's official brand site — wordmark, color usage, and the lavender accent token in context."
    - href: "https://github.com/google-labs-code/design.md"
      title: "The DESIGN.md specification"
      description: "Google Labs' open spec for machine-readable design system files — the format this page is built on."
  questions:
    - id: "primary-color"
      title: "What is Linear's primary brand color?"
      answer: "Linear's brand color is lavender-blue at \"#5e6ad2\" — a desaturated blue-violet sitting between Slack purple and Discord blurple. The hover state lightens to \"#828fff\" and the focus tint settles at \"#5e69d1\". Linear uses lavender scarcely: brand mark, primary CTA, focus ring, and link emphasis only. It never fills cards, never appears as a section background, and never pairs with a second chromatic accent."
    - id: "dark-mode"
      title: "Does Linear's design system have a light mode?"
      answer: "No — the marketing site is dark-only. Canvas is \"#010102\" (essentially pure black with a faint blue tint), with a four-step surface ladder running through \"#0f1011\", \"#141516\", \"#18191a\", and \"#191a1b\" for lifted panels. The spec documents three inverse tokens (\"#ffffff\" canvas, \"#f5f6f6\" and \"#f6f7f7\" surfaces) used only on the rare white pill CTA over a dark section opener. Linear's product UI ships a light theme, but it is not represented in this DESIGN.md."
    - id: "typography"
      title: "What typography does GodarAi use in this Linear adaptation?"
      answer: "GodarAi replaces Linear Display and Linear Text with Pelak for all Persian UI text. Persian text should not use Latin-style negative tracking, so the typography tokens set letterSpacing to 0. Technical identifiers such as API keys, URLs, emails, code, and model names use an LTR mono stack."
    - id: "negative-tracking"
      title: "Should Persian UI text use negative letter-spacing?"
      answer: "No. The original Linear spec uses aggressive negative tracking for Latin display type, but GodarAi uses Pelak for Persian UI and keeps letterSpacing at 0 to preserve Persian/Arabic glyph shaping and readability. Use hierarchy through font size, weight, spacing, and surface contrast instead."
    - id: "use-in-project"
      title: "Can I use this DESIGN.md to build my own React project management tool?"
      answer: "Yes — this GodarAi adaptation is designed to be fed into Claude, Cursor, or any AI tool that reads structured design tokens. The agent should reproduce Linear's restraint: dark canvas, single lavender accent, surface ladder over shadows, Pelak Persian typography, and RTL-first layout rules. Reference the tokens directly inside Tailwind config or CSS variables."
    - id: "known-gaps"
      title: "What's missing from this DESIGN.md spec?"
      answer: "A handful of things, documented in the Known Gaps section: form-field error and validation styling is not visible on the inspected pages; light mode is not documented because Linear's marketing site does not ship a light theme; and Linear's actual product UI uses a richer color-tag palette (red, orange, yellow, green, blue, purple) for issue priorities and project labels that lives inside the in-product surfaces shown in mockups. The four-step surface ladder values are extracted directly from Linear's canonical CSS variables."

colors:
  primary: "#5e6ad2"
  on-primary: "#ffffff"
  primary-hover: "#828fff"
  primary-focus: "#5e69d1"
  ink: "#f7f8f8"
  ink-muted: "#d0d6e0"
  ink-subtle: "#8a8f98"
  ink-tertiary: "#62666d"
  canvas: "#010102"
  surface-1: "#0f1011"
  surface-2: "#141516"
  surface-3: "#18191a"
  surface-4: "#191a1b"
  hairline: "#23252a"
  hairline-strong: "#34343a"
  hairline-tertiary: "#3e3e44"
  inverse-canvas: "#ffffff"
  inverse-surface-1: "#f5f6f6"
  inverse-surface-2: "#f6f7f7"
  inverse-ink: "#000000"
  brand-secure: "#7a7fad"
  semantic-success: "#27a644"
  semantic-overlay: "#000000"

fontAssets:
  primaryFamily: Pelak
  recommendedWebFormat: woff2
  basePath: /fonts/pelak/
  usage: "Use Pelak for all Persian UI text. Prefer WOFF2 in production; keep WOFF only as legacy fallback if needed."
  files:
    thin:
      weight: 100
      file: Pelak-Thin.woff2
    light:
      weight: 300
      file: Pelak-light.woff2
    regular:
      weight: 400
      file: Pelak-Regular.woff2
    medium:
      weight: 500
      file: Pelak-Medium.woff2
    semibold:
      weight: 600
      file: Pelak-SemiBold.woff2
    bold:
      weight: 700
      file: Pelak-Bold.woff2
    extrabold:
      weight: 800
      file: Pelak-ExtraBold.woff2
    black:
      weight: 900
      file: Pelak-Black.woff2
  nextFont:
    recommended: next/font/local
    variable: --font-pelak
    className: font-pelak

textDirection:
  default: rtl
  language: fa
  htmlDir: rtl
  textAlign: start
  layoutRules:
    - "Use CSS logical properties: margin-inline, padding-inline, inset-inline-start, inset-inline-end, border-inline."
    - "Do not hard-code left/right unless the element is intentionally LTR."
    - "Use dir=rtl at the document or app shell level."
    - "Use dir=ltr only for the technical identifiers listed below."
  ltrOnly:
    - api-key
    - url
    - email
    - code
    - model-name
  ltrRendering:
    direction: ltr
    unicodeBidi: plaintext
    textAlign: left
    typography: "{typography.mono}"
    examples:
      apiKey: sk-gdr-1234567890abcdef
      url: https://api.godarai.com/v1/chat/completions
      email: support@godarai.com
      code: "const model = 'gdr-fast';"
      modelName: gdr-fast

typography:
  display-xl:
    fontFamily: Pelak
    fontSize: 80px
    fontWeight: 600
    lineHeight: 1.05
    letterSpacing: 0
  display-lg:
    fontFamily: Pelak
    fontSize: 56px
    fontWeight: 600
    lineHeight: 1.10
    letterSpacing: 0
  display-md:
    fontFamily: Pelak
    fontSize: 40px
    fontWeight: 600
    lineHeight: 1.15
    letterSpacing: 0
  headline:
    fontFamily: Pelak
    fontSize: 28px
    fontWeight: 600
    lineHeight: 1.20
    letterSpacing: 0
  card-title:
    fontFamily: Pelak
    fontSize: 22px
    fontWeight: 500
    lineHeight: 1.25
    letterSpacing: 0
  subhead:
    fontFamily: Pelak
    fontSize: 20px
    fontWeight: 400
    lineHeight: 1.40
    letterSpacing: 0
  body-lg:
    fontFamily: Pelak
    fontSize: 18px
    fontWeight: 400
    lineHeight: 1.50
    letterSpacing: 0
  body:
    fontFamily: Pelak
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.50
    letterSpacing: 0
  body-sm:
    fontFamily: Pelak
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.50
    letterSpacing: 0
  caption:
    fontFamily: Pelak
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.40
    letterSpacing: 0
  button:
    fontFamily: Pelak
    fontSize: 14px
    fontWeight: 500
    lineHeight: 1.20
    letterSpacing: 0
  eyebrow:
    fontFamily: Pelak
    fontSize: 13px
    fontWeight: 500
    lineHeight: 1.30
    letterSpacing: 0
  mono:
    fontFamily: "Geist Mono, JetBrains Mono, SFMono-Regular, Consolas, monospace"
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.50
    letterSpacing: 0

rounded:
  xs: 4px
  sm: 6px
  md: 8px
  lg: 12px
  xl: 16px
  xxl: 24px
  pill: 9999px
  full: 9999px

spacing:
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 24px
  xl: 32px
  xxl: 48px
  section: 96px

components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: 8px 14px
  button-primary-pressed:
    backgroundColor: "{colors.primary-focus}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
    textColor: "{colors.on-primary}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
  button-secondary:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: 8px 14px
  button-tertiary:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: 8px 14px
  button-inverse:
    backgroundColor: "{colors.inverse-canvas}"
    textColor: "{colors.inverse-ink}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: 8px 14px
  pricing-card:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.lg}"
    padding: 24px
  pricing-card-featured:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.lg}"
    padding: 24px
  feature-card:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.lg}"
    padding: 24px
  product-screenshot-card:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.xl}"
    padding: 24px
  testimonial-card:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body-lg}"
    rounded: "{rounded.lg}"
    padding: 32px
  customer-logo-tile:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink-subtle}"
    typography: "{typography.caption}"
    rounded: "{rounded.xs}"
    padding: 16px
  text-input:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: 8px 12px
  text-input-focused:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: 8px 12px
  pricing-tab-default:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink-subtle}"
    typography: "{typography.button}"
    rounded: "{rounded.pill}"
    padding: 6px 14px
  pricing-tab-selected:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.pill}"
    padding: 6px 14px
  cta-banner:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    typography: "{typography.headline}"
    rounded: "{rounded.lg}"
    padding: 48px
  changelog-row:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.xs}"
    padding: 24px 0
  status-badge:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink-muted}"
    typography: "{typography.caption}"
    rounded: "{rounded.pill}"
    padding: 2px 8px
  top-nav:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.xs}"
    height: 56px
  footer:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink-subtle}"
    typography: "{typography.caption}"
    rounded: "{rounded.xs}"
    padding: 64px 32px
---

## Overview

Linear's marketing canvas is the deepest dark surface in this collection — `{colors.canvas}` is "#010102", essentially pure black with a faint blue tint. On top sits a four-step surface ladder (`{colors.surface-1}` through `{colors.surface-4}`) for cards, panels, and lifted tiles, with hairline borders running from `{colors.hairline}` ("#23252a") up through `{colors.hairline-strong}` and `{colors.hairline-tertiary}`. Light gray text (`{colors.ink}` "#f7f8f8") carries the body and headlines.

The single chromatic accent is **Linear lavender-blue** `{colors.primary}` ("#5e6ad2") — used on the brand mark, focus rings, and the primary CTA button. A lighter hover state (`{colors.primary-hover}` "#828fff") and a focus-tinted variant (`{colors.primary-focus}` "#5e69d1") extend the same hue. Linear avoids saturated greens, oranges, reds, etc. on the marketing canvas — the only semantic color is `{colors.semantic-success}` ("#27a644") for status pills and the rare success indicator.

GodarAi Persian UI type runs **Pelak** across display, headings, body, captions, and buttons. Letter spacing is set to `0` for Persian text to preserve glyph shaping and readability. A mono LTR stack is reserved only for API keys, URLs, emails, code, and model names.

The page rhythm is **dense product screenshots** — Linear's marketing leads with high-fidelity captures of the product UI (issue list, project view, dashboard) framed in `{colors.surface-1}` panels with `{rounded.xl}` 16px corners. The chrome is intentionally minimal so the app screenshots can do the heavy lifting.

**Key Characteristics:**
- **Dark-canvas marketing system** — `{colors.canvas}` ("#010102") is the deepest dark in this collection.
- **Lavender-blue brand accent** (`{colors.primary}` "#5e6ad2") — used scarcely on brand mark, focus, and the primary CTA.
- Four-step surface ladder (canvas → surface-1 → surface-2 → surface-3 → surface-4) carries hierarchy without shadow.
- **Pelak Persian UI typography** replaces Linear Display/Text; letter-spacing stays at `0` for Persian text.
- Cards use `{rounded.lg}` 12px corners with 1px hairline borders — never pill, rarely 16px.
- **RTL by default** for all Persian UI; LTR is used only for API keys, URLs, emails, code, and model names.
- **Product UI screenshots** dominate the page. The marketing chrome is a dark frame for the app.
- No second chromatic color. No atmospheric gradients. No spotlight cards.


## GodarAi Persian + RTL Adaptation

- Use `Pelak` for every Persian UI surface: headings, body, captions, buttons, empty states, loading states, errors, and navigation.
- Persian UI is `rtl` by default at the root/app-shell level.
- Use `text-align: start` for normal Persian text and avoid hard-coded `left`/`right` positioning.
- Use CSS logical properties for spacing and placement: `margin-inline`, `padding-inline`, `inset-inline-start`, `inset-inline-end`, and `border-inline`.
- Use `dir="ltr"` only for API keys, URLs, emails, code snippets, and model names.
- Technical LTR identifiers should use the mono token and preserve their internal character order.
- Do not apply negative letter-spacing to Persian text.

### Recommended font placement

Place Pelak WOFF2 files in:

```txt
/public/fonts/pelak/
```

Required files:

```txt
Pelak-Thin.woff2
Pelak-light.woff2
Pelak-Regular.woff2
Pelak-Medium.woff2
Pelak-SemiBold.woff2
Pelak-Bold.woff2
Pelak-ExtraBold.woff2
Pelak-Black.woff2
```

## Known Gaps

- The four-step surface ladder values are extracted directly from Linear's `--color-bg-level-3`, `--color-line-tint`, etc. CSS variables; they are Linear's canonical surface spec.
- Form-field error and validation styling is not visible on the inspected pages.
- Light mode is not documented because the marketing site does not ship a light theme.
- Linear's actual product UI uses a richer color-tag palette (red, orange, yellow, green, blue, purple) for issue priorities and project labels — those colors live in the in-product surfaces shown in mockups.
- The original Linear custom display, text, and mono families are replaced in this GodarAi adaptation. Use Pelak for Persian UI and the configured mono stack only for technical LTR identifiers.

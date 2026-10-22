---
name: GodarAI Website
description: A bright, orderly, Persian-first workshop where every team's AI work is visible, governed and paid for.
colors:
  accent: "oklch(0.5 0.1 232)"
  accent-hover: "oklch(0.43 0.09 235)"
  on-accent: "oklch(0.99 0 0)"
  ground: "oklch(0.985 0.005 220)"
  panel: "oklch(0.962 0.013 220)"
  panel-deep: "oklch(0.93 0.025 222)"
  line: "oklch(0.9 0.015 225)"
  ink: "oklch(0.21 0.025 245)"
  ink-soft: "oklch(0.46 0.025 240)"
  night: "oklch(0.2 0.03 245)"
  night-ink: "oklch(0.96 0.01 220)"
  night-soft: "oklch(0.78 0.02 225)"
  night-strong: "oklch(0.995 0 0)"
  tone-guard: "oklch(0.955 0.03 160)"
  tone-govern: "oklch(0.955 0.025 295)"
  tone-observe: "oklch(0.955 0.03 230)"
  tone-cost: "oklch(0.96 0.035 80)"
typography:
  display:
    fontFamily: "Vazirmatn, ui-sans-serif, system-ui, sans-serif"
    fontSize: "3.75rem"
    fontWeight: 600
    lineHeight: 1.375
    letterSpacing: "0em"
  headline:
    fontFamily: "Vazirmatn, ui-sans-serif, system-ui, sans-serif"
    fontSize: "3rem"
    fontWeight: 600
    lineHeight: 1.375
    letterSpacing: "0em"
  title:
    fontFamily: "Vazirmatn, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.6
  body:
    fontFamily: "Vazirmatn, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 400
    lineHeight: 2
  label:
    fontFamily: "Vazirmatn, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.6
rounded:
  md: "8px"
  lg: "10px"
  xl: "14px"
  2xl: "16px"
  3xl: "24px"
  full: "9999px"
spacing:
  section: "80px"
  section-lg: "128px"
  container: "72rem"
  container-hero: "80rem"
  gutter: "24px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    rounded: "{rounded.full}"
    padding: "0 28px"
    height: "48px"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  button-quiet:
    textColor: "{colors.ink-soft}"
    rounded: "{rounded.full}"
    padding: "0 20px"
    height: "48px"
  button-quiet-hover:
    textColor: "{colors.ink}"
  panel:
    backgroundColor: "{colors.panel}"
    rounded: "{rounded.3xl}"
    padding: "40px"
  product-window:
    backgroundColor: "{colors.ground}"
    rounded: "{rounded.3xl}"
  closing-band:
    backgroundColor: "{colors.night}"
    textColor: "{colors.night-ink}"
---

# Design System: GodarAI Website

## Overview

**Creative North Star: "The Lit Workshop"**

A good workshop in daylight: wide benches, every tool on its hook, each team's work laid out where the manager can see it. Nothing is hidden and nothing is shouting. The visitor is a technology lead deciding whether AI can join the daily work of every team, and the page should feel like walking through that workshop: real work in progress, clearly arranged, with the controls in plain sight.

The structure follows dust.tt, the reference the owner pinned: a bright open canvas, large confident headings, short copy, one idea per section, soft colored panels, real product UI resting on those panels, tabs by team, numbered pillars and one dark closing band. The product is the picture. Real screenshots of GodarChat and the admin panel, framed as windows on calm tinted panels, carry the argument; decorative illustration does not.

The system is product-centered and alive: motion shows an assistant doing work (a message arriving, a tool call completing, a cost meter settling) rather than decorating the page. Persian and right-to-left are the home language of every decision, and light and dark themes are equal citizens.

**Key Characteristics:**
- Dust's structure and clarity, in Persian, right to left.
- Real product screenshots in window frames on soft colored panels; the product is the hero image.
- A small family of calm pastel panels, one per differentiator, all at the same lightness.
- One steady blue accent for actions; neutrals tinted cool.
- Cards lift slightly off the bench with soft shadows.
- Motion that demonstrates work, never ambient decoration.

## Colors

Calm daylight neutrals tinted cool, one steady blue accent, and a set of soft pastel panels that differ by hue at matched lightness.

All values are OKLCH and live in `app/styles/theme.css` as `--site-*` and `--tone-*` custom properties, exposed to Tailwind as `site-*` and `tone-*` utilities. The frontmatter gives light-theme values; `.dark` redefines each one, except the night tokens, which are the same in both themes.

### Primary
- **Steady Blue** (accent): primary buttons, links, active tab indicator, focus outlines, and the single emphasized word in a heading. In dark mode it lifts to a lighter cyan-blue (`oklch(0.74 0.09 215)`) with dark text on it.
- **Steady Blue Deep** (accent-hover): hover and pressed state of primary actions.

### Neutral
- **Daylight Paper** (ground): the page canvas and the surface of product windows. Not pure white, not cream.
- **Workbench Grey** (panel): the default section panel and quiet inner containers.
- **Tool-Steel Mist** (panel-deep): icon wells, selected rows, the active tab background.
- **Pencil Line** (line): borders and dividers inside windows and cards.
- **Workshop Ink** (ink): headings and primary text.
- **Graphite** (ink-soft): body copy, captions, quiet buttons.
- **Evening Ink** (night) with **Lamp White** (night-ink), **Dusk Grey** (night-soft) and **Lamp Glow** (night-strong, the hover of the light button): the closing band, once per page.

### Differentiator tones
Four pastel tints, one per differentiator, so a visitor learns the color of each promise across pages. The mapping lives in `app/lib/pillars.ts` (`toneSurface`); pages never pick a tone by hand. Workshop Ink, Graphite and Steady Blue all hold WCAG AA on every tone in both themes (ink above 13:1, Graphite above 6:1, accent above 5:1).

- **Sage** (tone-guard, semantic guardrails): dark `oklch(0.26 0.03 160)`.
- **Lilac** (tone-govern, governance): dark `oklch(0.26 0.03 295)`.
- **Sky** (tone-observe, observability): dark `oklch(0.26 0.03 230)`.
- **Sand** (tone-cost, cost control): dark `oklch(0.27 0.03 80)`.

### Named Rules
**The Matched Lightness Rule.** Differentiator panels change hue, never lightness or chroma. Side by side they read as one family; none may outshout another.

**The One Accent Rule.** Steady Blue is the only saturated color on the page. Panels are tinted, never saturated, and are never used for text or buttons.

**The Color Means a Promise Rule.** Each differentiator keeps its panel color on every page it appears. A panel color is not chosen for variety.

## Typography

**Display and body:** Vazirmatn (self-hosted from `/public/fonts/vazirmatn`, weights 400 to 700).

**Character:** Large, even, confident Persian headings with generous leading; body text that reads like a calm advisor, not like UI chrome.

### Hierarchy
- **Display** (600, 2.25rem to 3.75rem responsive, line-height 1.375): the hero heading only, `text-balance`.
- **Headline** (600, 1.875rem to 3rem responsive, line-height 1.375): section headings, `text-balance`.
- **Title** (600, 1.25rem, line-height 1.6): card and panel titles.
- **Body** (400, 1.125rem, line-height 2): section copy in Graphite, measure 60 to 70 characters.
- **Label** (500, 0.875rem): buttons, tabs, window titles, captions; 0.75rem for window metadata.

### Named Rules
**The No Tracking Rule.** Never apply negative or positive letter-spacing to Persian text; it breaks Arabic-script joining.

**The Quiet Heading Rule.** One weight step above body is enough; size carries the hierarchy, not stacked bold, color and caps. No eyebrows above headings; numbering only where the sequence carries meaning.

Persian digits in running copy; Latin product and model names stay LTR inside the line.

## Layout

Content sections sit inside a 72rem container (narrow text sections such as the FAQ use 48rem; the hero uses 80rem) with a 24px gutter, and stack with 80px vertical rhythm, rising to 128px on large screens. The `Section` pattern applies this rhythm. Each section holds one idea: a headline, one or two sentences, one visual. Two-column pairs alternate sides; in RTL the first column is the right one and reading order follows it. Product windows may overhang their panel's edge to show they are real objects resting on the bench.

Mobile is a single column with the visual after the copy; tabs become a horizontal scroller; nothing depends on hover.

## Elevation & Depth

A light hybrid: surfaces step by tone, and objects that represent product (windows, cards that can be opened) lift off them with soft, ink-tinted shadows.

### Shadow Vocabulary
Each is a Tailwind utility defined in `theme.css`; no page writes a shadow value by hand.
- **Window lift** (`shadow-window`: `0 20px 25px -5px` and `0 8px 10px -6px`, tinted `oklch(0.21 0.025 245 / 0.1)`; in dark mode the tint is Workshop Ink's light value at the same alpha): product windows and screenshots resting on a panel.
- **Card rest** (`shadow-card`: `0 10px 40px -10px rgba(0,0,0,0.1)`, `0.5` alpha in dark mode): white cards on the ground, and tinted cards on hover and focus.
- **Hairline** (`shadow-hairline`: `0 1px 2px rgba(0,0,0,0.05)`): small controls and chips.

### Named Rules
**The Tinted Shadow Rule.** Shadows are tinted with Workshop Ink and offset downward. No zero-offset glows, no hard offset shadows, no glass or backdrop blur.

## Shapes

Friendly and precise. Panels and product windows use 24px corners; cards and inner containers 16px; icon wells 14px; buttons, chips and tab pills are fully round. Screenshots inside windows are clipped to the window's corner. No decorative blobs, waves or masks.

## Components

### Buttons
Round, calm and unmistakable.
- **Shape:** fully round pill, 48px tall.
- **Primary:** Steady Blue background, white text, 28px inline padding, label weight 500.
- **Hover / Focus:** background deepens to Steady Blue Deep; focus is a 2px Steady Blue outline offset by 2px.
- **Quiet:** no background; Graphite text that darkens to Workshop Ink on hover. Used for the secondary action beside a primary.

### Panels
- **Corner Style:** 24px.
- **Background:** Workbench Grey by default; a differentiator tint when the section belongs to one of the four promises.
- **Shadow Strategy:** none; panels are the bench, not objects.
- **Internal Padding:** 40px on desktop, 24px on mobile.

### Pillar Cards
- **Surface:** the differentiator's own tone, 24px corners, flat at rest.
- **Content:** an icon in a 40px well (Daylight Paper at 70%), the title in Title size, a one-line summary in Graphite, and a trailing arrow.
- **Hover / Focus:** Card rest shadow appears and the arrow nudges toward reading direction; focus also draws the 2px Steady Blue outline.

### Product Window (signature)
The hero image of the system: a real screenshot, or an illustrative coded window until screenshots arrive, framed as an app window.
- **Frame:** Daylight Paper surface, 24px corners, Window lift shadow, a title bar with the window name in label size and a Pencil Line divider below it.
- **Content:** real GodarChat or admin screenshots; coded mocks are labelled «نمونه» and never carry invented numbers presented as fact.
- **Motion:** one demonstration per window, such as a tool row resolving from a spinner to a Steady Blue check. Plays once on entry and respects reduced motion.

### Team Tabs
- **Style:** equal-width tabs over a Pencil Line rule; the active tab is semibold Workshop Ink with a 2px Workshop Ink underline, inactive tabs are Graphite. On mobile the row scrolls horizontally.
- **Behavior:** switching a tab swaps the media and task list beside it. Arrow keys move in reading direction (left is forward in RTL), Home and End jump to the ends, and only the selected tab is in the tab order.

### Navigation
- **Style:** Daylight Paper bar with a Pencil Line bottom border; label-size links in Graphite, Workshop Ink on hover; a mega menu of icon-and-description rows on Workbench Grey hover.
- **Actions:** a quiet «ورود» and a primary pill for the trial; the demo link sits beside them.
- **Mobile:** a full-height sheet with 48px rows and the two actions pinned at the bottom.

### Closing Band
- Evening Ink full-width band, used once per page, holding one large headline in Lamp White, one sentence in Dusk Grey and the primary action.

## Do's and Don'ts

### Do:
- **Do** show the real product: real screenshots in Product Windows on tinted panels.
- **Do** keep each differentiator on its own panel color everywhere it appears.
- **Do** use motion to demonstrate an assistant's work, once, on entry, and honor `prefers-reduced-motion`.
- **Do** build and check every section in light and dark themes.
- **Do** build pages from the shared patterns in `app/components/patterns` (live samples at `/_patterns`) and the `site-*` / `tone-*` utilities; never raw palette colors.

### Don't:
- **Don't** use water, ford or river illustration, ripples or wave motifs; that world is retired.
- **Don't** introduce a second saturated color or use a panel tint for text or buttons.
- **Don't** use indigo, slate or other raw Tailwind palette colors; feature pages that still do are debt to migrate.
- **Don't** apply letter-spacing to Persian text.
- **Don't** present invented numbers, logos or quotes inside mocks as real.
- **Don't** use glass, glows, gradients on text, or decorative blobs.

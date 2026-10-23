---
name: GodarAI Website
description: Bold and witty, on a confident navy ground with warm paper surfaces and one loud signal orange.
colors:
  surface: "#ffffff"
  ground: "#f6f4ef"
  panel: "#ebe7dd"
  sand: "#d9cdb4"
  ink: "#0f1b33"
  ink-soft: "#4b5569"
  line: "#d6d2c7"
  line-strong: "#7b8396"
  accent: "#ff6a13"
  accent-hover: "#ff7f36"
  on-accent: "#0f1b33"
  accent-text: "#ad3c08"
  focus: "#ad3c08"
  success: "#1f6f8b"
  danger: "#b3261e"
  night: "#0f1b33"
  night-raised: "#1d2d50"
  night-ink: "#f6f4ef"
  night-soft: "#a9b2c4"
typography:
  display:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "3.5rem"
    fontWeight: 800
    lineHeight: 1.15
    letterSpacing: "0em"
  h1:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "2.5rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "0em"
  h2:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "2rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "0em"
  h3:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.4
  lead:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 300
    lineHeight: 1.8
  body:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.8
  label:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.5
  caption:
    fontFamily: "Estedad, Vazirmatn, Tahoma, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.6
rounded:
  field: "4px"
  card: "12px"
  full: "9999px"
spacing:
  space-1: "4px"
  space-2: "8px"
  space-3: "12px"
  space-4: "16px"
  space-6: "24px"
  space-8: "32px"
  space-10: "40px"
  space-20: "80px"
  content: "75rem"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.label}"
    rounded: "{rounded.full}"
    padding: "10px 16px"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  button-secondary:
    textColor: "{colors.ink}"
    typography: "{typography.label}"
    rounded: "{rounded.full}"
    padding: "10px 16px"
  button-secondary-hover:
    backgroundColor: "{colors.panel}"
  button-on-night:
    textColor: "{colors.night-ink}"
    rounded: "{rounded.full}"
    padding: "10px 16px"
  tile:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.card}"
    padding: "24px"
  tile-hover:
    backgroundColor: "{colors.ground}"
  product-window:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.card}"
  night-band:
    backgroundColor: "{colors.night}"
    textColor: "{colors.night-ink}"
---

# Design System: GodarAI Website

Source: the «GodarAI» design-system artifact (Cowork, October 2026), applied to the website on 2026-10-23. From this point this file is the website's standalone reference: change it here, not in the artifact.

## Overview

**Creative North Star: "The Marked Crossing"**

A ford is safe only where someone has marked it. GodarAI is that marker: a confident navy ground, warm paper to stand on, and one signal orange that says "cross here". The visitor is a technology lead deciding whether AI can join the daily work of every team. The page should feel like an expert who knows the water and has a sense of humour about it: sure-footed, direct, a pun when it earns its place, never at the reader's expense.

Structure stays the Dust-like rhythm the site was built on: one idea per section, a large heading, short copy, a visual. The character is bolder: the hero sits on the navy field, the closing call to action is one inverted navy card, and orange marks exactly one thing per section.

Persian and right-to-left are the home language of every decision. Light and dark themes are equal citizens.

**Key Characteristics:**
- Navy for the big fields: the hero band, the closing card, the footer.
- Warm paper surfaces; sections alternate ground and panel.
- One signal orange per section, almost always the primary button.
- Flat tiles with hairlines; no shadows, no glass.
- Estedad, heavy for display, light for leads, with generous line height.
- Headlines in two short clauses, the second a turn.

## Colors

Warm neutrals on navy, with one loud orange that is rationed.

All values live in `app/styles/theme.css` as `--site-*` custom properties, exposed to Tailwind as `site-*` utilities. The frontmatter gives light-theme values; `.dark` redefines each one.

### Primary
- **Signal Orange** (accent): the primary button fill and nothing else that competes with it in the same section. On it, always **Signal Navy** text (on-accent), never white.
- **Ember** (accent-text): orange for words and links on light surfaces, where Signal Orange would fail contrast (2.6:1). Also the focus ring (focus).

### Neutral
- **Paper** (ground): the page background.
- **Linen** (panel): alternating section bands and hover fills on ground.
- **White Card** (surface): tiles, product windows, inputs.
- **Sand** (sand): warm feature fields: the tile under a product window, icon wells, alternate checklist columns.
- **Ink** (ink) and **Slate** (ink-soft): text and secondary text on every light surface.
- **Hairline** (line) and **Strong Hairline** (line-strong): decorative dividers and tile outlines; input and secondary-button borders.

### Night
- **Signal Navy** (night): the hero band, the closing card and the footer.
- **Deep Navy** (night-raised): layers and hover fills on navy.
- **Paper on Navy** (night-ink) and **Mist** (night-soft): text and secondary text on navy.

### Status
- **Success** is blue and **Danger** red, so they differ by more than hue; always pair each with a word or icon.

### Named Rules
**The One Signal Rule.** At most one Signal Orange element per section, usually the primary button. A check mark, a bullet or an icon is never orange.

**The Ember Rule.** Orange text on a light surface is always Ember, never Signal Orange. Ember holds 4.9:1 on Paper, Linen and White Card but only 3.9:1 on Sand, so it never sets small text on Sand.

Measured pairs (WCAG): Signal Navy on Signal Orange 6.0:1 (on the hover orange 6.8:1); Mist on Signal Navy 8.0:1 and on Deep Navy 6.4:1; Slate on Sand 4.8:1; white on Success 5.7:1.

**The Navy Field Rule.** Navy is for the big fields only: the hero band, one inverted card per page, the footer.

## Typography

**Family:** Estedad (`public/fonts/estedad/Estedad-Variable.woff2`, weights 100 to 900, SIL Open Font License), with Vazirmatn and Tahoma as fallbacks.

**Character:** Heavy, confident display lines; light, airy leads; body text with room to breathe.

### Hierarchy
Each step is a Tailwind utility (`text-display`, `text-h1` … `text-caption`) carrying its own line height.
- **Display** (800, 3.5rem, 1.15): the single hero line. On mobile it steps down to H1.
- **H1** (700, 2.5rem, 1.2): page titles, the closing card heading on desktop.
- **H2** (600, 2rem, 1.3): section headings on desktop.
- **H3** (600, 1.5rem, 1.4): tile and card titles; section headings on mobile.
- **Lead** (300, 1.125rem, 1.8): hero subhead and section intros.
- **Body** (400, 1rem, 1.8): running text.
- **Label** (500, 0.875rem, 1.5): buttons, navigation, form labels.
- **Caption** (400, 0.75rem, 1.6): metadata, window titles, footnotes.

### Named Rules
**The No Tracking Rule.** Never add letter-spacing to Persian; it breaks the joins.

**The Two Clause Rule.** Section headlines are two short clauses, the second a turn: «کنترل را جدی بگیرید. نه دست‌وپاگیر.»

Persian digits in running copy; Latin product and model names stay LTR inside the line. Always write the half-space (ZWNJ) in compounds.

## Layout

Content never runs wider than 1200px (`max-w-content`). Sections have 40px vertical padding on mobile and 80px on desktop, with a 16px mobile gutter (24px from `sm`). The `Section` pattern applies this rhythm. Inside a section: 16px between stacked items, 32px between the heading block and its content, 24px between tiles and inside them.

Sections alternate ground and panel. Feature grids are 2×2 or a row of tiles, each a bold H3 label and one line of body. In two-column rows the first column is the right one in RTL. Mobile is a single column with the visual after the copy; nothing depends on hover.

The navbar shows its full menu from 1280px; below that it collapses into the sheet, because Estedad sets wider than the previous face.

## Elevation & Depth

Flat. Depth comes from surface steps (ground, panel, surface, sand, navy) and 1px hairlines, never from shadows.

### Named Rules
**The Flat Tile Rule.** No box shadows, glows, glass or backdrop blur. A tile is a White Card with a Hairline outline; hover steps its fill one surface (surface to ground, ground to panel).

## Shapes

Buttons and badges are pills. Tiles, cards, product windows and media are 12px (`rounded-card`). Inputs and tags are 4px (`rounded-field`). No decorative blobs, waves or masks.

## Components

### Buttons
Pills in Label type, about 10px by 16px, at least 44px tall.
- **Primary:** Signal Orange fill, Signal Navy text; hover lightens to the hover orange.
- **Secondary:** transparent, 1px Strong Hairline border, Ink text; hover fills Linen.
- **On navy:** transparent with a Paper-on-Navy outline; hover fills Deep Navy.
- **Focus:** 2px solid Ember ring, 2px offset (Paper on Navy on the navy fields).
- **Labels** are verbs of two or three words: «شروع ۱۵ روز رایگان», «درخواست دمو».

### Tiles
- **Surface:** White Card with a Hairline outline, 12px corners, 24px padding (32px on desktop for large tiles).
- **Content:** an icon well (Sand, Ink icon), a bold H3 label, one line of Body in Slate, and a trailing arrow when it links.
- **Hover:** the fill steps to Paper; the arrow nudges toward reading direction.

### Product Window (signature)
Real product UI, or a drawn window labelled «نمونه» until screenshots arrive.
- **Frame:** White Card, Hairline outline, 12px corners, a title bar in Label type over a Hairline divider. No shadow.
- **Tile:** it sits centred on a Sand tile and stays inside it.
- **Motion:** tool rows arrive one by one (350ms apart), then the result resolves from a spinner to a Success check. It waits until the window scrolls into view, plays once, and shows the finished state under reduced motion or without script.

### Hero Band
The page's lead story on Signal Navy: the Display line (on Home, the tagline), a Lead in Mist, a primary and an on-navy button, and the product window on a Sand tile.

### Closing Card
One inverted navy card on the ground, 12px corners: an H1 headline whose second clause is set light in Mist, a Lead, a primary and an on-navy button. It stays a card so it does not merge into the navy footer.

### Team Tabs
Equal-width tabs over a Hairline rule; the active tab is semibold Ink with a 2px Ink underline. Arrow keys move in reading direction (left is forward in RTL), Home and End jump to the ends, and only the selected tab is in the tab order.

### Navigation and Footer
- **Navbar:** Paper bar with a Hairline bottom border; Label links in Slate, Ink on hover; mega menus are White Card panels with Hairline outlines and Sand icon wells. The primary button carries the only orange.
- **Footer:** Signal Navy with Paper-on-Navy headings and Mist links; the trust strip sits on top with Deep Navy icon wells; the dark logo variant.

## Do's and Don'ts

### Do:
- **Do** build pages from the shared patterns in `app/components/patterns` (live samples at `/_patterns`) and the `site-*` utilities and type scale; never raw palette colors.
- **Do** keep orange to one element per section, and use Ember for orange words and links.
- **Do** quote the tagline exactly, «بی‌گدار به آب و AI نزنید!», with its half-space and exclamation mark, at most once per page.
- **Do** write headlines as two short clauses with a turn, and buttons as two- or three-word verbs.
- **Do** build and check every section in light and dark themes.

### Don't:
- **Don't** use shadows, glows, glass or gradients.
- **Don't** put orange text on light surfaces, or white text on orange.
- **Don't** use navy for anything smaller than a field or a card.
- **Don't** add letter-spacing to Persian.
- **Don't** use emoji, blue-purple gradients, emoji cards or left-border accent cards.
- **Don't** present invented numbers, logos or quotes inside mocks as real.

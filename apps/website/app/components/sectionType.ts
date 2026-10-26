// Shared type and action classes for the GodarAI design system (DESIGN.md).

export const sectionHeading = 'text-balance text-h3 font-semibold text-site-ink md:text-h2';

export const sectionCopy = 'text-lead font-light text-site-ink-soft';

export const sectionSpace = 'py-10 lg:py-20';

const focusRing = 'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus';

// Primary: signal orange with navy text. Use one per section.
export const buttonPrimary = `inline-flex min-h-11 items-center justify-center gap-2 rounded-full bg-site-accent px-4 py-2.5 text-label font-medium text-site-on-accent transition-colors duration-150 hover:bg-site-accent-hover ${focusRing}`;

// Secondary: transparent with a strong hairline and ink text.
export const buttonSecondary = `inline-flex min-h-11 items-center justify-center gap-2 rounded-full border border-site-line-strong px-4 py-2.5 text-label font-medium text-site-ink transition-colors duration-150 hover:bg-site-panel ${focusRing}`;

// Secondary on the navy band: an on-navy outline.
export const buttonOnNavy =
  'inline-flex min-h-11 items-center justify-center gap-2 rounded-full border border-site-night-ink px-4 py-2.5 text-label font-medium text-site-night-ink transition-colors duration-150 hover:bg-site-night-raised focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-night-ink';

// Inline text link with a trailing arrow; the arrow nudges toward reading direction on hover.
export const textLink = `group/link inline-flex w-fit items-center gap-2 rounded-full text-label font-medium text-site-accent-text underline-offset-4 hover:underline ${focusRing}`;

// Icon well: icons stay ink, never orange.
export const iconWell = 'flex shrink-0 items-center justify-center rounded-card bg-site-sand text-site-ink';

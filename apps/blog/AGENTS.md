# Blog AGENTS

Astro site for GodarAI's marketing and technical blog.

## Purpose

- product announcements
- tutorials and deep dives
- company updates
- editorial content only

Keep API reference docs and product documentation out of this app.

## Working Rules

1. Use the monorepo workspace only; do not create a separate workspace or lockfile.
2. Run commands from the repository root with `pnpm --filter @shahrah/blog ...`.
3. Keep content GodarAI-specific instead of starter-template content.
4. Treat this as the canonical blog app for the platform.

## Commands

```bash
pnpm --filter @shahrah/blog dev
pnpm --filter @shahrah/blog build
pnpm --filter @shahrah/blog lint
```

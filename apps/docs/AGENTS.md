# Docs AGENTS

Docusaurus documentation site for Shahrah.

## Purpose

- developer onboarding
- API reference
- SDK guides
- billing and platform documentation

Keep this app focused on docs, not product marketing.

## Working Rules

1. Run commands from the repository root with `pnpm --filter @shahrah/docs ...`.
2. Keep the built-in Docusaurus blog disabled because `apps/blog` is the canonical blog app.
3. Replace starter content with Shahrah-specific documentation as soon as real content is available.
4. Prefer docs that explain current platform behavior over speculative future plans.

## Commands

```bash
pnpm --filter @shahrah/docs dev
pnpm --filter @shahrah/docs build
pnpm --filter @shahrah/docs lint
```

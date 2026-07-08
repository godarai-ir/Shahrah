# Agent Instructions

## Package Manager

- Always use `pnpm` instead of `npm` for dependency, script, and build commands.

## UI / Frontend Work

Before making UI changes:

1. Read `docs/design-system/DESIGN.md`.
2. Read `docs/design-system/COMPONENT_USAGE.md`.
3. Inspect existing `apps/tenant-admin/components/ui` and `apps/tenant-admin/components/app-ui`.
4. Reuse existing components before creating new ones.
5. Use semantic Tailwind tokens.
6. Avoid raw color utilities and arbitrary values.
7. Preserve RTL behavior.
8. Update the design system playground when adding reusable UI patterns.
9. Run available lint/typecheck/build commands before finishing.
10. Summarize any UI debt that remains.

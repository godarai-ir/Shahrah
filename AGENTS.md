# Agent Instructions

## Package Manager

- Always use `pnpm` instead of `npm` for dependency, script, and build commands.

## Go Coding Style and Linting

When writing or changing Go code:

1. Treat the repository-root `.golangci.yml` as the source of truth. All enabled linters and formatters are mandatory; do not weaken, bypass, or disable them to make a change pass.
2. Run `gofmt` and `goimports` on every changed Go file. Keep standard-library, repository-local, and third-party imports in the order produced by `goimports`.
3. Handle every error explicitly. Do not discard errors, use unchecked type assertions, shadow variables, leave unreachable code, or return a different error variable by mistake.
4. Close every acquired resource, including HTTP response bodies in tests. Prefer the package's existing cleanup helper when one is available.
5. Keep control flow simple and idiomatic: avoid unnecessary `else` blocks after terminating branches, preallocate slices when the size is known, and keep cyclomatic complexity below the configured limit.
6. Follow `wsl_v5` whitespace rules. Use blank lines to separate logical operations and place a blank line before a terminal `return` when it follows a multi-line operation block.
7. Write declaration comments as complete sentences ending with a period when comments are required. Use US English spelling.
8. Add `//nolint:<linter> // reason` only when the behavior is intentional, the named linter is specific, and the reason explains why the exception is safe.
9. Before finishing, run `go vet ./...`, `go test ./...`, and `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0` from each changed Go module.

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

# Execution Prompt: Clear Pre-Existing Lint, Format, Architecture, and Type Debt

Use this prompt with a coding agent that has full access to the Shahrah repository.

---

You are a senior engineer clearing four pre-existing quality failures that block
reliable CI gates. These failures were found while delivering credit pricing
phase 1 (PR godarai-ir/Shahrah#60). Fix the causes, not the symptoms. Do not
weaken or disable a linter, rule, or type check to make a stage pass (`AGENTS.md`).

Work on branch `chore/lint-and-types-debt`, cut from `main`. Do one stage per
commit, run that stage's verification before moving on, and stop and report if
a stage needs a decision listed under "Stop and Ask".

## Findings (verified)

| # | Failure | Root cause |
| --- | --- | --- |
| 1 | `golangci-lint`: `registerProxyRoutes` in `services/control-plane/internal/server/routes_proxy.go` has cyclomatic complexity 53, above the limit of 50 | One function registers every gateway route with a nil-check per handler. |
| 2 | `gofmt`: `services/control-plane/internal/billing/pricing_strategy_catalog_tokens.go:227` | The file was committed unformatted. |
| 3 | tenant-admin `validate:architecture`: 14 "feature X must not import feature Y" violations | `users`, `access-keys`, `roles`, `virtual-users`, and `personal-access-tokens` share domain helpers and types by importing each other's feature modules. |
| 4 | tenant-admin `tsc --noEmit`: ~20 errors such as "two different types with this name exist" in `components/ui/command.tsx`, `input-otp.tsx`, `JsonEditor.tsx`, and the guardrail wizard | Two `@types/react` copies are installed. tenant-admin resolves `19.2.14`. `apps/docs` (`^19.2.17`) makes `19.2.17` the hoisted copy, which third-party packages without their own link (`cmdk`, `@uiw/react-codemirror`, `input-otp`) pick up. Separately, `pnpm-workspace.yaml` still lists `sdks/*` packages that were deleted in `eeb158ed`, so every `pnpm install` rewrites `pnpm-lock.yaml`. That is the uncommitted lockfile diff. |

Related: the CI `golangci-lint` job is disabled (`if: ${{ false }}`, comment
"Temporarily disabled"). tenant-admin type checking and architecture validation
are not run in CI at all.

## Stage 1: Go Formatting (#2)

- Run `gofmt -w` on the file. No other edits.
- Verify: `gofmt -l internal cmd tests` prints nothing in `services/control-plane`.

## Stage 2: Split Proxy Route Registration (#1)

- Split `registerProxyRoutes` into cohesive helpers by API family (for example
  OpenAI-compatible core, media/audio, provider-native, and batch/files). Each
  helper must stay well under the gocyclo limit.
- Route paths, methods, middleware order, and nil-handler behavior must stay
  **identical**. Before refactoring, capture the registered route table
  (`app.GetRoutes()`) for a fully populated `Dependencies` and a nil-handler
  `Dependencies`, and add a test that asserts both tables are unchanged.
- Verify: `go vet ./...`, `go test ./internal/server/...`, and
  `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`
  report 0 issues for the whole control-plane module.

## Stage 3: Single `@types/react` Copy and a Stable Lockfile (#4)

1. Remove the four deleted `sdks/*` entries from `pnpm-workspace.yaml`.
2. Unify React types for the whole workspace with root `pnpm.overrides`:
   `@types/react` and `@types/react-dom` pinned to one version each that satisfies
   every app's range (`^19`, `^19.2.17`). Use `19.2.17` and `19.2.3`.
3. Run `pnpm install` and commit the regenerated `pnpm-lock.yaml`. This supersedes
   the uncommitted local lockfile diff.
4. Verify:
   - `node_modules/.pnpm` contains exactly one `@types+react@*`.
   - `pnpm install --frozen-lockfile` succeeds.
   - `tsc --noEmit` passes for tenant-admin, support-admin, website, and
     model-discovery-web, plus `pnpm --filter model-discovery typecheck`.
   - The tenant-admin and docs builds still pass.
5. If any `tsc` errors remain after deduplication, fix them in source as genuine
   type errors. Do not suppress them.

## Stage 4: Remove Cross-Feature Imports (#3)

Move shared domain code to the shared `lib/` layer (the established home for
cross-feature code, as in `lib/api/teams.ts`). Features may import `lib`; `lib`
must not import `features` or `app`.

| Shared code | New home |
| --- | --- |
| `features/roles/role-display.ts`, `effective-permissions.ts`, `contracts.ts`, and the permission-group translation used by users | `lib/access/` |
| `features/access-keys/api.ts` (tenant key list contract) and `key-status.ts` | `lib/api/access-keys.ts`, `lib/access/key-status.ts` |
| Identity helpers in `features/users/member-filters.ts` (`guessDisplayName`, `resolveIdentity`) and the `TenantAdminUser` contract | `lib/access/identity.ts`, `lib/api/tenant-users.ts` |
| Contracts that `access-keys` imports from `virtual-users` and `personal-access-tokens` | `lib/api/` |

Rules:
- Move code; do not duplicate it. Update every importer, including the owning
  feature. Do not leave re-export shims in `features/`.
- Move the existing unit tests with the code.
- Behavior must not change. The existing vitest suites for users, access-keys,
  roles, virtual-users, and personal-access-tokens must pass unchanged, apart
  from import paths.
- Verify: `pnpm run lint` (architecture, source encoding, eslint) passes for
  tenant-admin with 0 violations, along with `tsc --noEmit` and `pnpm test`.

## Stage 5: Keep It Fixed in CI

Found during execution: `ci.yml` runs only on `workflow_dispatch`, and **all four**
jobs (lint, test, javascript-checks, tenant-dom) are switched off with
`if: ${{ false }}`. Turning CI back on is a team decision, so the agent changes
only what is safe to land now and asks about the rest.

- Re-enabling the golangci-lint job as its own comment instructs
  (`if: needs.changes.outputs.go == 'true'`) **requires approval**, together with
  the other disabled jobs and the trigger.
- Add tenant-admin `pnpm --filter @shahrah/tenant-admin exec tsc --noEmit` and
  `pnpm --filter @shahrah/tenant-admin run lint` to the "JavaScript checks" job,
  under the existing tenant scope condition.
- Verify the workflow YAML parses, and that every command added to CI passes
  locally.

## Stop and Ask

- Unifying types requires an `@types/react` version that breaks another app's
  build.
- A shared module cannot move to `lib/` without importing `features/` or `app/`,
  which would need a new shared layer.
- Re-enabling CI lint surfaces findings outside these four failures.

## Report

Per stage: what changed, verification output, and anything deferred.

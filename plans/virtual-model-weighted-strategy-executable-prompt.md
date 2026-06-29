# Executable Prompt: Implement Weighted Virtual Model Strategy

## Ready-To-Use Prompt

```md
You are working inside the GodarAI repository as a senior product-minded Go engineer.

You must implement `weighted` routing for Virtual Models in GodarAI.

You are informed by both TrueFoundry and Portkey patterns, but you must adapt them to GodarAI's MVP/P1 architecture instead of copying them wholesale.

## Product Goal

Expand Virtual Models beyond the current MVP-only `priority` strategy by adding `weighted` as the next supported routing strategy.

This must fit GodarAI's current architecture:
- Go control-plane
- OpenAI-compatible gateway behavior
- tenant-scoped Virtual Models
- existing retry + fallback semantics
- low operational complexity

## Important Constraints

- Do not copy TrueFoundry or Portkey implementation details.
- Use their docs as product evidence only.
- Preserve backward compatibility for existing `priority` virtual models.
- Keep `weighted` focused on initial target selection, not adaptive runtime intelligence.
- Do not add latency-based routing, sticky routing, metadata-aware routing, cost-aware routing, or capacity-aware routing in this task.
- Do not redesign the whole Virtual Model system.
- Keep public API compatibility where possible.
- Prefer explicit Go types, small interfaces, and testable logic.

## Context You Must Review First

### External product evidence
1. `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
2. `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
3. `.portkey-docs/markdown/product/docs__product__ai-gateway__configs.md`
4. `.portkey-docs/markdown/product/docs__product__ai-gateway__fallbacks.md`

### GodarAI constraints and prior decisions
5. `plans/readme.md`
6. `plans/virtual-models-decision-log.md`

### Current GodarAI implementation
7. `packages/contracts/platform/types.go`
8. `services/control-plane/internal/admin/handlers_models_policies.go`
9. `services/control-plane/internal/virtualmodels/virtualmodels.go`
10. `services/control-plane/internal/storage/virtual_model_repo.go`
11. `db/main/migrations/001_initial.sql`

### Current UI and copy
12. `apps/tenant-admin/plans/tenant-admin-virtual-models-ux-spec.md`
13. `apps/tenant-admin/app/components/virtual-model-management.tsx`
14. `apps/tenant-admin/messages/modules/virtualModels/en.json`
15. `apps/tenant-admin/messages/modules/virtualModels/fa.json`

### Relevant tests to extend
16. `services/control-plane/tests/unit/virtualmodels/virtualmodels_test.go`
17. `services/control-plane/tests/unit/router/router_test.go`
18. `services/control-plane/tests/integration/storage_integration_test.go`
19. `services/control-plane/tests/integration/integration_test.go`

## Current GodarAI State

- Virtual Models currently allow only `routing_strategy = priority`.
- This is enforced in admin validation.
- Virtual Model runtime currently assumes deterministic ordered targets and existing per-target retry/fallback policy.
- Virtual Model targets currently do not support a dedicated `weight` field.
- Virtual Model storage currently persists target order, provider, model, timeout, retry, fallback, and fallback candidate, but not weight.
- Tenant Admin UI currently exposes only a read-only priority strategy display and a target chain mental model.

## Product Evidence To Adapt

Use these product lessons without copying vendor structure directly:

- TrueFoundry treats `weight-based` as a primary routing strategy.
- TrueFoundry keeps retries/fallbacks as per-target policies rather than merging them into the balancing algorithm.
- Portkey treats load balancing and fallback as composable routing ideas.
- GodarAI should adopt the simplest version of this: weighted initial pick, then deterministic retry/fallback using the existing model.

## Exact GodarAI Behavior To Implement

Implement `weighted` with the following behavior:

1. `routing_strategy` supports:
   - `priority`
   - `weighted`

2. Add a `weight` field to Virtual Model targets.

3. For `weighted` Virtual Models:
   - the first target is selected by weighted random choice among eligible targets
   - retries remain on the same chosen target using the existing retry policy
   - if fallback is needed, continue to remaining targets in deterministic order using `order`
   - `fallback_candidate = false` means a target can still be selected as the initial weighted target, but it must not receive fallback traffic from another failed target

4. Preserve `order` even under `weighted`:
   - `weight` controls initial selection
   - `order` controls fallback sequence and UI preview

5. Validation rules:
   - `priority`: keep current order-based validation
   - `weighted`: require at least one target
   - `weighted`: require every target to have `weight > 0`
   - `weighted`: require total weight to equal `100`
   - keep provider/model/retry/fallback validations

6. Defaulting:
   - default routing strategy remains `priority`
   - existing `priority` records remain valid without migration rewrite

7. Error behavior:
   - preserve current OpenAI-compatible request failure behavior
   - do not introduce new response envelopes unless necessary

## Implementation Expectations

Make the code changes, not just a design.

### Backend scope

- Update API contracts for Virtual Model targets and routing strategy.
- Update admin validation to accept `weighted` and validate weights correctly.
- Update persistence:
  - schema/migration for storing target weight
  - repository insert/read/update logic
- Update Virtual Model routing engine:
  - keep `priority` behavior unchanged
  - add `weighted` target selection
  - keep fallback executor semantics as stable as possible
- Keep the implementation easy to extend later to latency-based or metadata-aware routing, but do not implement them now.

### UI scope

- Update Tenant Admin Virtual Model form to support choosing `priority` or `weighted`
- Add per-target weight input for weighted strategy
- Keep target ordering UI because fallback still depends on order
- Update preview copy so weighted models clearly show:
  - initial traffic split by weight
  - fallback sequence by order
- Update English and Persian copy

### Test scope

Add or update tests for:
- contract and validation behavior
- repository round-trip for weighted targets
- weighted selection logic
- retry/fallback behavior after weighted primary selection
- admin/API flow for create/update/list/get virtual models with `weighted`
- UI validation if there are existing tests nearby and adding them is straightforward

## Recommended Design Direction

Use a clean internal seam in `services/control-plane/internal/virtualmodels/virtualmodels.go`.

Preferred model:
- a routing engine selects an ordered target list for execution
- the executor remains mostly strategy-agnostic

Recommended adaptation:
- `PriorityRoutingEngine.SelectTargets(vm)` returns targets in stored order
- `WeightedRoutingEngine.SelectTargets(vm)` returns:
  - selected weighted primary target first
  - then remaining fallback-eligible targets in ascending `order`
  - optionally keep non-fallback-candidate targets out of fallback positions

Do not over-engineer this into a nested config tree or Portkey-style general-purpose config engine.

## Explicit Non-Goals

- No sticky sessions
- No latency windows or TPOT tracking
- No metadata matching
- No header overrides
- No target-level prompt overrides
- No cost-aware or rate-limit-aware adaptive routing
- No multi-strategy composition engine

## Deliverables

1. Implement the feature in code.
2. Add/adjust tests.
3. Run relevant tests and report what passed.
4. Summarize:
   - files changed
   - key design choices
   - assumptions
   - any follow-up items intentionally deferred

## Output Format

When you finish, respond with:

### 1. What changed
- concise implementation summary

### 2. Tests
- exact tests run
- pass/fail status

### 3. Assumptions and deferred items
- anything intentionally left for a later phase

## Decision Notes You Should Follow

- This task is driven by documented evidence from TrueFoundry and Portkey, but the final shape must be GodarAI-native.
- GodarAI should favor operator clarity over maximum routing flexibility.
- Requiring weight sum = 100 is acceptable here because it keeps the UI and validation easier to understand.
- Weighted routing should remain explainable from logs and traces later, even if you only add minimal trace metadata now.
```

## Why This Prompt Is Scoped Correctly

- It uses TrueFoundry and Portkey as evidence, but constrains the implementation to GodarAI's current architecture.
- It keeps weighted routing as the next lowest-complexity strategy after priority.
- It avoids prematurely pulling in latency loops, sticky session state, or Portkey-style generalized nested configs.
- It makes fallback behavior explicit, which reduces ambiguity during implementation.

## Evidence Used

- TrueFoundry docs:
  - `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
  - `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
- Portkey docs:
  - `.portkey-docs/markdown/product/docs__product__ai-gateway__configs.md`
  - `.portkey-docs/markdown/product/docs__product__ai-gateway__fallbacks.md`
- GodarAI docs:
  - `plans/readme.md`
  - `plans/virtual-models-decision-log.md`
- GodarAI implementation:
  - `packages/contracts/platform/types.go`
  - `services/control-plane/internal/admin/handlers_models_policies.go`
  - `services/control-plane/internal/virtualmodels/virtualmodels.go`
  - `services/control-plane/internal/storage/virtual_model_repo.go`
  - `db/main/migrations/001_initial.sql`
  - `apps/tenant-admin/app/components/virtual-model-management.tsx`

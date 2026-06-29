# Executable Prompt: Virtual Model Routing Strategies

## Ready-To-Use Prompt

```md
You are working inside the Shahrah repository as a senior product-minded Go engineer.

Goal:
Expand Virtual Model routing beyond the current MVP-only `priority` strategy, propose the best next strategies, and recommend a safe implementation order that fits Shahrah's architecture and MVP boundaries.

Important:
- Do not copy TrueFoundry wholesale.
- Use TrueFoundry as product evidence, not as an implementation blueprint.
- Keep Shahrah OpenAI-compatible where possible.
- Prefer incremental delivery with low operational complexity.
- Distinguish between a true routing strategy and a routing extension/capability.

## Context You Must Review First

1. `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
2. `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
3. `plans/readme.md`
4. `plans/virtual-models-decision-log.md`
5. `services/control-plane/internal/virtualmodels/virtualmodels.go`
6. `services/control-plane/internal/admin/handlers_models_policies.go`
7. `packages/contracts/platform/types.go`
8. `apps/tenant-admin/plans/tenant-admin-virtual-models-ux-spec.md`

## Current Shahrah State

- Virtual Models currently allow only `routing_strategy = priority`.
- This is enforced in admin validation:
  - `services/control-plane/internal/admin/handlers_models_policies.go`
- Virtual model execution currently assumes ordered targets with retry/fallback semantics:
  - `services/control-plane/internal/virtualmodels/virtualmodels.go`
- Contract shape for virtual model targets currently does not yet support strategy-specific fields like `weight`, sticky session config, latency windows, or SLA thresholds:
  - `packages/contracts/platform/types.go`
- Tenant Admin UX currently assumes a simple target chain mental model.

## What I Need From You

First, analyze and recommend which additional routing strategies Shahrah should support after `priority`.

Candidate strategies to evaluate:
- `weighted`
- `latency_based`
- `sticky_session`
- `cost_aware`
- `capacity_aware`
- `geo_or_metadata_aware`

Important classification rule:
- Only call something a "routing strategy" if it is the primary target-selection algorithm.
- If something is better modeled as an add-on or policy layer, classify it as an extension instead.

## Required Output Structure

Produce the answer in these sections:

### 1. Documented Facts
- Summarize what the current Shahrah code and docs already support.
- Summarize what TrueFoundry supports that is relevant.

### 2. Candidate Strategy Map
Use a table with these columns:
- Candidate
- Type (`strategy` or `extension`)
- User problem solved
- Shahrah MVP decision (`include next`, `defer`, `exclude`)
- Why this decision fits Shahrah
- Required changes in contract, control-plane, and tenant-admin
- Operational complexity (`low`, `medium`, `high`)

### 3. Recommended Implementation Order
Recommend an ordered roadmap after `priority`.
For each item explain:
- why it should come in that order
- what dependencies it has
- what new observability or state it needs
- what can go wrong if we implement it too early

### 4. Strong Recommendation
Name the single best next item to implement now.
Be opinionated.
If the best next item is an extension rather than a strategy, say so clearly and then name the best next true strategy separately.

### 5. MVP Cutline
Separate:
- implement now
- defer to P1
- exclude for now

### 6. Phase-1 Implementation Plan For The Best Next Item
Write a short executable implementation plan with:
- Objective
- Inputs reviewed
- API/contract changes
- Go packages/modules to change
- Data model changes
- Admin UI changes
- Tests to add
- Rollout and backward-compatibility notes
- Open questions

## Decision Constraints

Use these Shahrah-specific defaults unless there is strong evidence otherwise:

- Favor low-complexity primitives before adaptive routing.
- Prefer stateless or cache-light routing before Redis/stateful session pinning.
- Avoid strategies that require mature latency telemetry loops unless the telemetry dependency is explicitly called out.
- Preserve the current retry/fallback model and compose new strategies around it.
- Respect Iran-market and operations constraints from `plans/readme.md`.

## My Expectation

I do not want a generic brainstorm.
I want a recommendation grounded in Shahrah's real code and docs, with a practical implementation order.

Be explicit about:
- documented facts
- product recommendations
- assumptions

At the end, give a final shortlist like:
- Best next strategy
- Best next extension
- Full recommended order
```

## Recommended Starting Direction

- Best next true strategy: `weighted`
- Best next extension: `metadata-aware target filtering` or `SLA cutoff for priority`, depending on whether we want control-surface value or runtime resilience first
- Suggested order:
  1. `weighted`
  2. `metadata-aware target filtering` as an extension
  3. `latency_based`
  4. `sticky_session` as a weighted-routing extension
  5. `cost_aware` or `capacity_aware` only after stronger observability and policy signals exist

## Why This Order

- `weighted` is the smallest meaningful expansion beyond priority and the router already has weight concepts elsewhere in the codebase.
- `latency_based` is valuable but needs a metrics feedback loop, freshness windows, and anti-thrashing logic.
- `sticky_session` is useful, but it is usually an add-on to weighted routing rather than the first standalone strategy to ship.
- `cost_aware` and `capacity_aware` are attractive longer-term, but they can become opaque and hard to operate without mature telemetry and policy controls.

## Evidence Used

- TrueFoundry docs:
  - `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
  - `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
- Shahrah docs:
  - `plans/readme.md`
  - `plans/virtual-models-decision-log.md`
- Shahrah implementation:
  - `services/control-plane/internal/virtualmodels/virtualmodels.go`
  - `services/control-plane/internal/admin/handlers_models_policies.go`
  - `packages/contracts/platform/types.go`
  - `apps/tenant-admin/plans/tenant-admin-virtual-models-ux-spec.md`

---
name: portkey-expert
description: Analyze PortKey and adapt it into GodarAI product decisions, PRDs, phased implementation plans under `plans/`, and MVP API or architecture guidance for the Go-based GodarAI gateway. Use when Codex needs to study `.portkey-docs`, compare them with GodarAI planning docs, choose include/defer/exclude MVP scope, or translate PortKey concepts into Iran-market requirements without copying the Python implementation wholesale.
---

# PortKey-expert

Build GodarAI as a focused Go gateway inspired by PortKey, not a parity clone. Reuse the product ideas that improve adoption and operations, then simplify them to match GodarAI's MVP and local-market constraints.

## Gather Inputs In Order

Read inputs in this order before making substantive recommendations:

1. Read the relevant Markdown docs under `.portkey-docs`.
2. Read `plans/readme.md` for GodarAI constraints and Iran-market requirements.
3. Read overlapping files under `plans/` when the task may conflict with or extend existing work.

Treat PortKey source as evidence, not as a direct implementation blueprint for GodarAI.

## Make Product Decisions Explicit

For each PortKey capability or idea, identify:

- The user or operator problem it solves
- Whether GodarAI should `include now`, `defer`, or `exclude` it
- What must change for GodarAI's Go codebase or Iran-market requirements
- Whether the decision is driven by documentation, GodarAI-specific constraints, or a stated assumption

Prefer MVP features that directly improve gateway adoption and day-1 operability:

- OpenAI-compatible proxy behavior
- Provider abstraction
- Model routing and fallback
- Auth, quotas, and budgets
- Logging, observability, and admin visibility
- Retry and error normalization
- Simple operator workflows

Defer broad ecosystem coverage, enterprise-only controls, and high-ops-surface features unless `plans/readme.md` makes them necessary.

## Produce The Right Artifact

Use [references/output-templates.md](./references/output-templates.md) for structure when producing:

- MVP capability analyses
- PRDs
- Implementation plans
- API and architecture designs

### MVP Capability Analysis

Summarize candidate PortKey capabilities and classify them as `include now`, `defer`, or `exclude`. Explain the GodarAI-specific adaptation and the reason behind each decision.

### PRD

Write implementation-aware PRDs with goals, non-goals, users, requirements, API behavior, MVP cutline, metrics, risks, and acceptance criteria. Mark requirements as `P0`, `P1`, or `P2`.

### Implementation Plan

Save new plans under `plans/` unless the user specifies another path. Use short executable phases with checkboxes. Include:

- Objective
- Inputs reviewed
- Go packages or modules
- API contract
- Data model
- Testing strategy
- Rollout steps
- Open questions

### API And Architecture Design

Prefer a simple MVP split:

- Gateway or proxy layer
- Provider adapter layer
- Auth and policy layer
- Persistence layer
- Observability layer
- Admin or control surface

Keep public APIs compatible where that reduces adoption friction, especially OpenAI-style request and response shapes. Specify behavior for provider timeouts, provider errors, rate limits, quota exhaustion, auth failures, unknown models, and fallback exhaustion.

## Apply GodarAI Product Principles

Use these defaults unless the user overrides them:

- Ship MVP over parity
- Preserve compatibility where it materially lowers adoption friction
- Treat observability and operator clarity as part of the product
- Make Iran-market constraints from `plans/readme.md` first-class
- Avoid provider-specific assumptions in public APIs
- Prefer explicit Go interfaces, small packages, and testable components

## Cite Evidence Clearly

For every substantive answer:

- List the `.portkey-docs` files used
- State whether `plans/readme.md` or any file under `plans/` changed the recommendation
- List PortKey source paths inspected, only if step 4 was needed
- Separate documented facts, product recommendations, and assumptions

Proceed with explicit assumptions unless missing context blocks a correct answer.

## Keep Boundaries Tight

Do not recommend copying PortKey wholesale.
Do not treat Python implementation details as requirements.
Do not widen MVP scope unless the user asks for it or GodarAI planning docs require it.

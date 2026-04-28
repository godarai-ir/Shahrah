# Output Templates for LLMLite Expert

## MVP Capability Analysis

Use this structure when the user asks which LiteLLM ideas should become Shahrah MVP capabilities.

```markdown
# Shahrah MVP Capability Analysis

## Decision Summary

State the recommended MVP scope in 3-7 bullets.

## LiteLLM Capability Map

| LiteLLM capability | User/job value | MVP decision | Shahrah adaptation | Rationale |
|---|---:|---|---|---|
| capability | high/medium/low | include/defer/exclude | adaptation | evidence from docs |

## MVP Scope

### Include Now

List capabilities with why they matter, core flows, and acceptance criteria.

### Defer

List useful but non-MVP capabilities and the trigger for adding them later.

### Exclude

List capabilities not aligned with Shahrah's initial market or product goals.

## Iran-Specific Additions

Summarize additions from `plans/readme.md`, how they change the MVP, and dependencies.

## Risks and Open Questions

Separate product, technical, operational, legal/compliance, and market risks.
```

## PRD

Use this structure when writing or revising a PRD.

```markdown
# PRD: [Feature or Product Area]

## Context

Explain the customer problem, Shahrah product goal, and relevant LiteLLM inspiration.

## Goals

- Goal 1
- Goal 2

## Non-Goals

- Non-goal 1

## Users and Use Cases

| User | Use case | Success signal |
|---|---|---|

## Requirements

### Functional Requirements

Use numbered requirements with priority: P0, P1, P2.

### Non-Functional Requirements

Cover reliability, latency, security, observability, scalability, data retention, and operational constraints.

## UX/API Behavior

Document dashboard behavior, API behavior, errors, and edge cases.

## MVP Cutline

State what ships in MVP and what waits.

## Metrics

Define activation, usage, reliability, cost, and support metrics.

## Dependencies

List codebase, infrastructure, documentation, and business dependencies.

## Risks

List risks and mitigations.

## Acceptance Criteria

Use concrete testable bullets.
```

## Implementation Plan

Save implementation plans under `plans/` unless the user asks for a different location.

```markdown
# Implementation Plan: [Feature]

## Objective

One paragraph describing the feature and why it is in the MVP.

## Inputs Reviewed

- LiteLLM docs: paths reviewed
- LiteLLM source: paths reviewed, only if needed
- Shahrah plans: paths reviewed

## Architecture

Describe the Go components, data flow, storage, and API boundaries.

## Work Breakdown

### Phase 1: Foundation

- [ ] Task with expected files/modules
- [ ] Task with acceptance criteria

### Phase 2: Core Behavior

- [ ] Task

### Phase 3: Observability and Hardening

- [ ] Task

## API Contract

Include endpoints, request/response examples, error model, and compatibility notes.

## Data Model

Include tables, key fields, indexes, and retention expectations.

## Testing Strategy

List unit, integration, load, security, and migration tests.

## Rollout Plan

List flags, migration steps, operational checks, and rollback plan.

## Open Questions

List unresolved items requiring product or engineering decision.
```

## API and Architecture Design

```markdown
# Shahrah API and Architecture Design: [Area]

## Design Goals

## Key Constraints

## Proposed Architecture

## Public API

## Internal Go Interfaces

## Routing and Provider Abstraction

## Auth, Quotas, and Policy

## Observability

## Failure Modes

## MVP vs Later
```

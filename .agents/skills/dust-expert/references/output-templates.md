# Output Templates for Dust Expert

Pick the template that matches the request. Trim sections that add nothing; never pad.

## 1. Idea Teardown (one Dust feature → GodarAI)

Use when the user asks "how does Dust do X" or "should we build X like Dust".

```markdown
# [Feature]: from Dust to GodarAI

## Recommendation
One paragraph: verdict (adopt / adapt / defer / skip), the GodarAI version in one sentence, and why.

## The job
Who has the problem, in their words, and how they cope today in GodarChat / tenant admin.

## How Dust designed it
- What the user sees (flow in 3–6 steps)
- Design intent and trade-offs
- Doc links

## What GodarAI already has
Existing LibreChat / Shahrah capability, with file paths.

## GodarAI design
- Flow (numbered steps) or ASCII wireframe
- Key states: empty, loading/running, success, error, no permission, out of credits
- Persian UI copy for the main actions (with English gloss)
- Admin controls and defaults
- Credit impact

## MVP cut vs. later

## Risks and open questions
Label assumptions.
```

## 2. Opportunity Map (many Dust ideas → prioritized list)

Use when the request is broad: "bring Dust ideas to GodarAI", "what should we steal from Dust".

```markdown
# Dust → GodarAI Opportunity Map

## Top picks
3–5 bullets: the ideas to do first and the outcome each moves (activation, team adoption, trust, revenue/credits).

## Map

| Dust idea | Job it solves | GodarAI surface | Today in GodarAI | Verdict | Effort | Impact |
|---|---|---|---|---|---|---|

## Top picks in detail
Short design sketch per pick (flow, states, credit impact).

## Deferred / skipped
One line each with the trigger that would bring it back.

## Assumptions and questions
```

## 3. Design Brief

Use before UI work on a single feature.

```markdown
# Design Brief: [Feature]

## Problem and user
## Inspiration (Dust) and what we change
## Principles for this feature (2–4)
## User flows
## Screens and states
## Content and copy (fa / en)
## Accessibility and RTL notes
## Success metrics
## Out of scope
```

## 4. PRD

Use when the user wants a PRD. Match the style of existing PRDs in `Shahrah/docs/` and `Shahrah/plans/` (Persian PRDs are fine if the user writes in Persian).

```markdown
# PRD: [Feature]

## Summary and problem
## Goals / Non-goals / Success metrics
## Users and use cases
| User | Use case | Success signal |
## Requirements (P0 / P1 / P2, numbered)
## UX behavior (flows, states, errors, permissions, credits)
## Admin controls
## Dependencies (LibreChat, control plane, gateway, billing)
## Rollout
## Risks
## Acceptance criteria
## Dust references
```

## 5. UX Critique (GodarChat / tenant admin through a Dust lens)

```markdown
# UX Critique: [Screen or flow]

## Summary — the 3 changes that matter most
## Findings
| # | Where | Issue | Dust pattern that addresses it | Suggested change | Severity |
## Quick wins vs. bigger bets
```

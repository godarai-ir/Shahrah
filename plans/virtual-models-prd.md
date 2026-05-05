# Virtual Models PRD (MVP)

## Documented Facts
- TrueFoundry Virtual Models define a stable alias over multiple provider targets with retry/fallback and routing.
- Portkey Configs/Fallbacks support model/provider failover with status-code triggers.
- Shahrah control-plane currently routes by model name to deployments and returns OpenAI-compatible responses.

## Recommendations
- Add tenant-scoped `virtual_models` as first-class alias resolution before direct router selection.
- Keep MVP strategy to `priority` only.
- Keep `/v1/chat/completions` and `/v1/completions` unchanged; only enhance `model` resolution.
- Add admin CRUD under `/api/v1/virtual-models`.

## Assumptions
- Existing billing/rate-limits continue against requested model alias in MVP.
- Streaming fallback for virtual models is deferred; sync path is P0.

## Goals
- Resilient provider failover without client-side model map logic.
- Stable virtual alias with tenant isolation.
- Backward compatibility for direct model calls.

## Non-Goals
- Weighted/latency/sticky routing.
- Metadata-based routing and header overrides.
- Recursive virtual model targets.

## Requirements
- P0: alias CRUD, priority target order, per-target retry/fallback policy, normalized fallback exhausted error, observability fields.
- P1: weighted/latency routing, richer policy expressions.
- P2: metadata rules and advanced traffic shaping.

## Acceptance Criteria
- Virtual alias request succeeds when primary fails and fallback succeeds.
- Virtual alias request returns normalized terminal error when all targets fail.
- Direct model route remains unchanged.
- `/v1/models` includes virtual aliases.

## Sources Used
- `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
- `.portkey-docs/markdown/product/docs__product__ai-gateway__configs.md`
- `.portkey-docs/markdown/product/docs__product__ai-gateway__fallbacks.md`

## Impact of `plans/README.md`
- Reinforced OpenAI-compatible surface and MVP-first scope.

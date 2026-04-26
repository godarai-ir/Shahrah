# Phase 3: Provider-Model Catalogue - Implementation Plan

> **Status:** Draft  
> **Created:** 2026-04-26  
> **Priority:** P0 - Foundational control surface for multi-provider routing, billing, and access control

---

## Objective

Build a versioned Provider-Model Catalogue for Shahrah that becomes the source of truth for:

- Which providers and models are available
- What capabilities each model supports
- What pricing metadata is used for billing and estimation
- Which tenants/teams/keys can access which offerings
- How routing should choose and fallback between offerings

This plan is product + architecture focused and intentionally does not require code in this phase.

---

## Inputs Reviewed

- LiteLLM docs:
  - `.llmLite_docs/proxy/model_management.md`
  - `.llmLite_docs/proxy/model_discovery.md`
  - `.llmLite_docs/proxy/sync_models_github.md`
  - `.llmLite_docs/provider_registration/add_model_pricing.md`
  - `.llmLite_docs/completion/model_alias.md`
- Shahrah plans:
  - `plans/README.md`
  - `plans/04_microservices_architecture.md`
  - `plans/09_shahrah_features.md`
  - `plans/phase1-core-gateway-stabilization.md`
  - `plans/phase2-liteLLM-provider-support.md`
- LiteLLM source (used to remove ambiguity on runtime behavior):
  - `d:/source/litellm/litellm/litellm_core_utils/get_llm_provider_logic.py`
  - `d:/source/litellm/litellm/litellm_core_utils/get_model_cost_map.py`
  - `d:/source/litellm/litellm/proxy/proxy_server.py`
  - `d:/source/litellm/litellm/proxy/auth/model_checks.py`

---

## Architecture

### Proposed Shahrah Components

1. **Catalogue Domain (Control Plane)**
2. Owns provider/model metadata, capabilities, pricing profiles, aliases, and lifecycle state.
3. Exposes versioned read APIs for router/proxy and admin APIs for curation.

4. **Catalogue Sync Service (Control Plane)**
5. Imports updates from configured sources (LiteLLM map, provider discovery, manual admin input).
6. Writes changes to a draft version, runs validations, and prepares a publishable diff.

7. **Catalogue Resolver (Runtime Path)**
8. Resolves requested model name -> alias -> canonical model -> allowed provider offerings.
9. Applies tenant/team/key policy filters and returns routing candidates to router.

10. **Catalogue Snapshot Cache**
11. Immutable published snapshot in Redis/in-memory for low-latency lookups.
12. Supports atomic snapshot switch, rollback, and stale snapshot fallback.

### Data Flow

1. Admin/sync updates draft catalogue.
2. Validation and policy checks run on draft.
3. Draft is published as `catalog_version = N`.
4. Runtime caches refresh to version `N`.
5. Proxy/router, `/v1/models`, billing estimator, and admin UI all read from the same version.

---

## Work Breakdown

### Phase 1: Domain Model and Governance (P0)

- [ ] Define conceptual entities and ownership boundaries:
  - `provider`
  - `provider_endpoint`
  - `model_canonical`
  - `provider_model_offering`
  - `model_alias`
  - `capability_profile`
  - `pricing_profile`
  - `catalog_version`
  - `tenant_catalog_policy`
  - `routing_hints`
- [ ] Define lifecycle states and governance:
  - `draft`, `validated`, `published`, `deprecated`, `sunset`
  - approval gate before publish (at least one reviewer rule)
- [ ] Define immutable publish model:
  - published versions are append-only
  - rollback points must stay queryable and restorable
- [ ] Define acceptance criteria for phase:
  - every routed request can be traced to one `catalog_version`
  - every listing endpoint can return model/provider metadata from same version

### Phase 2: API Contract and Runtime Read Path (P0)

- [ ] Define runtime read APIs:
  - `GET /v1/models` (OpenAI-compatible list)
  - `GET /v1/models/{model_id}` (model details)
  - internal resolver API for router candidate lookup
- [ ] Define admin APIs:
  - create/update draft entries
  - run validation
  - publish draft
  - rollback to version
  - diff versions
- [ ] Define alias and wildcard behavior:
  - user alias -> canonical
  - provider wildcard (example: `openai/*`)
  - per-tenant filtered listing
- [ ] Define resolver policy order:
  - alias resolution
  - tenant/team/key access policy
  - capability match (tools/vision/streaming/etc.)
  - route hints (priority/weight/cost class)

### Phase 3: Sync and Curation Pipeline (P1)

- [ ] Define source adapters:
  - LiteLLM-style model cost map import
  - optional provider endpoint discovery
  - manual admin curation
- [ ] Define validation gates:
  - schema validation
  - provider/model key uniqueness
  - pricing sanity checks
  - required capability defaults
  - lifecycle consistency (`deprecated` + replacement target)
- [ ] Define scheduled sync:
  - manual reload endpoint
  - periodic job
  - source status endpoint
- [ ] Define conflict policy:
  - source precedence
  - protected manual overrides
  - deterministic merge rules

### Phase 4: Observability and Operational Hardening (P1)

- [ ] Define catalogue metrics:
  - version age
  - publish frequency
  - sync success/failure
  - validation failure categories
  - resolver cache hit ratio
- [ ] Define audit requirements:
  - who changed what
  - source of change
  - before/after snapshot fingerprints
- [ ] Define failure behavior:
  - source fetch fails -> keep last published version
  - invalid draft -> block publish
  - cache refresh fails -> safe fallback to previous snapshot
- [ ] Define runbooks:
  - emergency rollback
  - bad pricing correction
  - alias conflict hotfix

---

## API Contract

### Public (OpenAI-Compatible)

1. `GET /v1/models`
2. Returns filtered model list available to requesting key under current `catalog_version`.

3. `GET /v1/models/{model_id}`
4. Returns model details including canonical model metadata and capability flags.

### Admin (Control APIs)

1. `POST /api/v1/catalog/drafts`
2. Create or upsert draft entities.

3. `POST /api/v1/catalog/drafts/{draft_id}/validate`
4. Run validation suite and return actionable failures.

5. `POST /api/v1/catalog/drafts/{draft_id}/publish`
6. Publish immutable `catalog_version`.

7. `POST /api/v1/catalog/versions/{version}/rollback`
8. Set active version to selected published version.

9. `GET /api/v1/catalog/versions/{from}/diff/{to}`
10. Return structured diff (provider, model, pricing, capability, policy changes).

### Error Model (Normalized)

- `invalid_catalog_state`
- `catalog_validation_failed`
- `catalog_version_not_found`
- `catalog_publish_conflict`
- `catalog_sync_failed`

---

## Data Model

### Core Entities

1. `providers`
2. `provider_endpoints`
3. `models_canonical`
4. `provider_model_offerings`
5. `model_aliases`
6. `capability_profiles`
7. `pricing_profiles`
8. `tenant_catalog_policies`
9. `routing_hints`
10. `catalog_versions`
11. `catalog_version_items` (snapshot linkage table)
12. `catalog_sync_jobs`

### Key Design Rules

1. Catalogue is versioned; runtime reads a single active version.
2. Offerings are the join between provider and canonical model.
3. Alias and wildcard resolution is explicit and deterministic.
4. Tenant policy is additive with explicit deny capability.
5. Pricing is stored in USD with conversion at billing time using Shahrah exchange-rate service.

### Suggested Indexes

- `(catalog_version, model_name)`
- `(catalog_version, provider, provider_model_name)`
- `(tenant_id, catalog_version, model_name)`
- `(alias, catalog_version)`
- `(catalog_version, lifecycle_state)`

---

## Testing Strategy

### Unit Tests

- alias resolution precedence
- wildcard expansion behavior
- capability filter logic
- pricing lookup fallback rules
- lifecycle state handling

### Integration Tests

- publish and rollback version flow
- resolver lookup under tenant policies
- `/v1/models` and `/v1/models/{id}` consistency
- router candidate generation from catalogue

### End-to-End Tests

- request path uses expected `catalog_version`
- billing estimation reads same model pricing profile
- access control denies disallowed provider/model combinations

### Non-Functional Tests

- high-QPS resolver read performance
- cache refresh under concurrent traffic
- sync failure resilience and stale-version safety

---

## Rollout Plan

1. **Stage 0: Shadow Mode**
2. Build catalogue and resolver in read-only shadow mode beside existing model/deployment config.

3. **Stage 1: Read Path Cutover**
4. Move `/v1/models` and model-detail endpoints to catalogue-backed reads.

5. **Stage 2: Routing Candidate Cutover**
6. Router gets candidates from catalogue resolver, with fallback to legacy deployment list.

7. **Stage 3: Billing and Policy Alignment**
8. Billing estimator and policy checks read pricing/capabilities from same catalogue version.

9. **Stage 4: Legacy Removal**
10. Remove duplicate model metadata paths after stability period and metrics signoff.

### Rollback

- Roll back active `catalog_version` to last known good.
- Keep resolver fallback to previous in-memory snapshot if store/cache unhealthy.

---

## Open Questions

1. Should tenant-specific pricing overrides be allowed in MVP, or only global pricing?
2. Should provider health signals be embedded in catalogue or kept fully runtime-only?
3. What is the exact ownership boundary between model lifecycle (`deprecated/sunset`) and product messaging/UI?
4. Should we support per-tenant alias maps in MVP or defer to platform-wide aliases first?
5. What is the required freshness SLA for external source sync (hourly, 6-hour, daily)?

---

## Suggested Delivery Sequence

1. Phase 1 + Phase 2 (foundation and API contract)
2. Phase 3 (sync and curation)
3. Phase 4 (observability and hardening)

Estimated delivery for full scope: **2-4 weeks**, depending on whether sync adapters are limited to one source in MVP.


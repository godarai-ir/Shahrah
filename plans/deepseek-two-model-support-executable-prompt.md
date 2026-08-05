# Executable Prompt: DeepSeek Two-Model Support

## Ready-to-use prompt

```md
You are working inside the GodarAI repository as a senior product-minded Go platform engineer. Implement production-gated direct DeepSeek support end to end. Do not stop at an analysis or a plan: make the code, catalog, migrations, tests, documentation, readiness tooling, and rollout controls complete within the approved scope.

## Objective

Add exactly these two canonical models from the direct DeepSeek provider to GodarAI:

- `deepseek-v4-flash`
- `deepseek-v4-pro`

Use the canonical provider slug `deepseek`. Keep Thinking/Non-Thinking as modes of each model; do not create fake model IDs such as `deepseek-v4-pro-thinking`.

The public GodarAI model identifier must continue to follow the existing runtime FQN contract:

`deepseek:<provider-account>:<model-id>`

The delivery is not production-ready merely because a generic OpenAI-compatible smoke test passes. Both models must remain behind provider/model release gates until live compatibility, billing reconciliation, privacy, and operational evidence are attached.

## Mandatory first steps

1. Read the repository-root `AGENTS.md` and obey it throughout the task.
2. Run `git status --short` and preserve all existing user changes. Never overwrite, reset, or reformat unrelated work. The worktree may already contain in-progress Moonshot and xAI changes; integrate carefully with the current state rather than assuming a clean branch.
3. Use `$deepseek-api` and read completely:
   - `C:/Users/Delpak/.codex/skills/deepseek-api/SKILL.md`
   - `C:/Users/Delpak/.codex/skills/deepseek-api/references/official-source-map.md`
4. Run the skill's official-source discovery helper:
   - `python C:/Users/Delpak/.codex/skills/deepseek-api/scripts/discover_official_docs.py`
5. Re-open the live official DeepSeek documentation listed below. Provider facts must come only from official DeepSeek sources. Do not use LiteLLM, OpenRouter, DeepInfra, model directories, or memory as API authority.
6. Read Shahrah product and implementation context:
   - `.llmLite_docs/providers/deepseek.md` only as LiteLLM product inspiration, never as current DeepSeek truth;
   - `plans/README.md`;
   - `plans/model-catalog-experience-prd.md`;
   - `plans/platform-managed-provider-credentials-provider-matrix.md`;
   - `plans/moonshot-four-model-production-readiness.md`;
   - relevant current Moonshot and xAI catalog, adapter, runtime-profile, billing, readiness, and documentation files.
7. Inspect the actual current schemas, import pipeline, provider registry, operation mapping, credential manager, billing service, and tests before choosing file changes. Reuse current abstractions; do not copy another provider mechanically.

## Official sources that must be reverified

- Documentation inventory: https://api-docs.deepseek.com/sitemap.xml
- API authentication: https://api-docs.deepseek.com/api/deepseek-api
- Models and pricing: https://api-docs.deepseek.com/quick_start/pricing
- Model discovery: https://api-docs.deepseek.com/api/list-models
- Chat Completions: https://api-docs.deepseek.com/api/create-chat-completion
- Thinking mode: https://api-docs.deepseek.com/guides/thinking_mode
- Tool calls: https://api-docs.deepseek.com/guides/tool_calls
- JSON output: https://api-docs.deepseek.com/guides/json_mode
- Context caching: https://api-docs.deepseek.com/guides/kv_cache
- Token usage: https://api-docs.deepseek.com/quick_start/token_usage
- Rate limits and isolation: https://api-docs.deepseek.com/quick_start/rate_limit
- Error codes: https://api-docs.deepseek.com/quick_start/error_codes
- Balance endpoint: https://api-docs.deepseek.com/api/get-user-balance
- Changelog and lifecycle: https://api-docs.deepseek.com/updates
- Responses API reference, for the explicit deferred-capability decision: https://api-docs.deepseek.com/api/create-response
- Anthropic compatibility, for the explicit deferred-capability decision: https://api-docs.deepseek.com/guides/anthropic_api

Record `verified_at`, source URL, effective date if stated, API-format scope, and lifecycle status beside provider-derived catalog facts. If current official sources disagree with this prompt, the current official source wins. Document the conflict and update the implementation and readiness plan accordingly.

## Baseline facts observed on 2026-08-05 — reverify before use

- The direct API model list contains `deepseek-v4-flash` and `deepseek-v4-pro`.
- OpenAI-format API base: `https://api.deepseek.com`.
- Authentication: HTTP Bearer API key.
- Non-billable/read-only discovery: `GET /models`.
- Primary inference surface for this delivery: `POST /chat/completions`, sync and SSE.
- Both models advertise 1M context, maximum 384K output, text output, Thinking/Non-Thinking, JSON Output, and caller-defined function tools.
- Thinking is enabled by default. Current effort behavior is model-specific and must be represented and tested per the live official matrix; do not pretend compatibility-mapped values are distinct native effort levels.
- In thinking mode, sampling parameters documented as ignored must not be presented to GodarAI users as effective controls.
- In a thinking-mode tool loop, the provider requires `reasoning_content` to be passed back in subsequent tool requests. Preserve it losslessly in the API contract and tunnel, but never log it as ordinary trace content.
- Context caching is automatic. Usage exposes separate cache-hit and cache-miss input token counts.
- Current documented direct prices per 1M tokens are:
  - Flash: cache-hit input `$0.0028`, cache-miss input `$0.14`, output `$0.28`.
  - Pro: cache-hit input `$0.003625`, cache-miss input `$0.435`, output `$0.87`.
- The pricing page announces a future peak/off-peak policy with 2x peak pricing but gives no effective date. Do not activate that condition until an official effective announcement exists. Design catalog revisions so it can be added without rewriting historical billing.
- Current documented account concurrency is 2500 for Flash and 500 for Pro. Treat these as provider/account reference metadata, not a universal GodarAI tenant quota.
- The provider can emit empty keep-alive lines for non-streaming requests and `: keep-alive` SSE comments for streaming requests. Parsers must tolerate both.
- Current documented errors include 400, 401, 402, 422, 429, 500, and 503.
- The 2026-07-31 changelog calls the official V4 Flash API a public beta. The official V4 Pro release is still announced as forthcoming. Do not label either model GA without newer official evidence.
- Responses API is currently documented for Flash but not Pro. It is outside this two-model P0 because the supported surface is asymmetric.
- Legacy `deepseek-chat` and `deepseek-reasoner` were retired on 2026-07-24. Do not add them to the public catalog or silently present them as current models.

## Product scope

### Include now — P0

- Provider catalog entry for direct DeepSeek.
- Exactly the two canonical models above.
- Tenant BYOK and platform-managed API-key credentials.
- Safe official-host base URL handling.
- Read-only credential/model probe through `GET /models`.
- OpenAI-compatible Chat Completions, synchronous and SSE.
- Thinking enable/disable and documented effort behavior per model.
- Lossless `reasoning_content` handling required for multi-turn and tool loops.
- Caller-defined function tools and documented `tool_choice` values.
- JSON Object output through the documented Chat contract.
- Automatic cache usage accounting with separate hit/miss token dimensions.
- Fail-closed USD token billing and existing IRR wallet conversion.
- Provider error normalization, cancellation, timeout, keep-alive, and bounded safe-read retry behavior.
- Catalog/Playground/Logs discoverability through the existing provider-account and FQN model.
- Persian developer documentation, readiness runner, dashboards/alerts, source provenance, canary, and rollback controls.

### Defer behind independent capability gates

- Responses API, even for Flash, until a separate provider-aware Responses scope is approved and Pro reaches compatible support.
- Anthropic-format public proxying.
- FIM Completion and Chat Prefix Completion because they are Beta and use a different base URL/contract.
- Strict tool mode because it is Beta and has a restricted JSON Schema subset.
- Provider-hosted web search, code execution, Files, embeddings, images, audio, Batch, MCP, or any built-in tool not documented for this direct API surface.
- Account balance exposure to tenants. `GET /user/balance` may be used only for an operator-owned, redacted readiness check if product/security approve it.
- Public legacy aliases for `deepseek-chat`, `deepseek-reasoner`, or `deepseek-coder`.
- Automatic enablement of newly discovered models.

### Exclude

- Third-party-hosted DeepSeek weights under the direct `deepseek` provider. Those belong under their actual provider, such as OpenRouter or DeepInfra.
- Any zero-price or generic-provider pricing fallback.
- Arbitrary upstream paths, arbitrary custom headers, or an unrestricted custom host.
- Blind retries of billable POST requests.

## Required implementation work

### Phase 0 — evidence and scope lock

- Create `plans/deepseek-two-model-production-readiness.md` with status `not production-ready` or the repository's equivalent gated status.
- Include a dated official-source manifest, documented facts, GodarAI product decisions, assumptions, known gaps, release gates, and rollback controls.
- Confirm whether a `deepseek` Provider already exists in model-discovery data. Add an idempotent migration/upsert only if needed; do not create duplicates.
- Confirm the exact current model-runtime schema and pricing dimensions before writing YAML.
- Record lifecycle truth separately from runtime availability: an API-accessible preview/public-beta model is not automatically GA.

Exit evidence:

- every provider fact in the plan has an official URL and observation date;
- scope contains exactly two models;
- no stale LiteLLM alias is treated as current DeepSeek truth.

### Phase 1 — provider catalog and import pipeline

- Add:
  - `apps/model-discovery/providers/deepseek/default.yaml`;
  - `apps/model-discovery/providers/deepseek/deepseek-v4-flash.yaml`;
  - `apps/model-discovery/providers/deepseek/deepseek-v4-pro.yaml`.
- Store canonical IDs, display names, lifecycle, modalities, supported modes, context/output limits, capabilities, source URLs, and runtime profiles.
- Represent Thinking as runtime-profile behavior, not an alias or separate model.
- Include exactly three current billing dimensions per model:
  - uncached/cache-miss input;
  - cache-hit input;
  - output.
- Use per-token decimal rates derived from the current official per-1M-token prices. Preserve exact decimal behavior in Go billing; do not rely on binary floating-point for wallet settlement.
- Do not add the announced 2x peak-price condition until its official effective date is published. Make source-drift detection flag the announcement/effective-date change for review.
- Add a dedicated DeepSeek model importer and allowlist validator following the current Moonshot/xAI pattern, but enforce official DeepSeek source domains and exactly two files.
- Register the importer in the current YAML import pipeline and add importer/schema tests.
- Add or update the provider database migration with canonical slug `deepseek`, official docs URL, API-key auth schema, and safe defaults. Keep migrations idempotent and reversible under the repository's existing conventions.
- Ensure model catalog results expose the correct canonical model, lifecycle label, capabilities, prices, and FQN without creating a new catalog entity.

Exit evidence:

- exactly two DeepSeek model documents import successfully;
- missing, extra, duplicate, incorrectly named, unofficially sourced, unpriced, or invalid-profile models fail deterministically;
- no enabled DeepSeek model can load with missing/zero required rates;
- Flash and Pro lifecycle/capability differences remain visible.

### Phase 2 — credentials, host security, and tenant isolation

- Add direct DeepSeek to platform-managed credential contracts and tenant BYOK.
- Auth mode: `api_key`; inject `Authorization: Bearer <key>` only in Silk Road at dispatch time.
- Canonical base URL: `https://api.deepseek.com`.
- Allow only the official `api.deepseek.com` host for this provider unless a future officially documented region is explicitly added.
- Reject userinfo, fragments, unexpected schemes/ports, deceptive suffixes, and arbitrary paths according to existing SSRF/base-URL validation.
- Use `GET /models` as the read-only credential probe. Verify the response contains the configured canonical model before claiming readiness.
- Cover success, 401, 402, 429, 5xx, timeout, malformed JSON, wrong model list, wrong host, and secret redaction.
- Do not print, return, persist in plaintext, or place the provider key in the tunnel payload outside the existing encrypted credential mechanism.
- Because platform-managed credentials can be shared across GodarAI tenants, use DeepSeek's documented `user_id` isolation safely:
  - never forward raw email, phone, tenant name, or another PII value;
  - derive or reuse a stable opaque tenant/account-scoped identifier using an existing approved HMAC/pseudonymization facility;
  - validate the provider's documented character and length constraints;
  - test that two tenants sharing a provider credential receive different opaque IDs and that logs do not reveal the source identity.
  - if the current architecture cannot do this safely, keep platform-managed DeepSeek disabled, document the blocker, and permit only isolated BYOK accounts until fixed.

### Phase 3 — public request contract and runtime validation

- Inspect `packages/contracts/openai` before adding fields. Reuse current `thinking`, `reasoning_effort`, `reasoning_content`, tools, response-format, and cache-usage fields where present. Extend contracts only when required and preserve backward compatibility.
- Add a dedicated DeepSeek runtime-profile validator under the Control Plane proxy, with model-specific positive and negative tests.
- Validate before dispatch:
  - endpoint and transport eligibility;
  - canonical model identity;
  - text-only documented input/output modalities for this scope;
  - supported message roles;
  - context/output bounds;
  - `thinking.type` values and default behavior;
  - current per-model effort values and compatibility mappings;
  - sampling parameters that are ineffective in Thinking mode;
  - caller-defined function tools and `tool_choice`;
  - JSON Object response format;
  - `user_id` shape after pseudonymization;
  - unsupported/Beta parameters and additional fields.
- Reject invalid or misleading combinations before provider dispatch with the existing stable OpenAI-like error envelope.
- Do not silently drop `reasoning_content`, tool calls, tool results, finish reasons, served model, system fingerprint, or usage extensions.
- Preserve `reasoning_content` in thinking-mode tool loops exactly as the official DeepSeek contract requires. Add sync and SSE golden fixtures for this behavior.

### Phase 4 — Silk Road adapter, streaming, and reliability

- Add a dedicated DeepSeek adapter instead of registering only a generic OpenAI-compatible URL.
- Register only fixed operations for this P0:
  - `GET /models`;
  - `POST /chat/completions`.
- Use the canonical official base URL and Bearer auth.
- Prevent arbitrary upstream path/header forwarding and never forward the customer's GodarAI Authorization header.
- Preserve synchronous and SSE bodies without lossy translation.
- Correctly handle:
  - SSE `data:` chunks and `[DONE]`;
  - `: keep-alive` comments;
  - non-streaming empty keep-alive lines;
  - optional final usage chunks;
  - client cancellation;
  - provider disconnect before `[DONE]`;
  - mid-stream provider errors without reporting false success.
- Retry only safe read operations with a small bounded budget and jitter. Do not retry Chat creation after an unknown/billable outcome.
- Normalize or pass through documented 400/401/402/422/429/500/503 behavior according to GodarAI's existing error contract. Preserve useful request IDs and approved rate-limit headers when officially documented; never expose secrets.

Exit evidence:

- direct adapter tests prove exact method, path, body, headers, and content type;
- sync/SSE golden tests cover both models and both Thinking modes;
- keep-alive, cancellation, disconnect, and mid-stream fault tests pass;
- unsupported operations never reach DeepSeek.

### Phase 5 — billing, wallet, observability, and privacy

- Add a DeepSeek catalog-pricing snapshot manager and fail-closed completeness validation.
- Estimate/reserve/finalize using the same immutable catalog/profile revision for one request.
- Map official usage fields exactly:
  - `prompt_cache_hit_tokens` -> cache-hit input dimension;
  - `prompt_cache_miss_tokens` -> uncached input dimension;
  - completion/output tokens -> output dimension.
- Confirm from current official usage docs how reasoning tokens contribute to completion totals. Never double-charge `reasoning_content`.
- Reconcile exact provider cost, GodarAI sell price, IRR wallet debit, spend log, trace cost, and stream finalization.
- Fail closed before dispatch if pricing is missing, stale beyond policy, ambiguous, internally inconsistent, or lacks provenance. Never use zero or another provider's price.
- Emit low-cardinality operational fields for provider, account, requested/served model, runtime-profile revision, pricing revision, Thinking enabled, effective effort, cache-hit/miss tokens, retry count, status class, latency, and upstream request ID.
- Never log API keys, raw prompts/responses, `reasoning_content`, tool arguments/results, raw `user_id` source material, or full upstream error bodies that may contain sensitive content.
- Alert on unpriced usage, served-model mismatch, pricing/source drift, 401/402, 429 surge, 5xx, malformed usage, cache-accounting mismatch, mid-stream failure, and billing reconciliation variance.

Pricing tests must use exact decimal assertions for both models and all three dimensions. Include boundary tests for a future condition-aware peak/off-peak catalog revision without activating an unannounced rate.

### Phase 6 — docs, readiness, canary, and release

- Add `apps/docs/docs/providers/deepseek.mdx` and register it in the existing docs navigation.
- Write Persian documentation consistent with GodarAI terminology and include:
  - BYOK and platform-managed setup;
  - canonical FQNs;
  - synchronous and streaming examples against the GodarAI base URL;
  - Thinking enable/disable and effort examples;
  - the required `reasoning_content` tool-loop behavior;
  - function tools and JSON Object examples;
  - pricing dimensions and IRR display caveat;
  - context/output limits, lifecycle labels, errors, keep-alive behavior, privacy, and troubleshooting;
  - an explicit list of deferred/unsupported endpoints and Beta features;
  - no provider secret in examples.
- Add a bounded `services/control-plane/cmd/deepseek-readiness` command modeled on current provider readiness runners. It must:
  - call GodarAI by default, not bypass it;
  - accept a virtual key and provider-account identifier;
  - require an explicit approved spend ceiling before any billable test;
  - never accept, print, or store a provider key in ordinary mode;
  - optionally support a separately gated direct-provider comparison mode using `DEEPSEEK_API_KEY` without including it in output;
  - produce a sanitized JSON report;
  - always report `production_ready: false` until all human release gates are attached.
- Live readiness matrix for each model:
  - model discovery and served identity;
  - plain sync and SSE;
  - Thinking disabled and enabled;
  - every currently documented native/compatibility effort value and negative combinations;
  - multi-turn reasoning;
  - sequential function-tool loop with required reasoning preservation;
  - JSON Object output;
  - cache miss followed by an eligible cache hit;
  - usage and exact-cost reconciliation;
  - 401, 402, 422, 429, 5xx, timeout, cancellation, keep-alive, and mid-stream failure behavior.
- Canary Flash first, then Pro. Keep their gates independent because lifecycle and capability maturity differ.
- Add provider/model kill switches and a rollback drill.
- Add a first-party source-drift check for sitemap, model IDs, lifecycle, context/output limits, effort mapping, Responses support, prices, future peak/off-peak effective date, and concurrency metadata. Drift must open a reviewed catalog/profile revision; never rewrite historical price rows.

## Required automated tests

At minimum add or extend tests for:

1. DeepSeek provider YAML/import allowlist, official-source validation, runtime-profile schema, canonical IDs, lifecycle, and pricing completeness.
2. Idempotent provider migration/upsert and no duplicate catalog rows.
3. Platform-managed and BYOK credentials, host allowlist, probe behavior, redaction, and opaque tenant isolation.
4. Control Plane model-specific runtime validation for modes, efforts, sampling, tools, JSON, roles, modalities, limits, and unsupported fields.
5. Silk Road exact dispatch contract, auth injection, safe operations, retries, keep-alive, sync/SSE, cancellation, and error propagation.
6. Lossless `reasoning_content` and tool-loop round trip for both sync and streaming.
7. Cache-hit/cache-miss usage parsing and exact decimal billing for both models.
8. Fail-closed behavior for missing/stale/ambiguous prices and served-model mismatch.
9. Trace/log/readiness redaction of secrets, prompts, reasoning, tools, and identity source material.
10. Documentation links/build and sanitized readiness output.

Do not make live billable calls in CI.

## Validation commands

Use `pnpm`, never `npm`.

Run at least:

- `pnpm --filter model-discovery lint`
- `pnpm --filter model-discovery typecheck`
- `pnpm --filter model-discovery test`
- `pnpm --filter model-discovery build`
- `pnpm --filter @shahrah/docs lint`
- `pnpm --filter @shahrah/docs build`

For every changed Go module, run formatting first on every changed Go file:

- `gofmt`
- `goimports`

Then, from each changed Go module such as `packages/contracts`, `services/control-plane`, and `services/silk-road`, run:

- `go vet ./...`
- `go test ./...`
- `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`

Do not weaken `.golangci.yml`, skip failing tests, add broad `nolint`, or change unrelated code to make checks pass. If a repository-wide pre-existing failure remains, prove it is pre-existing and report it precisely while still running focused tests for all touched code.

## Release gates

Do not label the provider or either model production-ready until all are complete:

1. Official facts reverified and catalog provenance approved.
2. Dedicated provider account and approved live-test spend ceiling.
3. Full live readiness matrix green for both platform-managed and isolated BYOK paths.
4. Cache miss/hit/output charges reconciled against the DeepSeek account/dashboard for each model.
5. IRR wallet settlement reconciled with exact USD provider cost and current FX policy.
6. Tenant isolation and pseudonymous `user_id` privacy review approved.
7. Error, timeout, keep-alive, cancellation, SSE, and 429 behavior verified through the Iran-Europe tunnel.
8. Logs/traces/readiness reports reviewed for secret, reasoning, tool, and prompt leakage.
9. Dashboards, alerts, on-call ownership, kill switches, and rollback drill complete.
10. Product, engineering, security/privacy, finance, and operations approvals attached.

Flash and Pro require separate sign-off. API availability is not GA evidence.

## Completion criteria

The task is complete only when:

- direct DeepSeek exists as a real provider in the catalog and provider-account flows;
- exactly the two target models are importable, discoverable, selectable, routable, billable, observable, and documented;
- Chat sync/SSE, Thinking, tools, JSON, cache accounting, and errors work end to end through Control Plane -> tunnel -> Silk Road -> provider mock/live readiness path;
- required prices are exact, sourced, versioned, and fail closed;
- platform-managed credentials cannot cause cross-tenant cache/identity leakage;
- no legacy/third-party model is mislabeled as direct DeepSeek;
- all automated validation passes;
- readiness still reports not production-ready until external release evidence is attached.

## Final response format

Return:

1. outcome summary;
2. documented provider facts versus GodarAI product decisions versus assumptions;
3. files changed by subsystem;
4. migrations and rollout/rollback behavior;
5. tests and exact command results;
6. official DeepSeek sources used, with `verified_at`;
7. live tests not run and why;
8. remaining release gates and current production-readiness status.
```

## Scope rationale

- `deepseek-v4-flash` is the default high-throughput/value model.
- `deepseek-v4-pro` is the premium complex-reasoning model, but its lifecycle must remain explicit until DeepSeek publishes newer GA evidence.
- Chat Completions is the common production surface across both models.
- Responses, Anthropic format, FIM, prefix completion, and strict tools remain separate capability gates to avoid claiming asymmetric or Beta support as full two-model parity.

## Evidence used to create this prompt

Verified on 2026-08-05 from official DeepSeek sources:

- [Models and pricing](https://api-docs.deepseek.com/quick_start/pricing)
- [Model discovery](https://api-docs.deepseek.com/api/list-models)
- [Authentication](https://api-docs.deepseek.com/api/deepseek-api)
- [Chat Completions](https://api-docs.deepseek.com/api/create-chat-completion)
- [Thinking mode](https://api-docs.deepseek.com/guides/thinking_mode)
- [Tool calls](https://api-docs.deepseek.com/guides/tool_calls)
- [JSON output](https://api-docs.deepseek.com/guides/json_mode)
- [Context caching](https://api-docs.deepseek.com/guides/kv_cache)
- [Rate limits and keep-alive](https://api-docs.deepseek.com/quick_start/rate_limit)
- [Error codes](https://api-docs.deepseek.com/quick_start/error_codes)
- [Changelog and lifecycle](https://api-docs.deepseek.com/updates)

Shahrah-specific constraints came from:

- `plans/README.md`;
- `plans/model-catalog-experience-prd.md`;
- `plans/platform-managed-provider-credentials-provider-matrix.md`;
- `plans/moonshot-four-model-production-readiness.md`;
- current model-discovery, Control Plane, Silk Road, billing, and readiness implementations.

LiteLLM documentation inspected:

- `.llmLite_docs/providers/deepseek.md`.

No LiteLLM source code was inspected because the current official DeepSeek contract and Shahrah's existing provider patterns were sufficient.

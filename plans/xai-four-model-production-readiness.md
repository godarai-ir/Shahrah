# xAI four-model full-support execution plan

**Status:** Local implementation in progress; external release gates pending  
**Baseline date:** 2026-07-31  
**Provider slug:** `xai`  
**Models:** `grok-4.3`, `grok-4.5`, `grok-imagine-image`, `grok-imagine-image-quality`  
**Current release gate:** Not production-ready

This plan converts the xAI/Grok support audit into an executable delivery
program. It covers provider discovery, credentials, model-specific request and
response contracts, public endpoints, upstream dispatch, pricing, usage
accounting, observability, automated validation, documentation, and rollout.

No model in this scope may be advertised as fully supported merely because its
payload can pass through the generic OpenAI-compatible adapter. A model becomes
production-ready only after all applicable gates in this document have direct
evidence.

## 1. Delivery outcome

GodarAI must provide complete, model-specific support for the four named xAI
models through its existing OpenAI-compatible public API, plus the xAI resource
operations needed by those models. The implementation must:

1. discover the correct model IDs and aliases under canonical provider `xai`;
2. expose only endpoints and parameters that the selected model supports;
3. preserve xAI request and response extensions without losing billing data;
4. validate incompatible options before reserving wallet funds or dispatching;
5. support platform-managed and tenant BYOK xAI credentials;
6. estimate costs from a pinned, effective-dated catalog snapshot;
7. finalize and reconcile provider cost using xAI's exact integer cost ticks;
8. account for caching, long context, priority, Batch, image input/output, and
   server-side tools;
9. preserve sync, SSE, deferred, stored-response, and batch resource semantics;
10. provide traces, metrics, errors, rate-limit headers, and safe retry behavior;
11. pass unit, contract, integration, live conformance, billing reconciliation,
    resilience, and canary gates.

Feature parity is required over GodarAI's downstream REST API. Exposing xAI's
native gRPC protocol to GodarAI clients is not required because the same model
features are available through xAI REST. Silk Road may use REST upstream. If a
future requirement demands downstream gRPC protocol parity, it should be a
separate transport project and must not delay correct REST support.

## 2. Source-of-truth policy

The supplied [Apidog `llms.txt`](https://grok-api.apidog.io/llms.txt) remains a
required historical reference, especially for deferred chat, streaming, and
the earlier OpenAI-compatible surface. It is not sufficiently current for the
four models in this plan: its model/pricing material predates current xAI model
cards and some old image statements conflict with the current API.

Use this precedence order when implementing or reconciling a conflict:

1. current official xAI REST reference and model-specific xAI model page;
2. current official xAI pricing and capability guides;
3. a sanitized, credentialed response from xAI's model endpoints for the
   deployment region and team;
4. the supplied Apidog reference for still-current endpoints not contradicted
   by current xAI documentation;
5. GodarAI's catalog and documentation, which describe implementation state but
   are never provider truth.

Required official sources:

- [Grok 4.3 model card](https://docs.x.ai/developers/models/grok-4.3)
- [Grok 4.5 model card](https://docs.x.ai/developers/models/grok-4.5)
- [Grok 4.5 model card](https://docs.x.ai/developers/models/grok-4.5)
- [Grok Imagine Image model card](https://docs.x.ai/developers/models/grok-imagine-image)
- [Grok Imagine Image Quality model card](https://docs.x.ai/developers/models/grok-imagine-image-quality)
- [Pricing](https://docs.x.ai/developers/pricing)
- [Inference REST API](https://docs.x.ai/developers/rest-api-reference/inference)
- [Chat and Responses REST reference](https://docs.x.ai/developers/rest-api-reference/inference/chat)
- [Image generation](https://docs.x.ai/developers/model-capabilities/images/generation)
- [Image editing](https://docs.x.ai/developers/model-capabilities/images/editing)
- [Multi-image editing](https://docs.x.ai/developers/model-capabilities/images/multi-image-editing)
- [Imagine Files integration](https://docs.x.ai/developers/model-capabilities/imagine/files)
- [Files management](https://docs.x.ai/developers/rest-api-reference/files/manage)
- [Public URLs](https://docs.x.ai/developers/files/public-urls)
- [Batch API](https://docs.x.ai/developers/advanced-api-usage/batch-api)
- [Deferred Chat Completions](https://docs.x.ai/developers/advanced-api-usage/deferred-chat-completions)
- [Context compaction](https://docs.x.ai/developers/advanced-api-usage/context-compaction)
- [Reasoning](https://docs.x.ai/developers/model-capabilities/text/reasoning)
- [Function calling](https://docs.x.ai/developers/tools/function-calling)
- [Structured outputs](https://docs.x.ai/developers/model-capabilities/text/structured-outputs)
- [Prompt caching](https://docs.x.ai/developers/advanced-api-usage/prompt-caching)
- [Cost tracking](https://docs.x.ai/developers/cost-tracking)
- [Rate limits](https://docs.x.ai/developers/rate-limits)
- [Error behavior](https://docs.x.ai/developers/debugging)

Every imported model, pricing rate, and runtime profile must retain source URL,
observation time, effective time, and catalog revision. When two current
official sources disagree, the relevant capability stays disabled until the
conflict is recorded and resolved. No billing value may be guessed.

### 2.1 Current repository baseline

This plan starts from the following observed state, not from a greenfield
assumption:

| Area | Current implementation | Planning consequence |
|---|---|---|
| Catalog | The worktree has moved the four files from provider `groq` to `xai`, but all four model files are empty and `xai/default.yaml` still contains Groq sources/defaults. | Phase 1 must preserve the canonical move, replace inherited metadata, and prove an exact four-model import. |
| Credentials | Platform-managed `xai` probing, Bearer auth schema, default base URL, and host allowlist exist. Tenant BYOK validation does not include `xai`. | Reuse the secure schema and add tenant flow plus model-specific tests. |
| Silk Road | `xai` is registered directly to the generic OpenAI-compatible adapter. | Introduce an xAI wrapper for headers, operations, retries, and runtime-profile enforcement. |
| Chat | Generic Chat can dispatch sync/SSE traffic, preserves several OpenAI-compatible fields, and validates only a generic reasoning set. Non-stream responses are decoded and reserialized through the common contract. | Add model-specific validation and typed/lossless xAI usage so cost/tool fields survive. |
| Responses | Public routes exist, but provider enablement allows only `openai`/`mock`; resource resolution and mappings hard-code `openai`. | Generalize provider-aware ownership before enabling xAI Responses. |
| Images | Generation and JSON edits can reach xAI and raw extra fields can pass through. Metadata/billing validation does not understand the full xAI image contract. | Add explicit aspect ratio, resolution, reference, storage, and JSON-only edit rules. |
| Batch and Files | OpenAI-style resource routes and fixed operations exist, but mappings/schema assume OpenAI and xAI Batch semantics have additional subresources. | Generalize Files ownership and implement xAI-native Batch operations rather than relabeling OpenAI Batch. |
| Pricing | Catalog snapshot infrastructure exists, but runtime provider selection does not include `xai`; unknown models can fall back to static/zero pricing. Image handlers already capture xAI ticks in a limited path. | Extend fail-closed catalog pricing to every target and make exact ticks cross-endpoint. |
| Headers/retries | Generic rate-limit headers and `Retry-After` are forwarded; arbitrary customer headers are intentionally not. The generic adapter performs a single request. | Add a narrow xAI header policy and operation-aware retries. |
| Tests/docs | A few xAI image tests exist. The xAI Postman collection is disabled and explicitly marked as a placeholder. | Replace placeholders with model-specific contract and live readiness coverage. |

No implementation phase may overwrite or discard unrelated user changes in the
current dirty worktree.

## 3. Verified target model contracts

### 3.1 Identity, modalities, limits, and aliases

| Canonical model | Aliases that must resolve | Input -> output | Context | Reasoning | Lifecycle |
|---|---|---|---:|---|---|
| `grok-4.3` | `grok-4.3-latest`, `grok-latest` | text, image -> text | 1,000,000 | `none`, `low`, `medium`, `high` | active |
| `grok-4.5` | `grok-4.5-latest`, `grok-build-latest` | text, image -> text | 500,000 | `low`, `medium`, `high`; default `high`; cannot disable | active |
| `grok-imagine-image` | `grok-imagine-image-2026-03-02` | text, image -> image | not token-context based | none | active |
| `grok-imagine-image-quality` | `grok-imagine-image-quality-20260403`, `grok-imagine-image-quality-latest`, legacy redirect `grok-imagine-image-pro` | text, image -> image | not token-context based | none | active |

Maximum output-token limits for `grok-4.3` and `grok-4.5` are not stated on the
current public model cards. Phase 0 must capture them from a current official
schema or credentialed model response. Until then the runtime profile must not
invent a limit, and local validation may enforce only a provider-confirmed
value. The legacy `grok-imagine-image-pro` identity must be represented as an
alias/redirect to the served quality model, never as a separately priced model.

### 3.2 Endpoint and feature target

Legend: **required** means current provider evidence confirms the capability;
**probe-gated** means the code and tests must be ready but catalog enablement
requires a successful model-specific live probe; **reject** means GodarAI must
return a deterministic 4xx without calling xAI.

| Capability | `grok-4.3` | `grok-4.5` | Imagine | Imagine Quality |
|---|---|---|---|---|
| `POST /v1/chat/completions` sync | required | required | reject | reject |
| Chat SSE streaming | required | required | reject | reject |
| Deferred chat create/retrieve | probe-gated | required | reject | reject |
| `POST /v1/responses` sync/SSE | required | required | reject | reject |
| Responses retrieve/delete/cancel/input items | probe-gated | probe-gated | reject | reject |
| `POST /v1/responses/input_tokens` | probe-gated | probe-gated | reject | reject |
| `POST /v1/responses/compact` | probe-gated | required | reject | reject |
| Function calling | required | required | reject | reject |
| Structured output / JSON schema | required | required | reject | reject |
| Web search, X search, code execution | live model matrix required | required | reject | reject |
| MCP and other server tools | live model matrix required | live model matrix required | reject | reject |
| Image understanding in text requests | required | required | not applicable | not applicable |
| `POST /v1/images/generations` | reject | reject | required | required |
| JSON `POST /v1/images/edits` | reject | reject | required | required |
| Multi-image editing, up to provider limit | reject | reject | required | required |
| Multipart image editing | reject for xAI | reject for xAI | reject | reject |
| `POST /v1/images/variations` | reject | reject | reject unless xAI documents it | reject unless xAI documents it |
| xAI Batch requests | required | reject; xAI explicitly disallows 4.5 | required at standard price | required at standard price |
| Files as image input / persisted output | not required for text-only baseline | not required for text-only baseline | required | required |
| URL and base64 image output | not applicable | not applicable | required | required |
| Image output streaming | reject | reject | reject | reject |

The current xAI Batch guide includes `grok-4.3` Responses requests with
server-side tools, image generation, image editing, and MCP examples. The live
conformance matrix, not the generic adapter, decides which tools are published
for each model and region.

### 3.3 Request contract

Text request support must include, when allowed by the selected model and
endpoint:

- `system`, `user`, `assistant`, and `tool` messages;
- `developer` messages only if a current xAI schema and live probe accept them;
  otherwise return an explicit model-capability error;
- string and typed multimodal content, including image URL and base64 input;
- `max_tokens` / `max_output_tokens`, `temperature`, `top_p`, `stream`,
  `stream_options`, `n`, `seed`, `logprobs`, and `top_logprobs` where documented;
- `reasoning_effort` with distinct 4.3 and 4.5 validation;
- no `presence_penalty`, `frequency_penalty`, or `stop` for 4.5 reasoning;
- function tools, `tool_choice`, parallel tool calls, and strict JSON schemas;
- Responses server tools and their provider-native configuration objects;
- `response_format` values `text`, `json_object`, and `json_schema` where
  supported;
- `previous_response_id`, `store`, `include`, truncation, metadata, and
  provider-documented Responses controls;
- `prompt_cache_key` on Responses;
- `x-grok-conv-id` on Chat, passed only from an explicit safe request header;
- `service_tier=priority` where xAI accepts it and effective tier reporting.

Sampling parameters must be governed by the selected model profile, not one
gateway-wide range. Tests must prove each accepted parameter and each
reasoning-related exclusion for both text models.

Image request support must include:

- prompt and model;
- `n` within the current provider limit;
- `aspect_ratio` with the current documented enum, including `auto`;
- `resolution` values `1k` and `2k`;
- `response_format` values `url` and `b64_json`;
- one or more image inputs as URL, data URL/base64, or xAI `file_id`;
- the current multi-image input limit;
- `storage_options.filename`, file expiry, public URL creation, and public URL
  expiry with the provider's 1-hour to 30-day constraints;
- output `file_output`, public URL metadata, expiry metadata, and partial public
  URL failure details;
- JSON-only editing for xAI; never translate it into multipart upstream.

Unknown fields are not automatically supported. Lossless forwarding is useful,
but model-specific allowlists and tests are still required before GodarAI claims
support.

### 3.4 Response contract

GodarAI must preserve the following without changing their meaning:

- actual served `model`, response IDs, timestamps, status, and
  `system_fingerprint`;
- final text, refusals, citations, function calls, server-tool items, and tool
  results;
- `reasoning_content` where Chat exposes it;
- encrypted reasoning content requested through Responses `include`;
- prompt/input, completion/output, cached, image, and reasoning token details;
- `num_server_side_tools_used` and any available per-tool usage;
- `cost_in_usd_ticks` as a signed-range-checked 64-bit integer;
- effective `service_tier`;
- streaming usage and running cost, with the final chunk treated as final cost;
- image URL or base64 data, `file_output`, public URL fields, and partial storage
  errors;
- upstream request IDs, rate-limit headers, `Retry-After`, and provider error
  bodies after secret-safe normalization.

## 4. Pricing and accounting target

All values below are effective-dated catalog inputs, not hard-coded runtime
constants.

### 4.1 Text rates

Prices are USD per 1M tokens. Long-context pricing applies to all tokens in the
request once prompt tokens reach 200,000.

| Model | Short input | Short cached | Short output | Long input | Long cached | Long output |
|---|---:|---:|---:|---:|---:|---:|
| `grok-4.3` | 1.25 | 0.20 | 2.50 | 2.50 | 0.40 | 5.00 |
| `grok-4.5` | 2.00 | 0.30 | 6.00 | 4.00 | 0.60 | 12.00 |

Additional conditions:

- `grok-4.3` Batch receives a 20% discount across input, output, cached, and
  reasoning token types;
- `grok-4.5` Batch is unsupported and must be rejected;
- priority processing is 2x standard text rates across all token types and is
  charged only when the response confirms `service_tier=priority`;
- reasoning tokens are reported separately but billed at the applicable output
  rate unless xAI publishes a distinct rate;
- cached tokens must be removed from ordinary input quantity and priced exactly
  once;
- long-context, Batch, and priority conditions must compose only in combinations
  the provider permits.

### 4.2 Image rates

| Model | Image input | 1K output | 2K output | Batch |
|---|---:|---:|---:|---|
| `grok-imagine-image` | $0.002/image | $0.02/image | $0.02/image | standard rates |
| `grok-imagine-image-quality` | $0.01/image | $0.05/image | $0.07/image | standard rates |

Text prompt input currently has no separately published image-model charge and
must not be assigned one. Each input image and each generated output must be
counted. The quality model's resolution-dependent output price makes
`resolution` a required billing condition.

### 4.3 Server tools and Files

At the baseline date, the xAI pricing page publishes:

| Dimension | Rate |
|---|---:|
| Web Search | $5 / 1,000 successful calls |
| X Search | $5 / 1,000 successful calls |
| Code Execution / Code Interpreter | $5 / 1,000 successful calls |
| File Attachments | $10 / 1,000 successful calls |
| Collections / File Search | $2.50 / 1,000 successful calls |
| File storage | $0.025 / GiB / day |
| Collection storage | $0.10 / GiB / day |
| File or collection downloads | $0.20 / GiB |

Only tool families enabled by a model's runtime profile may be estimated.
Successful billable calls must come from provider usage, not attempted tool
calls. Storage and download charges are asynchronous/provider-account costs and
need a reconciliation job before GodarAI resells them; they must not be silently
ignored if Files persistence is exposed.

### 4.4 Exact-cost invariant

xAI reports `cost_in_usd_ticks` on Chat, Responses, image, and streaming usage,
where 1 USD equals 10,000,000,000 ticks. The implementation must maintain two
separate values:

1. **GodarAI estimate and component ledger:** derived before dispatch from the
   pinned catalog revision for wallet reservation and explainable billing;
2. **xAI exact provider cost:** derived after completion from integer ticks and
   stored for reconciliation and provider-margin accounting.

Final tenant sell-price policy must be explicit. It may use exact provider cost
plus the configured commercial policy, but it must never mix raw ticks with a
catalog total twice. A reconciliation delta between the component ledger and
exact xAI cost must be recorded and alerted. Missing or malformed ticks on a
supposedly successful production response are a billing anomaly, not zero cost.

## 5. Architecture changes

### 5.1 Canonical provider identity

Use `xai` everywhere: provider catalog, connection schema, deployment,
resource mappings, billing snapshot, metrics, and model FQNs. Do not introduce
`grok` as a provider alias; Grok is the model family. Preserve the current
worktree move from `providers/groq` to `providers/xai`, then replace the copied
Groq metadata.

Add a central normalization test covering:

- `xai:<account>:<model>` FQNs;
- `xai/<model>` catalog IDs where accepted;
- canonical IDs and aliases;
- no collision between `xai`, `groq`, and the Grok model-family prefix.

### 5.2 Runtime profile extensions

The existing versioned runtime profile is the dispatch contract. Extend it
instead of placing xAI rules in unrelated handler switches. Required additions:

- request-header allowlist and per-endpoint header controls;
- long-context threshold and condition metadata;
- model-specific Batch eligibility;
- deferred-completion eligibility and retention semantics;
- response storage/retrieval/compaction controls;
- image controls: allowed aspect ratios, resolutions, response formats, maximum
  outputs, maximum reference images, JSON-only edits, Files input, and storage
  options;
- pricing-required response fields, including exact cost ticks;
- explicit unsupported parameter sets for reasoning models;
- tool families and provider names `x_search`, `attachment_search`,
  `collections_search`, and any confirmed MCP support.

Change both schema implementations together:

- [`packages/contracts/platform/model_runtime_profile.go`](../packages/contracts/platform/model_runtime_profile.go)
- [`apps/model-discovery/src/catalog/runtime-profile.schema.ts`](../apps/model-discovery/src/catalog/runtime-profile.schema.ts)

Runtime profiles remain immutable after import. Any correction creates a higher
profile version and a new catalog revision.

### 5.3 Dedicated xAI policy and adapter

Keep the finite operation compiler in Silk Road, but register a dedicated xAI
adapter rather than the generic adapter directly. The xAI wrapper owns:

- fixed endpoint/method allowlisting;
- safe forwarding of `x-grok-conv-id`;
- JSON-only edit enforcement;
- xAI Batch and Files/public-URL operations;
- provider-specific retry eligibility;
- response-header preservation;
- defense-in-depth checks against the pinned runtime profile.

Primary files:

- [`services/silk-road/pkg/server/adapters.go`](../services/silk-road/pkg/server/adapters.go)
- [`services/silk-road/pkg/server/openai_compatible.go`](../services/silk-road/pkg/server/openai_compatible.go)
- [`services/silk-road/pkg/server/openai_operations.go`](../services/silk-road/pkg/server/openai_operations.go)
- new `services/silk-road/pkg/server/xai.go`
- new `services/silk-road/pkg/server/xai_runtime_profile.go`
- new `services/silk-road/pkg/server/xai_retry.go`

The Control Plane performs customer-facing validation and billing. Silk Road is
the final SSRF/path/header boundary and must reject arbitrary paths even when a
Control Plane bug sends one.

### 5.4 Provider-aware resource ownership

Responses, Files, and Batch resource mappings currently hard-code provider
`openai`. Generalize resource ownership so each stored public ID pins:

- tenant;
- provider `xai`;
- provider-account ID;
- deployment ID;
- upstream resource ID;
- model and catalog/profile revision where relevant;
- status and expiry.

Continuation, retrieval, cancellation, deletion, file download, and public-URL
operations must use the pinned account and never reroute across xAI accounts.
The generalized code must keep OpenAI behavior unchanged.

Primary files:

- [`services/control-plane/internal/proxy/responses_handler.go`](../services/control-plane/internal/proxy/responses_handler.go)
- [`services/control-plane/internal/proxy/responses_operations_handler.go`](../services/control-plane/internal/proxy/responses_operations_handler.go)
- [`services/control-plane/internal/proxy/openai_batch_files_handler.go`](../services/control-plane/internal/proxy/openai_batch_files_handler.go)
- [`services/control-plane/internal/server/routes_proxy.go`](../services/control-plane/internal/server/routes_proxy.go)
- provider-native resource storage and migration code under
  `services/control-plane/internal/storage` and `services/control-plane/db/main/migrations`

### 5.5 Typed usage with lossless extension preservation

Extend the common contracts to retain xAI fields in both non-streaming and
streaming paths:

- [`packages/contracts/openai/types.go`](../packages/contracts/openai/types.go)
- Chat `Usage`: `cost_in_usd_ticks`, `num_server_side_tools_used`, and safe extra
  usage fields;
- Responses `ResponseUsage`: the same xAI fields and provider token details;
- image response data: typed `file_output` and an extra-field round trip;
- SSE aggregation: final ticks, model, tier, fingerprint, and tool usage.

The public response must preserve xAI extensions where OpenAI-compatible clients
can tolerate them. Internal billing must never depend on reparsing a redacted
trace body.

## 6. Phased implementation

No phase may be marked complete without its named evidence. A capability must
remain disabled if its billing and test work in the same phase is incomplete.

### Phase 0 — freeze the provider contract and red tests

Owner: provider integration + model catalog

Tasks:

- [ ] Save dated, sanitized fixtures for model metadata, Chat, Chat SSE,
  Responses, Responses SSE, image generation, image editing, deferred, Batch,
  Files, and provider errors.
- [ ] Resolve maximum output-token limits and request-size limits from an
  official schema or live account response.
- [ ] Record region/account-specific rate limits separately from model
  capabilities; do not convert them into tenant quotas automatically.
- [ ] Produce a model-by-endpoint-by-parameter conformance manifest.
- [ ] Add failing catalog schema tests for all four non-empty YAML files.
- [ ] Add failing provider-policy tests for incorrect reasoning, image
  multipart, image streaming, variations, and 4.5 Batch.
- [ ] Add failing billing tests for long context, cache, priority, Batch,
  image resolution, image count, tools, and exact ticks.
- [ ] Add failing resource-ownership tests proving that IDs cannot cross tenant
  or xAI provider-account boundaries.
- [ ] Record unresolved source conflicts in this plan's ambiguity register.

Exit evidence:

- a reviewed fixture/source manifest;
- red-test inventory mapped to every required row in section 3.2;
- no invented limits or prices;
- security review of new header and resource-ID handling.

### Phase 1 — catalog, profiles, aliases, and credentials

Owner: model discovery + provider accounts

Tasks:

- [ ] Replace Groq URLs and defaults in
  `apps/model-discovery/providers/xai/default.yaml` with xAI sources and only
  truly shared xAI parameters.
- [ ] Populate the four currently empty model YAML files with identity,
  lifecycle, modalities, capabilities, endpoints, transports, limits, sources,
  costs, and versioned runtime profiles.
- [ ] Add an xAI provider importer for condition-aware rates: normal token,
  cached, long context, Batch, priority, image input, 1K output, 2K output, and
  tool calls.
- [ ] Import aliases without creating duplicate billable models.
- [ ] Ensure `grok-imagine-image-pro` resolves to the quality model while the
  served model and price are recorded as quality.
- [ ] Extend catalog reconciliation to reject empty model documents, wrong
  provider documentation, missing required price dimensions, and duplicate
  aliases.
- [ ] Add `xai` to tenant BYOK credential probing in
  [`provider_credential_validation.go`](../services/control-plane/internal/storage/provider_credential_validation.go),
  using Bearer auth and the validated xAI model endpoint.
- [ ] Retain the existing platform-managed probe, host allowlist, encrypted
  secret handling, and write-only redaction; add xAI-specific tests.
- [ ] Document the authentication and environment contract. Production keys
  live in platform-managed or tenant provider accounts; Silk Road must not need
  a global `XAI_API_KEY`. Existing credential-store environment controls
  (`PLATFORM_PROVIDER_CREDENTIAL_MODE`,
  `PLATFORM_PROVIDER_CREDENTIAL_ENVIRONMENT`, and credential KEK variables)
  remain the deployment controls. Any environment break-glass credential must
  use the existing deployment path, be auditable, and never override BYOK
  account selection silently.
- [ ] Verify tenant-admin provider account setup reads the catalog connection
  schema correctly. UI changes are required only if the existing generic form
  cannot represent xAI.
- [ ] Ensure support-admin/model-discovery views show the xAI source revision,
  runtime-profile version, alias target, lifecycle, pricing readiness, and
  deployment enablement without implying empty/configured-only support.

Exit evidence:

- `pnpm exec tsx scripts/import-models-from-yaml.ts --providers xai --dry-run`
  reports exactly four target canonical models and zero schema errors;
- catalog API returns canonical identities, aliases, pricing rates, and profiles
  under provider `xai`;
- model discovery schema/import tests pass;
- platform-managed and BYOK probe tests cover success, 401/403, 429, timeout,
  malformed base URL, and secret redaction;
- no enabled xAI model has a zero or missing required rate.

Rollback:

- disable the four deployments; retain immutable catalog revisions.

### Phase 2 — xAI adapter, headers, Chat, and streaming

Owner: Silk Road + Control Plane proxy

Tasks:

- [ ] Register the dedicated xAI adapter with base URL
  `https://api.x.ai/v1` and Bearer auth.
- [ ] Add fixed operations for Chat and deferred retrieval; reject arbitrary
  upstream paths.
- [ ] Forward only documented safe request headers, including an explicitly
  validated `x-grok-conv-id`; never forward customer Authorization or arbitrary
  `x-*` headers.
- [ ] Apply per-model runtime validation before dispatch.
- [ ] Support text and image content input for 4.3 and 4.5.
- [ ] Validate 4.3 reasoning values independently from 4.5.
- [ ] Reject 4.5 reasoning disablement and incompatible penalty/stop fields.
- [ ] Round-trip function calls, tool results, strict structured outputs,
  refusals, citations, reasoning content, served model, and fingerprint.
- [ ] Preserve Chat SSE chunks, running usage/cost, final usage, cancellation,
  and mid-stream provider errors.
- [ ] Extend the provider response-header allowlist with current xAI request and
  rate-limit headers based on live evidence.
- [ ] Add bounded retries only for documented idempotent/read operations.
  Ordinary Chat create is not retried without an idempotency guarantee.

Exit evidence:

- unit and adapter tests for both text models and all reasoning modes;
- sync/SSE golden fixtures round-trip with no field loss;
- header tests prove cache affinity works and secrets/arbitrary headers do not
  pass;
- cancellation and mid-stream fault tests show no false `[DONE]` success;
- direct upstream mock tests prove exact method/path selection.

Rollback:

- per-model Chat feature flag and deployment disablement.

### Phase 3 — full Responses API and model tools

Owner: Control Plane proxy + provider integration

Tasks:

- [ ] Replace `supportsResponsesProvider` hard-coding with runtime-profile
  capability checks that allow `xai` only for the selected model.
- [ ] Generalize response resource resolution/mapping from OpenAI-only to the
  owning provider account.
- [ ] Preserve the raw Responses request contract rather than translating xAI
  Responses into a lossy Chat request.
- [ ] Support create sync/SSE, retrieve, delete, cancel, input items,
  input-token count, and compact where the live model matrix confirms them.
- [ ] Preserve `previous_response_id`, storage semantics, metadata, include,
  encrypted reasoning, truncation, typed output items, and server-tool results.
- [ ] Implement and validate function, web search, X search, code execution,
  and every additional model-confirmed tool family.
- [ ] Count successful server tool use without reducing it to only web/file
  search.
- [ ] Support `prompt_cache_key`, cached-token reporting, and effective service
  tier.
- [ ] Prevent stored response IDs from crossing tenant or xAI account
  boundaries.

Exit evidence:

- request/response golden tests for each operation and both text models;
- stateful continuation test pinned to the original account;
- tool matrix with positive and negative tests per model;
- encrypted reasoning and compacted output are byte-for-byte preserved where
  opaque;
- SSE and non-SSE usage/cost totals match.

Rollback:

- Responses capability flag independently from Chat.

### Phase 4 — image generation, editing, Files, and public URLs

Owner: image proxy + resource operations + billing

Tasks:

- [ ] Add typed xAI image validation for `aspect_ratio`, `resolution`, `n`,
  `response_format`, reference-image count, and `storage_options`.
- [ ] Support URL, data URL/base64, and `file_id` image inputs.
- [ ] Support JSON image generation and JSON-only image edits, including
  multiple images.
- [ ] Reject xAI multipart edits, variations, and streaming before dispatch.
- [ ] Round-trip URL/base64 outputs, `file_output`, public URL metadata,
  expiration, and partial public-URL errors.
- [ ] Generalize Files create/list/get/update/delete/content operations to xAI,
  including streaming uploads/downloads and resource-account pinning.
- [ ] Add `POST /v1/files/:file_id/public-url` and
  `POST /v1/files/:file_id/public-url/revoke` with provider capability checks.
- [ ] Enforce current xAI Files limits, public-URL content-type/size rules,
  expiry rules, and one-active-URL semantics from the profile/contract.
- [ ] Ensure image URLs and base64 payloads follow trace privacy policy and are
  not copied into ordinary logs.
- [ ] Count each input and output image and select the correct resolution rate.
- [ ] Capture exact cost ticks and reconcile them against image components.

Exit evidence:

- a complete matrix for two models x generation/edit x URL/base64/file input x
  URL/base64 output x 1K/2K x `n=1`/multiple;
- invalid ratio, resolution, count, expiry, multipart, variation, and stream
  tests fail before provider dispatch;
- multi-image and Files lifecycle tests pass without buffering large content;
- each price permutation has a decimal billing assertion;
- stored output remains pinned to the creating tenant/account.

Rollback:

- independent kill switches for edits, Files persistence, and public URLs while
  leaving ordinary generation available if safe.

### Phase 5 — deferred completion and xAI-native Batch

Owner: async resources + billing

Tasks:

- [ ] Add `GET /v1/chat/deferred-completion/:request_id` and persist ownership
  of the ID returned by `deferred=true` Chat requests.
- [ ] Preserve 202 pending, 200 one-time retrieval, provider errors, and the
  documented 24-hour/one-retrieval lifecycle.
- [ ] Implement xAI Batch semantics rather than assuming OpenAI's input-file
  Batch schema: create, list, get, cancel, add requests, get requests/results,
  and JSONL upload where documented.
- [ ] Add fixed provider operations and public routes for missing xAI Batch
  subresources.
- [ ] Accept `grok-4.3`, both image models, and model-confirmed tools in Batch.
- [ ] Reject `grok-4.5` Batch before provider dispatch.
- [ ] Apply 20% token discount only to `grok-4.3`; image Batch remains standard
  price.
- [ ] Settle completed, partially failed, cancelled, and expired jobs exactly
  once, with idempotent polling and cancellation.
- [ ] Download or persist one-hour signed image results according to documented
  ownership and retention policy.
- [ ] Enforce 25 MB per Batch request and current request-rate constraints
  without confusing provider reference limits with tenant policy.

Exit evidence:

- deferred pending/completed/already-consumed/expired tests;
- Batch state-machine and duplicate-settlement tests;
- mixed Batch tests containing 4.3 Responses, tools, image generation, and image
  edit requests;
- deterministic 4.5 rejection;
- settlement reconciliation for success, partial failure, cancel, and expiry;
- no resource crosses account or tenant boundaries.

Rollback:

- disable deferred and Batch separately; do not affect synchronous inference.

### Phase 6 — billing, observability, resilience, and privacy hardening

Owner: billing + SRE + security/privacy

This phase audits the cross-cutting implementation from phases 1-5. It is not a
license to defer endpoint billing until the end; each earlier capability remains
off until its own rates and accounting pass.

Tasks:

- [ ] Add `xai` to catalog pricing snapshot selection and eliminate static
  registry/zero-price fallback for every target model.
- [ ] Pin estimate and finalize operations to one catalog/profile revision.
- [ ] Add condition-aware component selection for long context, priority,
  Batch, image resolution, and tools.
- [ ] Parse integer cost ticks on Chat, Responses, SSE, deferred, image, and
  Batch results.
- [ ] Reconcile component totals, exact provider cost, tenant sell price, wallet
  debit, spend log, and trace cost without double charging.
- [ ] Fail closed before dispatch when a required price/profile snapshot is
  missing, stale beyond policy, ambiguous, or internally inconsistent.
- [ ] Emit requested model, served model, alias, endpoint, profile/catalog
  revision, provider account, reasoning effort, prompt threshold class, cached
  tokens, effective tier, Batch state, image resolution/count, tool use, exact
  ticks, reconciliation delta, retry count, request ID, and status class.
- [ ] Alert on unpriced usage, missing/malformed ticks, served-model mismatch,
  unexpected alias redirect, reconciliation variance, 401/403, 429 surge, 5xx,
  mid-stream failure, expired media URL, and catalog drift.
- [ ] Add administrative kill switches for each model and for Responses, tools,
  image edits, Files persistence, public URLs, deferred, Batch, and priority.
  Surface current state and the last readiness/reconciliation result to support
  operators without exposing secrets.
- [ ] Implement retry budgets with exponential jitter and `Retry-After`; retry
  only safe reads or provider-documented idempotent creates.
- [ ] Verify sync, stream, deferred, and Batch timeouts and client cancellation.
- [ ] Verify API keys, image data, signed URLs, reasoning blobs, MCP credentials,
  and raw prompts never leak into logs or readiness reports.
- [ ] Define ZDR behavior: under xAI ZDR, reject incompatible stored Responses,
  deferred completions, Files inputs/outputs, URL-format images, and public URLs;
  require base64 image output where provider policy demands it.

Primary billing files:

- [`services/control-plane/internal/billing/service.go`](../services/control-plane/internal/billing/service.go)
- [`services/control-plane/internal/billing/catalog_pricing_snapshot.go`](../services/control-plane/internal/billing/catalog_pricing_snapshot.go)
- [`services/control-plane/internal/billing/pricing.go`](../services/control-plane/internal/billing/pricing.go)
- new xAI pricing and reconciliation tests under
  `services/control-plane/internal/billing`

Exit evidence:

- clock-controlled pricing tests across the 200k threshold;
- decimal-exact billing tests for every rate and condition;
- a redacted provider-vs-GodarAI reconciliation report with no unexplained
  material variance over at least one staging billing day;
- fault-injection report for 429, 5xx, connection reset, mid-stream failure,
  timeout, cancellation, duplicate resource creation, and stale snapshot;
- security/privacy review and synthetic alert evidence.

Rollback:

- disable the affected capability/model/account; never fall back to zero cost.

### Phase 7 — documentation, live conformance, canary, and release

Owner: developer experience + release manager

Tasks:

- [ ] Replace the disabled placeholder xAI Postman collection with model-specific
  requests and assertions; remove unrelated embeddings/video placeholders from
  this four-model scope.
- [ ] Add a credentialed `cmd/xai-readiness` command producing a sanitized JSON
  report with a case for every applicable matrix row.
- [ ] Give the readiness command an explicit environment interface:
  `GODARAI_BASE_URL`, `GODARAI_VIRTUAL_KEY`,
  `GODARAI_XAI_PROVIDER_ACCOUNT`, and an approved spend ceiling. Permit
  `XAI_API_KEY` only for an opt-in direct-provider comparison mode; never write
  it to the report or forward it through GodarAI client requests.
- [ ] Add current xAI examples to GodarAI docs for Chat, Responses, reasoning,
  tools, images, Files, deferred, Batch, pricing, errors, and limitations.
- [ ] Document canonical IDs, aliases, 4.5 Batch rejection, JSON-only edits,
  image non-streaming, cache affinity, priority billing, ZDR restrictions, and
  URL expiry.
- [ ] Add SDK examples using GodarAI's base URL and virtual key without exposing
  provider credentials.
- [ ] Run live conformance once with platform-managed credentials and once with
  tenant BYOK against each available xAI region.
- [ ] Run billing reconciliation, load, cancellation, long-context, multimodal,
  and media-download tests under an approved spend ceiling.
- [ ] Canary one model at a time: `grok-4.3`, `grok-4.5`,
  `grok-imagine-image`, then `grok-imagine-image-quality`.
- [ ] Hold each canary for an agreed observation window and compare exact xAI
  cost, GodarAI ledger, error rate, latency, cache rate, and served identity.
- [ ] Complete a rollback drill before general availability.
- [ ] Schedule a first-party documentation/model/pricing diff job. Changes open
  a reviewed catalog/profile revision; they never mutate a published revision.

Exit evidence:

- sanitized readiness reports with `failed=0` and `blocked=0` for platform and
  BYOK credentials;
- all automated suites green;
- signed billing, SRE, security/privacy, and release approvals;
- no P0/P1 issue open;
- dashboard and alert links;
- completed rollback drill;
- public documentation matches the enabled runtime profile.

Only then may all four models be labeled `production-ready`.

## 7. Test plan

### 7.1 Unit and schema tests

- model-discovery YAML parsing, aliases, runtime profile, and pricing dimensions;
- Go/TypeScript runtime-profile parity;
- provider normalization and FQN parsing;
- credential probe and base-URL validation;
- per-model parameters, roles, modalities, endpoint eligibility, and negative
  combinations;
- usage parsing and response round trip, including unknown safe extension fields;
- cost ticks bounds and conversion;
- condition-aware decimal pricing;
- resource mapping and tenant/account isolation;
- retry classification and header allowlists.

### 7.2 Contract and integration tests

- Control Plane -> Silk Road -> xAI mock for every fixed operation;
- exact upstream method, path, query, body, headers, and content type;
- sync and SSE golden responses;
- Responses state transitions and opaque content preservation;
- deferred and Batch state machines;
- Files streaming, image input/output, public URL lifecycle, and signed URL
  privacy;
- provider error/status/header passthrough;
- trace/log redaction and metrics labels;
- estimate, reserve, finalize, debit, and settlement behavior.

This integration suite is the automated end-to-end validation of GodarAI's
local service chain. Credentialed staging in section 7.3 is the external
end-to-end validation against xAI.

### 7.3 Live readiness matrix

For each text model:

- Models identity/aliases/limits;
- plain text sync/SSE;
- system/user/assistant/tool and provider-confirmed developer-role behavior;
- image understanding by URL and base64;
- all reasoning efforts plus negative combinations;
- JSON object and strict JSON schema;
- sequential and parallel function calls;
- every enabled server tool;
- cached prompt with affinity key;
- short/long-context pricing boundary;
- default/priority effective tier;
- Responses storage, continuation, retrieval, delete/cancel, input tokens, and
  compaction where enabled;
- deferred behavior;
- Batch for 4.3 and negative Batch for 4.5;
- exact ticks and final usage reconciliation.

For each image model:

- generation and edit;
- one and multiple references;
- URL, base64, and file ID input;
- every aspect ratio;
- 1K and 2K;
- one and multiple outputs;
- URL and base64 response;
- stored private output and public URL output;
- Files retrieve/content/delete and public URL revoke;
- Batch generation/edit;
- negative multipart, variations, streaming, invalid expiry, and over-limit
  cases;
- exact ticks and per-image reconciliation.

## 8. Required implementation artifacts

The program is incomplete until the repository contains:

1. four complete xAI model YAML files and a corrected xAI default;
2. an xAI model importer and catalog/profile schema coverage;
3. dedicated xAI adapter, runtime policy, operations, and retry policy;
4. provider-aware Responses/Files/Batch resource routing;
5. xAI usage and exact-cost contracts for sync and streaming;
6. xAI catalog pricing snapshot and reconciliation logic;
7. full automated test fixtures and suites;
8. a credentialed xAI readiness command;
9. a real, enabled-by-credentials Postman collection;
10. developer/user documentation and operational runbooks;
11. dashboards, alerts, and a provider source-drift job;
12. signed release evidence.

## 9. Prioritized backlog

### P0 — required before any production enablement

- canonical `xai` catalog with non-empty model documents;
- fail-closed catalog pricing and exact tick capture for Chat/Responses/images;
- model-specific runtime validation;
- dedicated xAI adapter and safe header/path handling;
- Responses support for xAI without OpenAI-only resource mappings;
- correct image controls and JSON edits;
- credentialed conformance for each model;
- tenant/account isolation for stored provider resources.

### P1 — required before the claim of full model support

- all confirmed server tools and structured output;
- deferred Chat;
- xAI-native Batch, including image requests and 4.5 rejection;
- Files-backed image inputs/outputs and public URL lifecycle;
- priority, caching affinity, long-context, and reasoning accounting;
- resilience, observability, reconciliation, and ZDR enforcement;
- complete documentation and test collection.

### P2 — operational maturity after feature gates are green but before GA sign-off

- automated first-party source drift and alias/lifecycle monitoring;
- provider rate-limit snapshot tooling;
- sustained load and regional comparison;
- daily cost reconciliation and anomaly tuning;
- release/rollback automation.

## 10. Ambiguity and change register

These items must remain explicit until evidence closes them:

| Item | Current treatment | Closure evidence |
|---|---|---|
| Text-model maximum output tokens | unknown; do not invent | current official schema or credentialed model response |
| `developer` message role | probe-gated | current REST schema plus live success/expected rejection per endpoint |
| Exact operation set for Responses resource methods per model | probe-gated | model-specific live matrix and current REST reference |
| Full server-tool set for `grok-4.3` | probe-gated despite Batch examples | successful live matrix and official model/tool evidence |
| MCP/attachment/collection tools for `grok-4.5` | probe-gated | model-specific live matrix |
| Deferred availability for `grok-4.3` | probe-gated | live deferred create/retrieve |
| Image `n` and reference-image maximum | use current official guide values at implementation time | dated schema fixture and boundary tests |
| Region/team rate limits | reference metadata only | credentialed account snapshot; never universalize account quota |
| Apidog image parameters vs current xAI image controls | current xAI wins | record conflict in catalog provenance |
| Provider Files/storage charge export | not assumed available per request | official cost/usage export or invoice reconciliation |

## 11. Completion audit

Before changing the status from Draft/Not production-ready, the release manager
must map every row in sections 3.2, 3.3, 3.4, and 4 to:

- a catalog/profile revision;
- implementation source;
- an automated positive or negative test;
- a live readiness result where the capability incurs provider behavior;
- a billing/reconciliation assertion where the capability can incur cost;
- documentation;
- a rollback control.

Missing, indirect, generic-provider, or planned-only evidence counts as not
complete. A green generic OpenAI-compatible smoke test cannot close a
model-specific row.

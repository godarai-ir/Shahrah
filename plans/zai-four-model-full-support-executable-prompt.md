# Executable Prompt: Z.ai Four-Model Full Support

## Ready-to-use prompt

```md
You are working inside the GodarAI/Shahrah repository as a senior product-minded Go platform engineer. Implement production-gated, first-class direct Z.ai support end to end. Do not stop at analysis, scaffolding, a catalog-only change, or a generic OpenAI-compatible proxy. Make the provider, four model profiles, credentials, public contracts, all documented modalities and parameters, runtime adapters, exact billing, tenant tiers, UI flows, tests, documentation, readiness tooling, canary controls, and rollback path complete within the approved scope.

## Mission

Add exactly these four canonical models from the direct Z.ai general API:

- `glm-5.2`
- `glm-4.7-flashx`
- `glm-5v-turbo`
- `glm-ocr`

Use `zai` as the canonical provider slug. Use `Z.ai` only as the display name. Do not use `xai`, `z-ai`, `z.ai`, `bigmodel`, `zhipu`, or `zhipuai` as alternate provider slugs. `xai` is an existing, unrelated provider and must remain completely isolated.

The public GodarAI model identifier must follow the existing provider-account FQN contract:

`zai:<provider-account>:<model-id>`

Examples:

- `zai:team-zai:glm-5.2`
- `zai:team-zai:glm-4.7-flashx`
- `zai:team-zai:glm-5v-turbo`
- `zai:team-zai:glm-ocr`

The intended user outcome is frictionless, truthful access: an authorized tenant can add or select a Z.ai provider account, discover each supported model, create a deployment, use every documented and verified input modality and parameter, call the correct public endpoint, stream where supported, execute function-tool loops where supported, observe exact usage and cost, and troubleshoot stable errors without learning provider-specific infrastructure details.

“Frictionless” does not permit invented compatibility. When official Z.ai sources conflict, preserve the uncertainty in the readiness plan, choose the safe fail-closed contract, and run a bounded live probe before advertising or enabling the disputed capability.

## Mandatory first steps

1. Read the repository-root `AGENTS.md` and every more-specific `AGENTS.md` governing files you change.
2. Run `git status --short`. Preserve all user changes and unrelated work. Never reset, overwrite, or broadly reformat an existing dirty worktree.
3. Read:
   - `plans/README.md`;
   - `plans/model-catalog-experience-prd.md`;
   - `plans/platform-managed-provider-credentials-provider-matrix.md`;
   - `plans/deepseek-two-model-support-executable-prompt.md`;
   - `plans/moonshot-four-model-production-readiness.md`;
   - `plans/xai-four-model-production-readiness.md`.
4. Inspect the current implementation rather than relying on the plans:
   - `apps/model-discovery/providers` and its import/schema pipeline;
   - `apps/model-discovery/src/catalog/runtime-profile.schema.ts`;
   - provider and credential migrations under `apps/model-discovery/prisma/migrations` and `db/main/migrations`;
   - provider-account contracts and handlers in `packages/contracts/platform` and `services/control-plane/internal/admin`;
   - request validation, FQN resolution, wallet reservation, billing finalization, logs, traces, and route registration in `services/control-plane`;
   - provider registry, credential probe, operation mapping, SSRF controls, SSE behavior, and adapters in `services/silk-road`;
   - current Mistral OCR, Moonshot, DeepSeek, Gemini multimodal, and xAI implementations and tests;
   - tenant-admin provider-account, deployment, model-selection, Playground, logs, and pricing surfaces;
   - public model catalog and Persian docs.
5. Before changing any UI, read completely:
   - `docs/design-system/DESIGN.md`;
   - `docs/design-system/COMPONENT_USAGE.md`;
   - the relevant app-local `AGENTS.md`;
   - existing components in `apps/tenant-admin/components/ui` and `apps/tenant-admin/components/app-ui`.
6. Fetch the current official Z.ai `llms.txt` and `openapi.json`. Store a normalized, reviewed source snapshot or checksum using the repository’s existing provenance mechanism. Generate a machine-readable model/capability/parameter matrix from the current OpenAPI document and compare it with the prose guides.
7. Treat only current first-party Z.ai sources and a separately approved live Z.ai account as provider authority. Do not use model directories, aggregators, third-party gateways, or memory to fill gaps.

## Official sources that must be reverified

- Documentation inventory: https://docs.z.ai/llms.txt
- Machine-readable API contract: https://docs.z.ai/openapi.json
- General API quick start: https://docs.z.ai/guides/overview/quick-start
- Model overview: https://docs.z.ai/guides/overview/overview
- Pricing: https://docs.z.ai/guides/overview/pricing
- Core parameters: https://docs.z.ai/guides/overview/concept-param
- HTTP, API key, JWT, and base URL: https://docs.z.ai/guides/develop/http/introduction
- Chat Completions: https://docs.z.ai/api-reference/llm/chat-completion
- GLM-5.2: https://docs.z.ai/guides/llm/glm-5.2
- GLM-4.7 family: https://docs.z.ai/guides/llm/glm-4.7
- GLM-5V-Turbo: https://docs.z.ai/guides/vlm/glm-5v-turbo
- GLM-OCR: https://docs.z.ai/guides/vlm/glm-ocr
- Layout Parsing: https://docs.z.ai/api-reference/tools/layout-parsing
- Thinking modes: https://docs.z.ai/guides/capabilities/thinking-mode
- Deep Thinking: https://docs.z.ai/guides/capabilities/thinking
- Streaming: https://docs.z.ai/guides/capabilities/streaming
- Tool-call streaming: https://docs.z.ai/guides/capabilities/stream-tool
- Function calling: https://docs.z.ai/guides/capabilities/function-calling
- Context caching: https://docs.z.ai/guides/capabilities/cache
- Structured output: https://docs.z.ai/guides/capabilities/struct-output
- Built-in web search, only for the explicit capability decision: https://docs.z.ai/guides/tools/web-search
- Errors: https://docs.z.ai/api-reference/api-code
- Account-specific rate limits: https://z.ai/manage-apikey/rate-limits
- Release notes: https://docs.z.ai/release-notes/new-released
- Privacy policy: https://docs.z.ai/legal-agreement/privacy-policy
- Terms of use: https://docs.z.ai/legal-agreement/terms-of-use

For every provider-derived fact, record the source URL, `verified_at`, observed OpenAPI version or checksum, effective date when stated, lifecycle, and whether the fact came from prose, OpenAPI, account UI, or a live response. Current official sources override the baseline below. Never rewrite historical catalog pricing rows when facts change; publish a new effective-dated catalog/profile revision.

## Baseline facts observed on 2026-08-17 — reverify before implementation

### Provider and authentication

- General API base: `https://api.z.ai/api/paas/v4/`.
- Primary text and vision endpoint: `POST /chat/completions` relative to that base.
- OCR endpoint: `POST /layout_parsing` relative to that base.
- Authentication is `Authorization: Bearer <API_KEY>`.
- Z.ai also documents locally generated short-lived HS256 JWTs derived from the API key. Do not require tenants to store a transient JWT. If JWT mode is implemented, store only the encrypted source API key and generate short-lived JWTs inside Silk Road.
- `Content-Type: application/json` is required for the four selected model endpoints.
- The documented `Accept-Language` value is `en-US,en`. It controls provider API messages, not model output language. Do not let callers inject arbitrary upstream headers.
- The Coding Plan base path is a separate subscription product intended for supported coding tools. Do not use or resell the Coding Plan endpoint. All GodarAI traffic in this scope must use the general API.
- The official OpenAPI document currently exposes no `GET /models` endpoint. Do not invent one as an upstream credential probe. GodarAI’s public `GET /v1/models` remains its tenant-scoped local catalog.

### Model and endpoint matrix

| Model | GodarAI role | Verified upstream endpoint | Input modalities | Output modalities | Context | Maximum output |
|---|---|---|---|---|---:|---:|
| `glm-5.2` | flagship text/agent model | `POST /chat/completions` | text | text | 1,000,000 tokens | 131,072 tokens |
| `glm-4.7-flashx` | low-cost, high-speed text model | `POST /chat/completions` | text | text | 200K tokens | 131,072 tokens |
| `glm-5v-turbo` | multimodal understanding and visual agent model | `POST /chat/completions` | text, image, video, file/document | text | 200K tokens | 131,072 tokens |
| `glm-ocr` | document layout parsing/OCR | `POST /layout_parsing` | PDF, JPG, JPEG, PNG by URL or documented base64 form | Markdown text, structured layout details, crop/visualization image URLs | not a chat context | provider-defined tokenized result |

Do not list `audio` as a `glm-5v-turbo` input merely because the Chat endpoint summary mentions it generically; the current vision request schema enumerates text, image, video, and file only.

### Text-model contract

For `glm-5.2` and `glm-4.7-flashx`, the current Chat contract includes:

- message roles: `system`, `user`, `assistant`, and `tool`;
- synchronous JSON and SSE streaming with terminal `data: [DONE]`;
- `do_sample`;
- `stream`;
- `thinking.type`: `enabled` or `disabled`;
- `thinking.clear_thinking`;
- `temperature` in `[0.0, 1.0]`;
- `top_p` in `[0.01, 1.0]`;
- `max_tokens` from `1` through the model-specific maximum;
- `tool_stream` where verified;
- caller-defined function `tools`, with no more than 128 functions;
- `tool_choice: auto` only;
- `stop`;
- `response_format.type`: `text` or `json_object`;
- `request_id` of 6–64 characters;
- `user_id` of 6–128 characters without sensitive information.

The OpenAPI schema exposes `reasoning_effort` only for `glm-5.2`, with requested values `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, and `max`. Current compatibility mappings are:

- `none` and `minimal` skip thinking;
- `low` and `medium` map to effective `high`;
- `xhigh` maps to effective `max`;
- `high` and `max` retain their effective meanings.

Expose requested and effective effort separately in logs and model metadata. Do not claim that compatibility aliases are distinct native compute tiers. Reject `reasoning_effort` for `glm-4.7-flashx` unless current official evidence proves support.

The `stop` contract is internally inconsistent: its prose says one stop string, while the OpenAPI constraint currently permits up to four. Fail closed to one stop string until a live contract test or corrected official source proves the larger value.

### Thinking and tool loops

- Thinking is enabled by default for the GLM-5.2 and GLM-4.7 families.
- Interleaved thinking is supported in documented tool loops.
- Standard API preserved thinking is disabled by default because `clear_thinking` defaults to `true`.
- When `clear_thinking: false`, the caller must return complete, unmodified, correctly ordered historical `reasoning_content` in subsequent assistant messages.
- Preserve `reasoning_content` losslessly through synchronous and SSE responses and subsequent tool-loop requests. Never put it in ordinary prompt/response logs or error payloads.
- Function responses include `tool_calls[].id`, `type`, `function.name`, and JSON-string `function.arguments`; tool results use matching `tool_call_id`.
- Current function selection documents `tool_choice: auto` only. Do not silently map `none`, `required`, or a named function to `auto`.
- Parallel function calling is not currently documented. Do not claim or infer it.

### GLM-5V-Turbo multimodal contract

Support every currently documented form and limit:

- text parts: `{ "type": "text", "text": "..." }`;
- image parts: `image_url.url` as HTTPS URL or documented base64 input;
- image formats: JPG, JPEG, PNG;
- maximum image size: 5 MB each;
- maximum image dimensions: 6000 x 6000 pixels;
- maximum images: 150 per request for the GLM-5V family, subject to current live confirmation;
- video parts: `video_url.url`;
- video formats: MP4, MKV, MOV;
- maximum video size: 200 MB each;
- maximum videos: 2 per request for the GLM-5V family;
- file parts: `file_url.url`;
- documented file families include PDF, TXT, Word, JSONL, XLSX, and PPTX;
- maximum files: 50;
- `file_url` cannot be combined with `image_url` or `video_url` in the same request according to the current schema.

The Chat OpenAPI contract for the vision model currently includes `do_sample`, `stream`, `thinking`, `temperature`, `top_p`, `max_tokens`, `tools`, `tool_choice`, `stop`, `request_id`, and `user_id`, but the same schema description says function tools are supported only by the GLM-4.6V series and AutoGLM Phone. The GLM-5V-Turbo model page separately claims function-call capability. Treat this as a blocking documentation conflict for that one capability: implement the lossless contract and tests behind a model-specific feature gate, run a live tool-loop probe, and expose function calling for `glm-5v-turbo` only after the live evidence passes. Do not block its verified text/image/video/file chat features while this one gate remains closed.

Do not expose `response_format` for `glm-5v-turbo` unless the current OpenAPI and a live response both prove it. Do not reuse the text-model parameter list blindly.

If GodarAI accepts uploaded local files rather than public URLs, use the existing secure file facility to create short-lived, provider-fetchable signed URLs. Enforce tenant ownership, content type, size, expiry, malware policy, and SSRF controls. Never expose an internal storage URL or another tenant’s file. Do not invent a Z.ai upstream Files API for these models.

### GLM-OCR contract

The current `layout_parsing` request includes:

- required `model`, fixed to `glm-ocr`;
- required `file`, as supported URL or documented base64 content;
- `return_crop_images`, default `false`;
- `need_layout_visualization`, default `false`;
- `start_page_id`, minimum 1;
- `end_page_id`, minimum 1;
- `request_id`, 6–64 characters;
- `user_id`, 6–128 characters without sensitive information.

The current response includes `id`, `created`, `model`, `md_results`, nested `layout_details`, `layout_visualization`, `data_info`, `usage`, and `request_id`. Preserve these fields losslessly on the Z.ai-compatible public route. The unified GodarAI OCR route may provide normalized convenience fields, but it must not discard provider-native information.

There is an official-source conflict on PDF page limits: the GLM-OCR guide says up to 100 pages, while the current OpenAPI request description says up to 30 pages. Until corrected or proven live, enforce 30 pages per upstream call. If product requires one-request processing up to 100 pages, implement explicit, tested page-range chunking and deterministic result aggregation behind its own gate; preserve ordering, page numbers, partial-failure status, cost, and idempotency. Never silently truncate a document.

Current file constraints are:

- image: at most 10 MB;
- PDF: at most 50 MB;
- formats: PDF, JPG, and PNG in the current official contract;
- output: Markdown, layout detail structures, and optional image URLs.

The prose guide lists several supported languages but does not explicitly guarantee Persian. Do not advertise Persian OCR support until the required Persian evaluation passes.

### Pricing and usage

Current public provider prices are USD per 1 million tokens:

| Model | Uncached input | Cached input | Cache storage | Output |
|---|---:|---:|---:|---:|
| `glm-5.2` | $1.40 | $0.26 | limited-time free | $4.40 |
| `glm-4.7-flashx` | $0.07 | $0.01 | limited-time free | $0.40 |
| `glm-5v-turbo` | $1.20 | $0.24 | limited-time free | $4.00 |

Current `glm-ocr` pricing is:

- input: $0.03 per 1 million input tokens;
- output: $0.03 per 1 million output tokens;
- no documented cache discount.

Convert these rates to exact per-token decimals in the catalog. Never use binary floating-point for authoritative settlement. The phrase “limited-time free” is a promotion without a published end date. Represent it as dated promotional provenance and source drift, not as an eternal contractual zero.

For Chat usage:

- `prompt_tokens` is total input;
- `usage.prompt_tokens_details.cached_tokens` is cache-hit input;
- uncached input is `prompt_tokens - cached_tokens`, validated as nonnegative;
- `completion_tokens` is the output dimension;
- `total_tokens` is a reconciliation field, not an additional billable dimension.

Verify whether reasoning tokens are already included in `completion_tokens`; never charge `reasoning_content` twice. For multimodal input, verify that provider `prompt_tokens` fully includes image, video, and file tokenization and reconcile at least one live sample of each modality.

For OCR, map the returned usage fields exactly and prove provider-dashboard reconciliation. Do not reuse page-based Mistral OCR prices or dimensions.

The context-caching guide contains a generic “usually 50%” statement that conflicts with the exact model prices above. The effective-dated model pricing table wins.

### Rate limits and service tiers

- Z.ai’s public rate-limit link currently resolves to an authenticated, account-specific page rather than a stable public numeric tier table.
- Do not invent RPM, TPM, concurrency, or daily values.
- Store verified account-level limits as provider-account metadata with observation time and source; keep them separate from tenant plan entitlements.
- Parse and honor documented upstream rate-limit headers if current live evidence identifies them.
- Apply GodarAI tenant rate limits and commercial-plan entitlements before upstream dispatch.
- Use adaptive backpressure and bounded queues so a shared platform credential cannot let one tenant starve another.
- The selected model requests do not currently document a `service_tier` parameter. Reject it rather than pretending compatibility with OpenAI tiers.
- Catalog runtime profiles should expose only the verified standard serverless tier unless Z.ai publishes additional request-selectable tiers.

### Errors and termination

Z.ai uses both HTTP status and an inner business error code. Preserve the safe provider code and map it to the stable GodarAI error envelope. Cover at least:

- authentication: `1000`, `1001`, `1003`, `1005`;
- insufficient balance/package: `1113`;
- generic API failures: `1200`, `1230`, `1234`;
- parameters/model/method: `1210`–`1215`;
- permission/offline/nonexistent API: `1220`–`1222`;
- context too long: `1261`;
- safety: `1301`;
- request limit/overload: `1302`, `1305`, and applicable account-limit codes.

SSE can terminate with a `finish_reason` instead of an HTTP error. Handle `stop`, `tool_calls`, `length`, `sensitive`, `model_context_window_exceeded`, and `network_error` distinctly. Never report a truncated, sensitive-blocked, context-overflow, or network-error stream as a normal success. Finalize already reported usage exactly once and release or adjust wallet reservations deterministically.

## Product scope

### Include — P0

- A first-class direct provider named `Z.ai` with canonical slug `zai`.
- Exactly the four model IDs listed in this prompt.
- Tenant BYOK and platform-managed credentials.
- Encrypted provider secret storage, rotation, activation, disablement, safe testing, audit history, and hot-path immutable credential snapshots.
- Official-host base URL validation for `https://api.z.ai/api/paas/v4`.
- Tenant-scoped discovery through GodarAI `GET /v1/models`.
- OpenAI-compatible `POST /v1/chat/completions` for the three Chat models, sync and SSE.
- A lossless Z.ai-compatible Chat surface for official Z.ai SDK migration; it must share authorization, routing, validation, billing, and observability with the canonical handler rather than creating a bypass.
- `POST /v1/layout_parsing` preserving Z.ai’s native GLM-OCR request and response contract.
- `POST /v1/ocr` as the existing GodarAI unified OCR entry point, generalized so model FQN selects Mistral or Z.ai without Mistral-specific assumptions.
- Thinking enable/disable, turn-level control, interleaved reasoning, optional preserved thinking, and model-specific effort mappings.
- Caller-defined function tools and tool streaming where verified per model.
- JSON Object output for the verified text models.
- Automatic context-cache accounting.
- All verified text, image, video, file/document, PDF, and OCR controls listed above.
- Exact fail-closed USD billing, existing IRR conversion/wallet settlement, catalog sell prices, pricing provenance, and tier enforcement.
- Provider error normalization, cancellation, timeout, mid-stream failure, safe retry policy, metrics, traces, logs, alerts, and upstream request correlation.
- Complete tenant-admin, public catalog, Playground, and Persian documentation experiences.
- Automated source-drift checks, bounded live readiness tests, canary, provider/model kill switches, and rollback drills.

### Capability-gated, but implementation-ready

- `glm-5v-turbo` custom function calling, pending resolution of the model-page/OpenAPI conflict.
- OCR processing above 30 pages through explicit page-range chunking, pending live proof up to the guide’s 100-page claim.
- Built-in `web_search` or `retrieval` tools inside Chat. The OpenAPI schemas contain these shapes while adjacent prose says only functions are currently supported. Implement no public promise until per-model live tests and tool-cost accounting pass.
- JWT-derived upstream authorization. Raw API-key Bearer auth is the simple documented path; enable generated JWT only after compatibility, rotation, clock-skew, and expiry tests pass.

### Explicitly exclude from this four-model delivery

- `glm-5.3`; its public API was still described as coming soon on the observation date.
- All other Z.ai text, vision, image-generation, video-generation, ASR, translation, and agent models.
- Coding Plan endpoints, subscriptions, quotas, or provider keys.
- Provider image generation, video generation, audio transcription, agents, web reader, or translation endpoints.
- Tokenizer for the selected models. The current tokenizer OpenAPI enum does not include any of these four IDs.
- A fabricated upstream model-list endpoint.
- Responses API, Assistants API, Batch API, fine-tuning, embeddings, moderation, realtime, or arbitrary provider path proxying.
- Third-party-hosted GLM weights mislabeled as direct Z.ai.
- Arbitrary base URLs, arbitrary upstream headers, or fallback to the generic OpenAI adapter without Z.ai validation.
- Zero-price, another provider’s price, estimated “typical” cache discounts, or unversioned manual pricing.
- Automatic enablement of a newly discovered Z.ai model or capability.

## Required implementation work

### Phase 0 — evidence, conflict register, and scope lock

- Create `plans/zai-four-model-production-readiness.md` with initial status `not production-ready`.
- Add a source manifest, OpenAPI checksum, exact observation date, model matrix, parameter matrix, pricing matrix, error matrix, legal/privacy unknowns, conflict register, product decisions, release gates, canary order, and rollback plan.
- Include the known documentation conflicts from this prompt and any new ones found.
- Confirm whether `zai` already exists in model-discovery data. Add an idempotent provider upsert only if needed; never collide with `xai`.
- Confirm the current runtime-profile and pricing schemas before writing YAML. Extend shared schemas only for capabilities that cannot be represented truthfully.
- Every unsupported or unknown fact must have an explicit state; do not silently omit it in a way that makes the UI imply support.

Exit evidence:

- exactly four direct Z.ai model IDs are in scope;
- every provider fact has first-party provenance;
- each conflict has a safe current behavior and a named live gate;
- no catalog or route can confuse `zai` with `xai`.

### Phase 1 — provider catalog, profiles, importer, and migrations

- Add:
  - `apps/model-discovery/providers/zai/default.yaml`;
  - `apps/model-discovery/providers/zai/glm-5.2.yaml`;
  - `apps/model-discovery/providers/zai/glm-4.7-flashx.yaml`;
  - `apps/model-discovery/providers/zai/glm-5v-turbo.yaml`;
  - `apps/model-discovery/providers/zai/glm-ocr.yaml`.
- Store canonical model ID, display name, lifecycle, release/verification date, modes, modalities, context and output limits, request parameters, tools, transports, exact prices, source URLs, and a model-specific runtime profile.
- Use `chat_completions` for the three Chat models and add a provider-neutral `layout_parsing` endpoint type only if `ocr` cannot preserve the Z.ai native contract. Do not label layout parsing as Mistral-specific.
- Extend modality schemas only as needed to distinguish `file`/document URL behavior without losing current `pdf` and `document` semantics.
- Represent function calling, structured output, thinking, tool streaming, and cache accounting per model. Do not copy one model profile to all four.
- Add a dedicated Z.ai importer and exact allowlist validation. It must reject extra, duplicate, empty, unofficially sourced, incorrectly named, unpriced, incomplete, or schema-incompatible model files.
- Add an idempotent provider migration with display name `Z.ai`, slug `zai`, official docs URL, API-key credential schema, default base URL, enabled modes, and provider capability metadata.
- Add the provider logo/mark through the existing catalog asset pipeline using an official or legally approved asset; do not hotlink a mutable remote asset.
- Add source-drift comparison for model IDs, OpenAPI schemas, limits, tool support, prices, promotion state, and error codes.

Exit evidence:

- exactly four YAML model documents import;
- public catalog data shows the correct modality and endpoint matrix;
- every enabled model has complete required pricing dimensions;
- profile/schema tests fail deterministically on a Z.ai/xAI slug mix-up;
- no missing rate can become zero.

### Phase 2 — provider accounts, credentials, host security, and tenancy

- Add Z.ai to platform-managed and tenant BYOK credential contracts and UI.
- Stored auth mode: `api_key`.
- Inject `Authorization: Bearer <credential>` only inside Silk Road at dispatch time.
- If JWT mode is enabled, generate a short-lived HS256 token inside Silk Road from the encrypted source API key. Handle split/key validation, expiry, clock skew, rotation, and redaction. Never persist generated JWTs.
- Allow only scheme `https`, hostname `api.z.ai`, default port, and the fixed `/api/paas/v4` base path. Reject userinfo, fragments, deceptive suffixes, alternate ports, path traversal, and caller-controlled query strings.
- Do not use the Coding Plan host/path.
- Since no non-billable model-list endpoint is documented, implement credential testing in two layers:
  1. local format, encryption/decryption, host, and connectivity validation without claiming provider authentication;
  2. an explicit, approval-gated, billable live test with a displayed maximum spend and model-specific result.
- The cheapest bounded Chat probe may use `glm-4.7-flashx`, `thinking.disabled`, minimal safe input, and `max_tokens: 1`, but it proves only that model and account path. OCR and other model entitlements need separate optional probes.
- Never run a hidden billable probe on save, page load, health check, or gateway hot path.
- Cover key create/version/activate/disable/rotate flows, concurrent activation, stale snapshots, 401, 403, 429, insufficient balance, timeout, malformed response, redaction, and audit events.
- Derive upstream `user_id` from an approved stable HMAC of tenant, provider account, and optional end-user identifier. Never forward email, phone, tenant slug, IP address, or raw user-supplied PII. Enforce 6–128 characters.
- Generate or validate an upstream `request_id` of 6–64 characters from the GodarAI correlation ID while preserving a separate public correlation ID.
- Prove two tenants sharing one platform-managed credential receive distinct opaque `user_id` values and cannot observe one another’s provider-cache, file, trace, or billing metadata.

### Phase 3 — public routes and typed API contracts

- Keep GodarAI authentication and tenant isolation authoritative on every route.
- Add or generalize typed contracts in `packages/contracts` for:
  - Z.ai Chat request/response extensions;
  - multimodal text/image/video/file content parts;
  - thinking and preserved `reasoning_content`;
  - requested/effective reasoning effort;
  - function tools and streamed tool-call deltas;
  - cache usage details;
  - layout-parsing request/response, layout details, data info, and usage;
  - partial/chunked OCR aggregation if enabled;
  - stable provider error details.
- Preserve backward compatibility for current OpenAI-compatible clients.
- Public routes:
  - `GET /v1/models`: tenant-scoped GodarAI catalog only;
  - `POST /v1/chat/completions`: canonical Chat route for `glm-5.2`, `glm-4.7-flashx`, and `glm-5v-turbo`;
  - `POST /v1/layout_parsing`: lossless Z.ai-compatible GLM-OCR contract;
  - `POST /v1/ocr`: provider-neutral OCR route using the same pipeline and billing, with no duplicate dispatch;
  - `POST /api/paas/v4/chat/completions`: Z.ai SDK-compatible alias using a GodarAI credential and GodarAI FQN, fully sharing the canonical Chat pipeline;
  - `POST /api/paas/v4/layout_parsing`: Z.ai SDK-compatible alias using a GodarAI credential and GodarAI FQN, fully sharing the canonical layout-parsing pipeline.
- Do not create an unauthenticated compatibility bypass.
- Publish the route and parameter matrix in the generated OpenAPI for GodarAI.
- Reject unknown fields or unsupported combinations according to the repository’s current compatibility policy. Never silently strip a parameter that changes cost, safety, modality, reasoning, tools, or output shape.
- Return stable OpenAI-like errors on the OpenAI-compatible route and the documented Z.ai shape on the Z.ai-compatible route, while retaining one internal normalized error.

### Phase 4 — model-specific validation and multimodal preprocessing

- Validate before wallet reservation and provider dispatch:
  - provider, account, deployment, FQN, model, endpoint, and transport eligibility;
  - message roles and nonempty user input;
  - context and model-specific output limits;
  - exact parameter type, range, default, and cross-field rules;
  - thinking/effort compatibility;
  - tool count, schema size, JSON validity, `tool_choice`, and tool-stream support;
  - structured output only where verified;
  - stop-string count under the current safe contract;
  - every multimodal part count, media type, byte limit, pixel limit, and mutual-exclusion rule;
  - URL scheme and signed-file ownership;
  - OCR page range, ordering, size, format, and visualization flags;
  - `request_id` and opaque `user_id` constraints.
- Do not download large remote media in the Control Plane hot path. If inspection or signed-URL conversion is required, use bounded streaming, content-length limits, timeouts, MIME sniffing, and cancellation.
- Do not trust filename extensions alone.
- Prevent SSRF if any GodarAI component fetches a caller URL: block loopback, link-local, private, metadata, unix socket, non-HTTP(S), redirect-to-private, DNS rebinding, and oversized decompression paths.
- Do not log base64 input, URLs containing secrets, file contents, OCR document text, or extracted layout text.
- Estimate the wallet reservation conservatively for unknown multimodal tokenization without making normal requests impossible. Reconcile and release the difference after final usage.

### Phase 5 — dedicated Silk Road adapter and transport behavior

- Add a dedicated `zai` adapter and register it explicitly. Do not route through the generic OpenAI-compatible adapter alone.
- Fixed upstream operations in this scope:
  - `POST /chat/completions`;
  - `POST /layout_parsing`.
- Use the canonical general base URL, fixed paths, Bearer auth, approved headers, and provider-specific validation.
- Never forward the customer’s GodarAI `Authorization` header upstream.
- Preserve synchronous Chat responses, SSE deltas, reasoning deltas, tool-call fragments, finish reasons, usage, model identity, request IDs, OCR layout structures, and visualization links without lossy translation.
- SSE handling must cover comments/blank lines if observed, multi-line events, UTF-8 boundaries, partial JSON chunks, `[DONE]`, optional final usage, client cancellation, provider disconnect, slow consumers, and mid-stream terminal errors.
- Reassemble streamed `tool_calls[].function.arguments` by index and ID without corrupting JSON.
- Support `tool_stream` as provider-defined incremental argument delivery rather than pretending ordinary SSE automatically proves it.
- Retry only operations proven safe. Do not retry Chat or OCR after an unknown billable outcome. A pre-dispatch connection failure may be retried only if the transport can prove no request body was accepted, under a small bounded policy with jitter.
- Preserve approved upstream request IDs and retry hints. Strip cookies, internal infrastructure headers, secrets, and unsafe provider bodies.
- Enforce response-size ceilings for sync, SSE, OCR Markdown, layout arrays, and returned visualization URLs.

Exit evidence:

- adapter tests prove exact method, URL, path, headers, body, and content type;
- unsupported operations cannot reach Z.ai;
- sync, SSE, reasoning, tool calls, media, OCR, cancellation, timeout, and mid-stream failures have golden tests;
- no secret or content appears in adapter logs.

### Phase 6 — functions, thinking, structured output, and cache continuity

- Implement complete custom function-tool loops for `glm-5.2` and `glm-4.7-flashx`:
  - one and multiple sequential tool calls;
  - sync and SSE;
  - streamed arguments when `tool_stream` is enabled;
  - tool error/result messages;
  - matching `tool_call_id`;
  - preserved interleaved `reasoning_content` when requested;
  - malformed JSON arguments as model output, not trusted executable input.
- Never execute user-defined tools inside GodarAI. GodarAI transports schemas, calls, and results; the caller executes the function unless a separately approved server-tool product exists.
- Validate function name, description, parameters schema, aggregate schema bytes, nesting/complexity limits, and maximum 128 tools. Reuse repository-wide safe JSON Schema limits.
- Implement thinking defaults and turn-level enable/disable exactly per model.
- Implement `clear_thinking: false` without rewriting, trimming, logging, or reordering prior reasoning blocks.
- Implement the `glm-5.2` effort mapping and expose effective effort. Reject it elsewhere until verified.
- Support `response_format: {"type":"json_object"}` for verified text models. Validate that the response is syntactically valid JSON before marking structured-output success; preserve provider content for diagnostics without exposing sensitive bodies.
- Account for automatic cache hits using returned usage. Add a two-request readiness case proving cache miss then eligible hit.
- Never let the cache or preserved-reasoning implementation cross tenant, credential, model, or deployment boundaries.

### Phase 7 — pricing, wallet settlement, plans, and provider tiers

- Add a Z.ai catalog-pricing snapshot manager and model-specific completeness checks.
- Estimate, reserve, dispatch, finalize, trace, and debit using one immutable catalog/profile revision per request.
- Required Chat pricing dimensions:
  - uncached input;
  - cached input;
  - output;
  - cache storage promotional metadata, even when current rate is zero.
- Required OCR dimensions:
  - OCR input tokens;
  - OCR output tokens.
- Use exact decimal/tick arithmetic throughout. No `float64` for authoritative provider cost, sell price, IRR conversion, or wallet writes.
- Reconcile requested model, served model, usage, exact provider cost, GodarAI sell price, margin, exchange-rate snapshot, IRR wallet reservation, IRR final debit, spend log, and trace cost.
- Fail before dispatch when a price is missing, stale beyond policy, ambiguous, negative, internally inconsistent, missing provenance, or incompatible with the runtime profile.
- Fail finalization and alert if cached tokens exceed prompt tokens, total tokens do not reconcile, the served model differs, OCR usage is absent, or a pricing dimension is unknown. Do not silently charge the estimate as exact.
- Integrate with current commercial plans and tenant entitlements. Provider account limits, GodarAI plan limits, and per-key limits are separate layers and must be visible as such.
- Store account-specific Z.ai limit observations without treating them as universal model tiers.
- Use fair per-tenant concurrency for shared credentials and surface actionable `429` responses with reset/retry metadata when available.
- Add alerts for unpriced use, zero-cost fallback, pricing drift, promotion changes, missing usage, served-model mismatch, cache reconciliation failure, repeated 1113 balance errors, 1302/1305 spikes, and wallet/provider variance.

Pricing tests must assert exact costs for each of the four models, every pricing dimension, cache miss/hit mixtures, maximum output reservations, multimodal usage, OCR input/output, IRR conversion, and idempotent finalization.

### Phase 8 — catalog, tenant-admin, Playground, and user experience

- Add Z.ai to provider-account creation and edit flows with:
  - official display name and mark;
  - API key field with masked readback and rotation;
  - fixed official base URL shown as non-editable unless current product policy exposes an advanced validated field;
  - clear BYOK versus platform-managed state;
  - explicit billable connection-test confirmation and spend ceiling;
  - model-specific test results;
  - upstream account-limit state as verified/unknown rather than invented.
- Make all four models discoverable in public catalog and tenant model selectors with accurate:
  - model identity and lifecycle;
  - endpoint modes;
  - input/output modalities;
  - context and output limits;
  - tools and structured-output capability;
  - thinking defaults and effort values;
  - pricing dimensions and currency unit;
  - verified/conditional capability labels.
- Extend the Playground so a user can complete without manual JSON editing:
  - text Chat, sync and streaming;
  - thinking toggle and `glm-5.2` effort selector;
  - function schema editor, tool-call display, tool-result continuation, and tool streaming;
  - JSON Object output;
  - image upload/URL/base64 within limits;
  - video URL or secure upload within limits;
  - file/document selection within formats and count limits;
  - OCR upload/URL, page range, crop images, layout visualization, Markdown result, layout detail viewer, and partial/chunked status if applicable;
  - usage, cached tokens, estimated/final USD cost, and IRR charge.
- Reuse existing components and semantic Tailwind tokens, preserve RTL, accessibility, keyboard operation, loading/empty/error states, cancellation, and narrow-screen layout.
- Update the design-system playground only when adding a reusable UI pattern.
- Never render raw reasoning content as ordinary logs. If the product intentionally exposes it to the initiating user, require a separate privacy-reviewed surface and retention policy.

### Phase 9 — documentation, SDK compatibility, and examples

- Add Persian provider documentation under `apps/docs/docs/providers/zai.mdx` and register it in navigation.
- Follow the GodarAI technical-writing skill and repository terminology.
- Document:
  - adding BYOK and using a platform-managed account;
  - all four FQNs;
  - public routes and required scopes;
  - text sync/SSE examples;
  - thinking, effort mapping, preserved reasoning, and function loops;
  - JSON Object output;
  - image, video, and file examples with exact limits;
  - OCR request, page range, Markdown/layout response, and large-PDF behavior;
  - exact provider-cost dimensions and IRR display caveat;
  - account-specific rate limits and tenant plan limits;
  - errors and finish reasons;
  - unsupported/deferred features;
  - privacy, retention unknowns, and troubleshooting.
- Provide small, equivalent examples for REST, OpenAI Python SDK, OpenAI Node SDK, and official Z.ai SDK only where the public GodarAI route is genuinely compatible.
- Use `YOUR_GODARAI_TOKEN`, a GodarAI base URL, and Z.ai FQNs. Never show an upstream Z.ai key in ordinary usage examples.
- Add/update the GodarAI OpenAPI document and Postman collection if the repository owns them. Keep request/response examples free of user content and secrets.
- Add deep links from model catalog details to the correct provider guide, pricing section, and troubleshooting section.

### Phase 10 — observability, readiness, Persian evaluation, canary, and rollback

- Emit low-cardinality fields for provider, provider account ID, requested/served model, endpoint, modality summary, runtime-profile revision, pricing revision, thinking enabled, requested/effective effort, tool usage count, cached/uncached/output tokens, OCR tokens/pages, retry count, status class, finish reason, latency, and upstream request ID.
- Never log provider keys/JWTs, raw prompts/responses, `reasoning_content`, function arguments/results, base64 media, signed URL query strings, document contents, OCR text, or raw user IDs.
- Add dashboards and alerts for latency, time to first token, stream completion, errors by safe code, 401/403, insufficient balance, 429, 5xx, cancellations, malformed SSE, missing usage, served-model mismatch, source drift, pricing drift, and billing variance.
- Add a bounded `services/control-plane/cmd/zai-readiness` command following existing provider readiness patterns. It must:
  - call GodarAI by default, not bypass billing/routing;
  - accept a virtual key and provider-account identifier;
  - require explicit approval and a maximum USD spend before billable cases;
  - never accept, print, or persist a provider key in ordinary mode;
  - allow separately gated direct-provider comparison through `ZAI_API_KEY` without including it in output;
  - emit a sanitized JSON report with per-model/per-capability status and evidence references;
  - remain `production_ready: false` until all human gates are attached.
- Live matrix for `glm-5.2`:
  - sync/SSE, thinking on/off, every requested/effective effort mapping, JSON Object, sequential and multi-call tool loops, tool streaming, preserved reasoning, cache miss/hit, maximum safe context sample, errors, cancellation, served identity, and billing reconciliation.
- Live matrix for `glm-4.7-flashx`:
  - sync/SSE, thinking on/off, rejection of unsupported effort, JSON Object, sequential tool loop, tool streaming, cache miss/hit, errors, cancellation, served identity, and billing reconciliation.
- Live matrix for `glm-5v-turbo`:
  - text-only, image URL, image base64, multiple images, video, each file family, signed upload flow, format/count/size negatives, mutual-exclusion negatives, sync/SSE, thinking, cache usage, function-tool conflict probe, served identity, and billing reconciliation.
- Live matrix for `glm-ocr`:
  - JPG, PNG, PDF URL/base64 as documented, Markdown, tables, formulas, handwriting, crop images, visualization, page ranges, 30-page boundary, optional chunk aggregation, malformed and oversized files, served identity, usage, and billing reconciliation.
- Required Persian product evaluation:
  - Persian instruction following and JSON fields for both text models;
  - Persian function arguments and tool results;
  - mixed Persian/English code and RTL content;
  - Persian screenshots and UI analysis for `glm-5v-turbo`;
  - printed Persian, Persian tables, invoices, receipts, identity-document-like synthetic samples, mixed Persian/English text, digits, dates, and handwriting for `glm-ocr`;
  - no real identity documents or personal data in test fixtures.
- Do not advertise Persian support for a model/capability until its explicit acceptance thresholds pass.
- Canary in this order with independent gates: `glm-4.7-flashx`, `glm-5.2`, `glm-5v-turbo`, `glm-ocr`.
- Add provider-wide and per-model kill switches. Test rollback without deleting catalog history, usage, pricing snapshots, provider accounts, or audit evidence.

## Required automated tests

At minimum, add or extend tests for:

1. Z.ai provider migration/upsert, slug isolation from xAI, credential schema, and no duplicates.
2. Exactly four model YAML files, official source validation, OpenAPI provenance/checksum, lifecycle, modalities, endpoint matrix, limits, and pricing completeness.
3. Import rejection for missing/extra models, bad source domain, missing rates, unknown endpoint, unsupported parameter, and profile mismatch.
4. BYOK/platform-managed create, rotate, activate, disable, snapshot refresh, redaction, audit, and shared-credential tenant isolation.
5. Official-host/base-path validation, deceptive hosts, ports, userinfo, fragments, redirects, and arbitrary-path rejection.
6. Explicit billable credential-test approval, spend ceiling, no hidden calls, and per-model result semantics.
7. FQN resolution, deployment selection, local `GET /v1/models`, tenant isolation, and catalog UI data.
8. Text Chat parameter ranges/defaults, unsupported combinations, model-specific effort mapping, stop conflict behavior, and request/user ID handling.
9. Sync/SSE round trips, `[DONE]`, tool-call fragments, reasoning fragments, usage, finish reasons, disconnects, cancellation, slow consumers, and size ceilings.
10. Function tools, 128-tool boundary, schema limits, `tool_choice: auto`, tool results, interleaved/preserved reasoning, and no parallel-tool claims.
11. JSON Object output validity and negative behavior.
12. Cache-token parsing, uncached subtraction, malformed usage, and exact mixed-rate billing.
13. `glm-5v-turbo` image/video/file formats, URL/base64 support, counts, sizes, pixels, mutual exclusion, secure signed URLs, ownership, expiry, and SSRF.
14. The model-specific function gate for `glm-5v-turbo` and evidence-controlled activation.
15. GLM-OCR request/response preservation, page ranges, 30-page fail-closed limit, optional chunk aggregation, partial failure, visualization, layout details, and no silent truncation.
16. Exact decimal provider cost, GodarAI sell price, IRR conversion, wallet reservation/finalization, idempotency, served-model mismatch, and zero-price fail closed for all four models.
17. Upstream HTTP/business-code normalization and SSE `finish_reason` failure semantics.
18. Fair shared-account limiting, GodarAI plan entitlements, account-specific limit metadata, and 429 recovery.
19. Secret/content/reasoning/media/OCR redaction in logs, traces, audits, readiness output, and UI errors.
20. Tenant-admin/catalog/Playground behavior, accessibility, RTL, loading/error/cancel states, documentation links, docs build, and generated OpenAPI examples.

Do not make live or billable provider calls in CI.

## Validation commands

Use `pnpm`, never `npm`.

Determine exact scripts from the current package manifests, then run at minimum:

- model-discovery lint, typecheck, tests, import dry-run, and build;
- tenant-admin lint, typecheck, tests, and build;
- model-discovery-web lint, typecheck, tests, and build;
- docs lint, link validation, and build;
- any Postman/OpenAPI contract validation owned by this repository.

For every changed Go file, run `gofmt` and `goimports`.

From every changed Go module, including as applicable `packages/contracts`, `services/control-plane`, and `services/silk-road`, run:

- `go vet ./...`
- `go test ./...`
- `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`

Do not weaken `.golangci.yml`, skip errors, add broad `nolint`, or change unrelated code to force green checks. If a repository-wide pre-existing failure remains, prove it is pre-existing, report it precisely, and still run every focused check for touched code.

Render and inspect all changed UI surfaces at desktop and narrow widths. Verify RTL, keyboard navigation, focus management, file controls, streaming updates, tool calls, OCR result layout, long FQNs, long errors, and cost display.

## Production release gates

Do not mark Z.ai or any model production-ready until its independent gates pass:

1. Current official sources, OpenAPI checksum, prices, limits, and lifecycle approved.
2. Documentation conflicts resolved by corrected official evidence or bounded live proof.
3. Dedicated Z.ai account, approved live-test spend, and general-API entitlement confirmed.
4. BYOK and platform-managed credential flows, rotation, redaction, and tenant isolation approved.
5. Full model-specific live matrix green through the real GodarAI Control Plane -> Silk Road -> approved European egress path -> Z.ai flow.
6. Every advertised modality, parameter, function behavior, stream behavior, and public contract verified.
7. Exact token/cache/OCR usage reconciled against provider account records and GodarAI billing.
8. Exact USD provider cost, sell price, FX snapshot, and IRR wallet debit reconciled.
9. Rate-limit/backpressure behavior proven for a shared credential without cross-tenant starvation.
10. Security/privacy review covers credentials, JWT mode if enabled, opaque identity, remote URLs, uploads, SSRF, media retention, reasoning, OCR content, and logs.
11. Persian evaluation passes before any Persian-support claim.
12. Dashboards, alerts, on-call ownership, source-drift process, kill switches, canary, and rollback drill complete.
13. Product, engineering, security/privacy, finance, and operations approvals attached.

One model’s readiness never unlocks another. A green text Chat probe does not prove vision, OCR, tools, cache billing, or platform-managed isolation.

## Definition of done

The task is complete only when:

- direct Z.ai is a real, distinct provider across catalog, credentials, routing, billing, UI, docs, and operations;
- exactly the four target models are importable, discoverable, selectable, deployable, callable, billable, observable, documented, and independently gated;
- all verified modalities and parameters work end to end without hidden JSON workarounds;
- text sync/SSE, thinking, effort, functions, tool streaming, JSON output, preserved reasoning, and cache accounting are lossless;
- `glm-5v-turbo` handles every verified text/image/video/file form and limit safely;
- `glm-ocr` preserves its complete layout-parsing contract, never truncates documents, and bills returned usage exactly;
- users can configure credentials and exercise each capability from the documented API and Playground;
- provider account limits, tenant plan tiers, and pricing revisions remain distinct and accurate;
- no request can fall back to zero or another provider’s price;
- no secret, reasoning, media, document, tool payload, or cross-tenant identity leaks;
- all automated validation passes;
- readiness remains false for unresolved human/live gates rather than overstating support.

## Required final report

Return:

1. Outcome summary by model and endpoint.
2. Verified provider facts, GodarAI product decisions, unresolved conflicts, and assumptions in separate sections.
3. Complete model/capability/parameter/modality matrix with `verified`, `gated`, or `unsupported` status.
4. Files and migrations changed by subsystem.
5. Public routes, DTOs, FQNs, credential modes, scopes, and error contracts.
6. Pricing dimensions, exact rates, effective dates, cache handling, provider-account limits, GodarAI tiers, and billing reconciliation evidence.
7. Test, lint, typecheck, build, import, migration, OpenAPI, docs, and browser-verification results with exact commands.
8. Live tests executed, spend used, sanitized evidence, and tests not run with reasons.
9. Security/privacy and Persian-evaluation results.
10. Remaining gates labeled `BLOCKER`, `GATED`, `DEFERRED`, or `FOLLOW-UP`, plus current readiness for each model.
```

## Why this prompt is intentionally strict

The four models do not share one uniform contract. The two text models, the multimodal Chat model, and the OCR model use different modalities, parameters, output shapes, and billing assumptions. Current Z.ai documentation also contains material conflicts around GLM-5V-Turbo function tools, OCR page limits, and stop-sequence count. The prompt therefore requires a dedicated provider adapter, model-specific runtime profiles, lossless native OCR support, source provenance, and live capability gates instead of treating Z.ai as only a base-URL change.

## Evidence used to create this prompt

Verified from official Z.ai sources on 2026-08-17:

- [Documentation inventory](https://docs.z.ai/llms.txt)
- [OpenAPI contract](https://docs.z.ai/openapi.json)
- [Model overview](https://docs.z.ai/guides/overview/overview)
- [Pricing](https://docs.z.ai/guides/overview/pricing)
- [HTTP API and authentication](https://docs.z.ai/guides/develop/http/introduction)
- [Chat Completions](https://docs.z.ai/api-reference/llm/chat-completion)
- [GLM-5.2](https://docs.z.ai/guides/llm/glm-5.2)
- [GLM-4.7 family](https://docs.z.ai/guides/llm/glm-4.7)
- [GLM-5V-Turbo](https://docs.z.ai/guides/vlm/glm-5v-turbo)
- [GLM-OCR](https://docs.z.ai/guides/vlm/glm-ocr)
- [Layout Parsing](https://docs.z.ai/api-reference/tools/layout-parsing)
- [Thinking modes](https://docs.z.ai/guides/capabilities/thinking-mode)
- [Function calling](https://docs.z.ai/guides/capabilities/function-calling)
- [Context caching](https://docs.z.ai/guides/capabilities/cache)
- [Structured output](https://docs.z.ai/guides/capabilities/struct-output)
- [Errors](https://docs.z.ai/api-reference/api-code)

Shahrah-specific constraints were derived from the current model-discovery runtime-profile schema, provider catalog, provider-account credential matrix, Control Plane routes and billing code, Silk Road adapters, Mistral OCR support, Gemini multimodal support, and the existing DeepSeek, Moonshot, and xAI readiness plans.

# Executable Prompt: xAI Grok Imagine Video Generation with an Extensible Video Layer

## Ready-to-use prompt

```md
You are working inside the GodarAI/Shahrah repository as a senior Go platform engineer who specializes in LLM gateways and is an expert on xAI's Grok Imagine video API. Implement production-gated, first-class support for xAI video generation end to end: public API, tenancy, async job lifecycle, billing, catalog, tenant-admin UX, docs, readiness tooling, canary, and rollback. Do not stop at analysis, scaffolding, a catalog-only change, or a passthrough proxy.

Build it on a provider-neutral video-generation layer, so that later providers (Veo on Google's enterprise platform, Gemini Omni, Kling, Wan, Seedance) can each be added as one adapter, one runtime profile, and one pricing rule, with no change to public routes, the job store, the settlement worker, or the tenant-admin UI. Do not implement those providers now. Prove the seam is real with a test-only fake second provider.

## Why xAI, not Sora or Veo (verified 2026-10-31)

- OpenAI removed the Videos API and every `sora-2*` model on 2026-09-24, with no replacement (https://developers.openai.com/api/docs/deprecations). The openai/openai-openapi spec still lists `/videos` schemas; ignore it for provider selection.
- Google shut down the Gemini API Veo 3.1 preview models (`veo-3.1-generate-preview`, `-fast-`, `-lite-`) on 2026-10-22 (https://ai.google.dev/gemini-api/docs/video).
- xAI's video API is live and documented, and `xai` is already a first-class provider in this repository.

## Mission

Add exactly these canonical models from the direct xAI API:

| Model | Role | Price (Models page) |
|---|---|---:|
| `grok-imagine-video-1.5` | flagship: native 1080p T2V/I2V, references, voices, pinned frames | $0.080 / s |
| `grok-imagine-video-1.5-lite` | lowest cost T2V/I2V, 1080p upscaled from 720p | $0.020 / s |
| `grok-imagine-video` | the only model with video **edits** and **extensions**; up to 720p | $0.050 / s |

Use `xai` as the provider slug (`services/control-plane/internal/providerregistry/registry.go`). Public model identifiers follow the existing FQN contract `xai:<provider-account>:<model-id>`, for example `xai:team-xai:grok-imagine-video-1.5`.

User outcome: an authorized tenant can select an xAI provider account, discover all three models, create a deployment, and then:

1. generate a video from a text prompt (text-to-video);
2. generate a video from a prompt plus a first-frame image (image-to-video);
3. use reference images, first/last frames, and keyframes where the model supports them;
4. edit and extend a completed video with `grok-imagine-video`;
5. poll status and progress, list their own jobs, and delete them;
6. download the finished MP4;
7. see an exact reservation at creation, exact settlement on completion, a full release where xAI does not charge, and stable, documented errors.

Two client styles must both work unchanged:

- **OpenAI-style:** the official OpenAI SDK `client.videos.create / retrieve / download_content / list / delete`, pointed at the GodarAI base URL. This is the canonical public shape for every video provider.
- **xAI-native:** the xAI REST shape (`POST /v1/videos/generations`, `GET /v1/videos/{request_id}`, `POST /v1/videos/edits`, `POST /v1/videos/extensions`), and the xAI SDK where its base URL is configurable. The control plane already serves xAI-native routes (`/v1/language-models`, `/v1/image-generation-models`, `/v1/files:initialize`), so this follows house practice.

"Complete" does not permit invented compatibility. When sources conflict, record the conflict, pick the fail-closed contract, and run a bounded live probe before enabling or advertising the disputed capability.

## Mandatory first steps

1. Read the repository-root `AGENTS.md`, `services/control-plane/AGENTS.md`, `services/silk-road/AGENTS.md`, and every more-specific `AGENTS.md` that governs files you change. Use `pnpm`, never `npm`. Follow the Go lint rules exactly.
2. Run `git status --short`. Preserve all user changes. Never reset, overwrite, or broadly reformat a dirty worktree. Work on a feature branch and commit with explicit pathspecs only (never stage unrelated files such as `ua-*.pem`).
3. Read `plans/README.md`, `plans/model-catalog-experience-prd.md`, `plans/platform-managed-provider-credentials-provider-matrix.md`, `plans/xai-four-model-production-readiness.md`, and `plans/zai-four-model-full-support-executable-prompt.md` (the rigor bar).
4. Inspect the current implementation rather than relying on plans. These starting points were verified on 2026-10-31:
   - **Public routes:** `services/control-plane/internal/server/routes_proxy.go` registers no `/v1/videos*` routes. This is the primary gap. xAI-native routes are registered near lines 218–223.
   - **Closest handler precedents:**
     - `services/control-plane/internal/proxy/images_handler.go`: pipeline, budgeter, usage logger, multipart handling, `providerNativeResourceStore`, and xAI image billing.
     - `services/control-plane/internal/proxy/xai_deferred_handler.go`: `HandleXAIDeferredGet`, xAI's existing poll-by-`request_id` pattern.
   - **Tenant-scoped async ownership:** `platform.ProviderNativeResourceMapping` (`packages/contracts/platform/types.go:728`) already carries `PublicID`, a hidden `UpstreamID`, `Status`, `ReservationID`, `PricingRevision`, `EstimatedUsage`, and `ExpiresAt`. Its store interface is in `services/control-plane/internal/proxy/responses_operations_handler.go:28`. Reuse it with `resource_kind = "video"`. Do not create a parallel ownership table unless you prove a concrete gap.
   - **xAI file ownership:** `services/control-plane/internal/proxy/xai_file_references.go` (`rewriteXAIFileReferences`) already maps tenant-visible file IDs to upstream xAI file IDs. Reuse it for every `file_id` in video requests.
   - **Async settlement precedents:** `mistral_batch_worker.go` (`RunBatchWorker` settles each reservation exactly once) and `anthropic_batch_worker.go`.
   - **Silk Road xAI adapter:** `services/silk-road/pkg/server/xai.go`, including `Dispatch`, `DispatchRaw`, `dispatchWithRetry`, `validateXAIRuntimeProfile`, `validateXAIImageRequestBody`, `validateXAIImageStorageOptions`, and `xaiRuntimeEndpoint`. It has no video operations today.
   - **Operation contracts:** the `ProviderOperationXAI*` constants in `packages/contracts/platform/types.go:527-577`. Add the xAI video operations there.
   - **Billing:**
     - `services/control-plane/internal/billing/provider_pricing_strategy.go` (`xAIPricingStrategy`, `reconcileXAIProviderCost`).
     - `pricing_strategy_images.go`, which already bills xAI images from the provider's `cost_in_usd_ticks` and fails closed when that field is missing.
     - `catalog_pricing_snapshot.go` and `credit_rate_card.go`.
     - There is no video pricing today.
   - **Catalog:**
     - `apps/model-discovery/providers/xai/*.yaml`. `grok-imagine-image-quality.yaml` is the template: aliases, params, costs, sources, `runtimeProfile`.
     - `platform.ModelTypeVideoGeneration`.
     - The runtime-profile endpoint family `videos` and output modality `video`, in `packages/contracts/platform/model_runtime_profile.go` and `apps/model-discovery/src/catalog/runtime-profile.schema.ts`.
   - **Readiness:** `services/control-plane/cmd/xai-readiness` already probes xAI image endpoints. Extend it rather than adding a new command.
   - **UI:** `apps/model-discovery-web/components/provider-models-list.tsx:773-775` shows `/v1/videos/generations` for video models. That is correct for the xAI-native route. Make the snippet provider-aware, and show `POST /v1/videos` as the canonical example.
   - **Dead Sora code:**
     - Silk Road maps `openai_videos_*` operations to OpenAI `/videos` (`services/silk-road/pkg/server/openai_operations.go:485-513`).
     - The `openai` registry entry declares `CapabilityVideos`.
     - Both target an API that no longer exists. Remove the capability and the dead operations, with their tests, so nothing can route a video request to OpenAI. Keep any enum that the stored-data or migration history still needs.
   - **Not in scope:** video *input* to chat (Z.ai `glm-5v-turbo`, Moonshot) is a separate, existing feature. Do not touch or conflate it.
5. Before any UI change, read all of `docs/design-system/DESIGN.md`, `docs/design-system/COMPONENT_USAGE.md`, and the existing components in `apps/tenant-admin/components/ui` and `apps/tenant-admin/components/app-ui`. Reuse them. Use semantic tokens. Preserve RTL.
6. Re-fetch the sources below as Markdown (append `.md` to any `docs.x.ai/developers/...` URL; `https://docs.x.ai/llms.txt` is the index). Store normalized snapshots or checksums using the repository's existing provenance mechanism. Build a machine-readable model × endpoint × parameter × enum matrix, and diff it against the baseline below.
7. Treat only first-party xAI sources and a separately approved live xAI account as authority. Never fill gaps from memory, aggregators, or third-party gateways.

## Official sources that must be reverified

- Index: https://docs.x.ai/llms.txt
- REST reference, videos: https://docs.x.ai/developers/rest-api-reference/inference/videos.md
- Video overview and capability matrix: https://docs.x.ai/developers/model-capabilities/video/overview.md
- Generation, polling, and errors: https://docs.x.ai/developers/model-capabilities/video/generation.md
- Image-to-video: https://docs.x.ai/developers/model-capabilities/video/image-to-video.md
- Reference-to-video, frames, and voices: https://docs.x.ai/developers/model-capabilities/video/reference-to-video.md
- Editing: https://docs.x.ai/developers/model-capabilities/video/editing.md
- Extension: https://docs.x.ai/developers/model-capabilities/video/extension.md
- Models and pricing: https://docs.x.ai/developers/models.md and the per-model pages under `/developers/models/`
- ZDR video storage: https://docs.x.ai/build/settings/zdr-video-storage.md
- Files API, for `file_id` inputs and `storage_options` outputs

For every provider-derived fact, record the source URL, `verified_at`, the snapshot checksum, and whether it came from prose, the REST reference, the account UI, or a live response. Current sources override the baseline. Never rewrite historical pricing rows; publish a new effective-dated revision.

## Baseline facts (xAI docs fetched 2026-10-31) — reverify every line

### Endpoints (base `https://api.x.ai`)

| Method and path | Purpose | Response |
|---|---|---|
| `POST /v1/videos/generations` | text-to-video, image-to-video, reference-to-video | `{ "request_id": "..." }` |
| `POST /v1/videos/edits` | edit a source video (`grok-imagine-video`) | `{ "request_id": "..." }` |
| `POST /v1/videos/extensions` | extend a source video (`grok-imagine-video`) | `{ "request_id": "..." }` |
| `GET /v1/videos/{request_id}` | poll | `status`, `progress`, `model`, `video`, `usage`, `error` |

xAI has **no list and no delete endpoint**. GodarAI serves list from its own job records. Delete removes GodarAI's record and stops serving content; it cannot cancel or delete upstream.

### Generation request fields

- `model`, and `prompt`. The prompt is required for T2V and R2V and optional for I2V.
- `duration`: an integer from 1 to 15, default 8. xAI also accepts `seconds` "for OpenAI API compatibility", as a number or a string.
- `aspect_ratio`: `1:1`, `16:9`, `9:16`, `4:3`, `3:4`, `3:2`, `2:3`, `21:9`, `5:2`. It is ignored for I2V; the output matches the input image.
- `resolution`: `480p` (the default), `720p`, `1080p`. `grok-imagine-video` stops at 720p, and reference-to-video stops at 720p on every model.
- `image`: `{ url }` (a public URL or a base64 data URL; JPEG, PNG, or WebP) or `{ file_id }`. This is the first frame.
- `last_frame` and `keyframes[]` (at most 4, with `timestamp_s` strictly inside the clip): `grok-imagine-video-1.5` only.
- `reference_images[]`: up to 14 on `-1.5`, 7 on `grok-imagine-video`, none on `-lite`.
- `reference_audios[]` (at most 3; `voice_id` or `url`): `-1.5` only. Caller-supplied voice clips are "available to trusted partners on request".
- `storage_options` (`filename`, `expires_after` ≤ 2,592,000 s, `public_url`): stores the output in the xAI Files API.
- `output.upload_url`: xAI PUTs the MP4 to a caller-signed URL.
- `user`.

### Edit and extension request fields

- Edit: `model`, `prompt` (required), `video` (required; `{ url }` must be an `.mp4`, or `{ file_id }`), `storage_options`, `output`, `user`. The source must be at most 8.7 s. The output resolution matches the source, capped at 720p.
- Extension: the edit fields plus `duration`, the length of the new segment, 2–10 s with a default of 6.

### Poll response

- `status`: `pending`, `done`, or `failed` (from the docs table). The generation guide's example script also checks `expired`; this is in the conflict register.
- `progress`: 0–99 while pending, 100 when done, omitted when failed.
- `video`: `url` (temporary, lifetime not stated), `duration` (int seconds), `respect_moderation` (bool; when false, `url` is empty), `file_output { file_id, filename, expires_at, public_url, ... }`, and `storage_error`.
- `usage`: `cost_in_usd_ticks` (1 USD = 10,000,000,000 ticks), plus input and output token details.
- `error { code, message }`, where `code` is one of `invalid_argument`, `permission_denied`, `failed_precondition`, `service_unavailable`, or `internal_error`. Authentication, model-not-found, and rate-limit errors come back as HTTP errors, not here.

### Pricing

The Models page lists a flat rate per second: `-1.5` $0.080, `-1.5-lite` $0.020, `grok-imagine-video` $0.050. The overview and generation guide say that resolution also affects cost. Settle from `usage.cost_in_usd_ticks`. The catalog rate is only for the reservation estimate and the drift check.

## Public contract design

### Canonical OpenAI-style routes (every provider)

| GodarAI route | Behavior for xAI |
|---|---|
| `POST /v1/videos` | JSON or multipart. Maps `seconds` to `duration` and `size` to `aspect_ratio` + `resolution` using the profile's size table. Multipart `input_reference` becomes a base64 data URL in `image.url`, within the profile's byte limit. JSON `input_reference { image_url \| file_id }` becomes `image { url \| file_id }`, with `file_id` rewritten through `rewriteXAIFileReferences`. Returns an OpenAI `video` object with `status: "queued"`. |
| `GET /v1/videos/{video_id}` | Returns the OpenAI `video` object from GodarAI's record, refreshed from upstream when it is not terminal. |
| `GET /v1/videos` | Served entirely from the tenant's job records, with cursor pagination (`after`, `limit` ≤ 100, `order`). |
| `DELETE /v1/videos/{video_id}` | Returns `{ id, object: "video.deleted", deleted: true }`. Marks the record deleted and stops serving content. A pending job keeps settling, because xAI will still charge for it. |
| `GET /v1/videos/{video_id}/content` | Streams the MP4 from the upstream `video.url`, or from `file_output`. Only `variant=video` is supported for xAI. `thumbnail` and `spritesheet` return a stable `unsupported_variant` error, driven by the profile. |
| `POST /v1/videos/edits`, `POST /v1/videos/extensions` | OpenAI-style `video: { id }` takes a GodarAI public video ID, which is resolved to the completed job's upstream output. xAI-native `video: { url \| file_id }` is also accepted. |

`remix` is not offered: OpenAI deprecated it and xAI has no equivalent.

Provider-specific options (for xAI: `aspect_ratio`, `resolution`, `reference_images`, `last_frame`, `keyframes`, `reference_audios` with `voice_id` only, and `storage_options`) arrive as top-level fields, which is how the OpenAI SDK's `extra_body` sends them. They are accepted only if the model's runtime profile allow-lists them. Reject any unknown field. Never pass free-form fields through.

### xAI-native routes

- `POST /v1/videos/generations` takes the xAI request shape and returns `{ "request_id": "<GodarAI public id>" }`.
- `GET /v1/videos/{id}` collides with the canonical route. Use one route, and choose the response shape from the `public_api_shape` that was recorded on the job when it was created: native jobs get the xAI poll shape, canonical jobs get the OpenAI `video` object.
- In both shapes, every ID and URL the tenant sees comes from GodarAI:
  - The upstream `request_id` never appears.
  - The upstream `video.url` never appears either. Replace it with the GodarAI `/v1/videos/{id}/content` URL.
  - Any `file_id` goes through the existing xAI file mapping.

### Status mapping

| xAI | Canonical |
|---|---|
| `pending`, progress 0 | `queued` |
| `pending`, progress > 0 | `in_progress` |
| `done` with `respect_moderation: true` | `completed` |
| `done` with `respect_moderation: false` | `failed`, `error.code = "content_policy_violation"` |
| `failed` | `failed`, with the xAI `error.code` mapped to a stable GodarAI code |
| `expired`, or an unknown value | `failed`, `error.code = "upstream_expired"` or `"upstream_unknown_status"`, plus an alert |

### Fail-closed P0 choices

- Reject `output.upload_url`: it hands delivery to an arbitrary caller URL outside GodarAI's controls. Gate it for later.
- Reject `reference_audios[].url`: it is a trusted-partner feature. Allow `voice_id` from the profile's allow-list.
- `storage_options` is allowed only when the provider account and the existing xAI file-ownership mapping can record the resulting `file_id` for the tenant. Otherwise, reject it.

## Product scope

### Include — P0

- All of the canonical and xAI-native routes above, for all three models, within each model's capability matrix.
- Tenant-scoped ownership. Get, list, content, edit, extend, and delete across tenants all return the same 404 as an unknown ID.
- Reservation at create; settlement on completion; release when xAI does not charge (decided by probe for `failed` and for moderation-blocked jobs). Exactly once, idempotent, and crash-safe.
- Content streaming through the Silk Road egress boundary with the existing SSRF and host allow-list controls. Allow-list only the xAI video hosts that probes observe (the docs example uses `vidgen.x.ai`). Correct `Content-Type` and `Content-Length`. No buffering of the whole file in memory.
- Deployment policy, model allow-lists, budget limits, and rate limits apply as on every other endpoint.
- Usage logs and traces record model, seconds, resolution, aspect ratio, status transitions, provider cost ticks, and settled cost. Never log prompts, images, or videos beyond what the existing request-logging policy permits.

### Capability-gated, implementation-ready

- `output.upload_url`, caller-supplied `reference_audios[].url`, and ZDR video-storage behavior.
- Optional GodarAI-side persistence of finished videos past the upstream URL lifetime. This is a product decision; record it in the final report.

### Explicitly exclude

- Veo, Gemini Omni, Kling, Wan, and Seedance adapters, profiles, and pricing. The seam is in scope; those providers are not.
- xAI Batch API video jobs, and GodarChat/LibreChat UI integration.

## Extensible architecture (the seam future providers plug into)

1. **Public contract.** Put typed request and response contracts for both the canonical shape and the xAI-native shape in `packages/contracts` next to the images contracts.
2. **Runtime-profile video block.** Extend the model runtime profile (Go contract and TypeScript schema together) with a typed `video` block containing:
   - allowed durations (a range or a list), and the `size` → aspect-ratio/resolution table;
   - allowed resolutions and aspect ratios, and whether I2V ignores the aspect ratio;
   - reference modes: first frame, last frame, keyframes (with a maximum), reference images (with a maximum), and voice references (with the allowed sources);
   - support flags for edit and extend, with source and segment limits;
   - audio generation, content variants, maximum input-image bytes, and the allow-listed extension fields.

   All request validation reads from this block. Adding a provider later must be data, not code. Write all three xAI profiles from the capability matrix.
3. **Provider adapter interface in Silk Road.** One small interface with methods to create a job (generate, edit, or extend), poll it, and open its content stream. It returns canonical state, actual seconds, the provider cost if present, and an output locator. List and delete belong to the control plane, not the adapter, so providers without them (like xAI) need no stubs. The xAI implementation lives in or next to `xai.go` and reuses its retry, validation, and header-isolation code.
4. **Job lifecycle.** Use `ProviderNativeResourceMapping` with `resource_kind = "video"`. Store the public ID, the hidden upstream `request_id`, the reservation, the pricing revision, `public_api_shape`, the requested parameters, and the output locator. One reconciliation worker, modelled on `mistral_batch_worker.go`, polls non-terminal jobs with backoff (xAI's SDK polls every 1–5 s; use per-account budgets), settles exactly once, and expires stuck jobs after a configured ceiling. Client `GET`s may refresh a job opportunistically, but settlement only ever goes through the shared, idempotent path.
5. **Pricing.** Add a video cost path that is keyed by pricing unit, not by provider. Units are `per_second` (with optional resolution and audio tiers), `per_video` (with duration tiers), and `per_output_token`. xAI uses `per_second`:
   - **Estimate:** requested seconds × the highest catalog rate for the model. For extensions, use the new segment's seconds. For edits, use the source's seconds, capped at 8.7. The estimate covers the reservation only.
   - **Finalize:** `cost_in_usd_ticks` through `reconcileXAIProviderCost`. If the ticks are missing, settle at the catalog rate × the actual `video.duration`, mark the record `estimated_settlement`, and alert. Never leave the job unsettled.
6. **Content delivery.** One streaming download path for every provider, whether the upstream gives a signed URL (xAI) or raw bytes (future providers).
7. **Proof of extensibility.** Add a test-only fake adapter with a different native state machine, a bytes-based content download, and a `per_video` price. It must pass the same public contract suite as xAI with no change to routes, handler, worker, or UI. If it needs one, the seam is wrong; fix the seam.

Do not build anything else "for later". No adapter stubs, feature flags, or config for future providers.

## Required implementation work

### Phase 0 — evidence, conflict register, scope lock

Snapshot the sources and build the matrix. Start the conflict register with at least these items:

1. Price per second: flat on the Models page vs "resolution affects cost" in the guides.
2. Whether `expired` is a real poll status.
3. The lifetime of `video.url`.
4. Whether `failed` and moderation-blocked (`respect_moderation: false`) jobs are billed.
5. Whether extension `video.duration` is the new segment or the stitched total.
6. How edits are priced (input seconds, output seconds, or both).
7. Rate limits and GA status for each video model (not stated in the docs).
8. Whether the xAI SDK's base URL can point at GodarAI.

For each item, record the fail-closed choice and the probe that resolves it. Get the probe budget approved before any live call.

### Phase 1 — contracts, runtime profiles, catalog

- Add the `ProviderOperationXAIVideos*` operations, the typed contracts, and the runtime-profile `video` block, keeping Go and TypeScript in parity.
- Add the three `apps/model-discovery/providers/xai/grok-imagine-video*.yaml` files, the import, and the migrations, following the existing pattern.
- Add `CapabilityVideos` to the `xai` registry entry. Remove it from `openai`, along with the dead Sora operations.

### Phase 2 — Silk Road

- Add the adapter interface and the xAI video implementation.
- Validate requests against the profile.
- Convert multipart inputs to data URLs.
- Map statuses and errors.
- Stream content through the SSRF-guarded egress.
- Never let caller-supplied headers cross the boundary.

### Phase 3 — control-plane routes and handler

Register every route in `routes_proxy.go` behind a nil-guarded `VideosHandler`, as is done for images. Run the full pipeline: auth, deployment and FQN resolution, policy, profile validation, file-ID rewriting, reservation, tunnel, ID and URL rewriting, and logging.

### Phase 4 — lifecycle worker and settlement

Add the reconciliation worker, idempotent settlement, and release on every terminal path that is not charged. It must survive restarts. Add metrics for queue depth, time-to-complete, failure reasons by code, settlement drift, and estimated settlements.

### Phase 5 — billing and pricing

Add the pricing units and the effective-dated catalog rows for the three models. Convert through the credit rate card, apply plan or tier gating, and alert on drift between estimate and actual.

### Phase 6 — tenant-admin, catalog UI, docs

- Show the three models with a video badge and their capability matrix.
- Make the `provider-models-list.tsx` snippet provider-aware.
- Add a Playground panel only if an existing pattern can be reused; otherwise record it as UI debt.
- Update `scripts/postman/generate-godarai-postman.mjs`.
- Write Persian and English docs with OpenAI-SDK and xAI-REST examples for text-to-video, image-to-video, poll, download, edit, and extend.

### Phase 7 — readiness, canary, rollback

- Extend `cmd/xai-readiness` with video checks under a hard spend cap. Use 1–2 s `480p` clips on `grok-imagine-video-1.5-lite`, plus one edit and one extension on `grok-imagine-video`.
- The checks must cover create, poll, download, the cost-ticks presence check, the forced-failure release, and a moderation-block case if one can be triggered safely.
- Roll out on a canary tenant allow-list first.
- Rollback disables the routes and the catalog entries without stranding reservations. The worker keeps settling in-flight jobs.

## Required automated tests

- **Validation:** per-profile rejection of out-of-range `duration`/`seconds`, a disallowed resolution or `size`, a feature the model lacks (keyframes on lite, edits on `-1.5`), oversized `input_reference`, unknown extension fields, `output.upload_url`, and audio `url` references. All of these fail before any reservation is made.
- **Mapping:** OpenAI `size`/`seconds`/`input_reference` → xAI fields, the status table, and the error-code table.
- **Tenancy:** cross-tenant get, list, content, edit, extend, and delete return the same 404 as an unknown ID. No upstream `request_id`, `video.url`, or file ID appears in any response body or header.
- **Settlement:** a completed job is charged once from cost ticks; missing ticks lead to an estimated settlement plus an alert; failed and moderation-blocked jobs follow the probe-decided rule; duplicate worker ticks, and a crash between upstream completion and settlement, still settle exactly once; a pending job that is deleted still settles.
- **Streaming:** a large MP4 streams with bounded memory; an expired URL returns a stable error; non-allow-listed hosts are refused.
- **Response shape:** a job created natively returns the xAI poll shape from `GET /v1/videos/{id}`, and a job created through the canonical route returns the OpenAI `video` object.
- **Extensibility:** the fake adapter passes the same public suite with no handler, route, worker, or UI changes.
- **Profile schema:** the TypeScript runtime-profile schema matches the Go contract.

## Validation commands

From each changed Go module, run `go vet ./...`, `go test ./...`, and `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`. For changed TypeScript apps, run their `pnpm` lint, typecheck, test, and build scripts. Run `xai-readiness` video checks only against the approved account and budget.

## Definition of done

- A tenant can run text-to-video, image-to-video, edit, and extend end to end, with both the unmodified OpenAI Python/Node SDK and plain xAI REST, against GodarAI.
- Every in-scope route is registered, tenant-isolated, billed exactly, logged, and documented.
- Every conflict-register item is either resolved by a probe, or recorded as fail-closed with its reason.
- No code path can route a video request to OpenAI.
- The fake-adapter test proves that a second provider plugs in through the adapter, the profile, and the pricing rule only.
- All lint, type, test, build, and readiness commands pass, with nothing weakened, skipped, or `//nolint`ed without a specific reason.

## Required final report

List the files changed by area; the conflict register with the resolution of each item; live probe results and spend; routes enabled vs gated; pricing revision IDs; the persistence decision; remaining UI debt; and a one-page "adding provider N" guide showing exactly which files a Kling or Veo adapter would touch.
```

## Why this prompt is shaped this way

- xAI is the only verified, live video API among the providers requested. It also already fits the repo: `xai` is a provider with images, files, deferred polling, and cost-tick billing. The prompt points at those pieces so the implementer reuses them.
- The OpenAI Videos shape stays the canonical public API even though Sora is gone, because it is the most widely supported client shape and xAI itself accepts `seconds` for OpenAI compatibility. The xAI-native routes match house practice for xAI.
- Extensibility comes from profile data, one adapter interface, and a pricing-unit enum, and is proven by a fake provider in tests. Stubbing future providers now would be speculative code.
- List and delete stay in the control plane because xAI has neither. That keeps the adapter interface small and honest.

## Evidence used to create this prompt

Provider documentation fetched on 2026-10-31:

- xAI: `llms.txt`, REST reference `inference/videos.md`, `video/overview.md`, `video/generation.md`, `video/image-to-video.md`, `video/editing.md`, `video/extension.md`, `video/reference-to-video.md`, `models.md`.
- OpenAI: the deprecations page and the video-generation guide.
- Google: the Gemini API video page and the Omni page.

Repository inspection on 2026-10-31:

- control-plane: `routes_proxy.go`, `images_handler.go`, `xai_deferred_handler.go`, `xai_file_references.go`, `responses_operations_handler.go:28`, `providerregistry/registry.go`, `provider_pricing_strategy.go`, `pricing_strategy_images.go`, `mistral_batch_worker.go`, `cmd/xai-readiness`.
- silk-road: `xai.go` and `openai_operations.go:485-513`.
- contracts: `types.go:527-577` and `:728`.
- catalog and UI: `apps/model-discovery/providers/xai/grok-imagine-image-quality.yaml` and `provider-models-list.tsx:773-775`.

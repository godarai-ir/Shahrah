# GodarAI Google Gemini support evidence

- Verified (UTC): 2026-08-15
- Provider: Gemini Developer API (`google-gemini`)
- Policy: Google-owned Gemini Developer API documentation is authoritative. Vertex AI and third-party registries are not used to infer Developer API behavior.

## Current support summary

GodarAI has a fixed allowlist of ten Gemini Developer API models. Nine model IDs were returned by a live Google model-list request on 2026-08-02. `gemini-3.7-flash` was added from Google's GA model page and latest-model guide on 2026-08-15; it is contract verified but has not yet been exercised with the live credential. The gateway exposes both the native Gemini `v1beta` surface and a limited OpenAI-compatible surface. Requests for a model/endpoint combination that is not present in the model runtime profile fail closed before upstream dispatch.

The production data plane does **not** currently import or use `google.golang.org/genai`. It constructs native Gemini and OpenAI-compatible requests with Go `net/http`. The official SDK was installed only in a temporary external probe at version `v1.66.0`; the probe files and dependency graph were removed after verification.

The following status terms are used in this report:

- **Implemented**: a public GodarAI route, tunnel operation, model guard, and runtime-profile endpoint exist in the repository.
- **Upstream verified**: the operation was executed directly against Google with the supplied test credential. This does not replace an end-to-end test through a deployed GodarAI Control Plane and Silk Road.
- **Contract verified**: repository unit/handler tests passed, but no live upstream request was sent for that operation.
- **Not supported**: the combination is rejected by routing or by the model runtime profile.

## Supported model and endpoint matrix

This matrix describes the current GodarAI runtime contract, not every action returned by Google's model-list API.

| Model | OpenAI-compatible endpoints | Native Gemini endpoints | Resource endpoints | Live upstream evidence on 2026-08-02 |
| --- | --- | --- | --- | --- |
| `gemini-3.1-pro-preview` | `POST /v1/chat/completions` | `generateContent`, `streamGenerateContent`, `countTokens`, `batchGenerateContent` | models, Interactions, Files, cached contents, batches | model list and `countTokens` succeeded; generation returned `429 RESOURCE_EXHAUSTED` because this credential has zero Free Tier generation quota for the Pro model |
| `gemini-3.5-flash` | `POST /v1/chat/completions` | `generateContent`, `streamGenerateContent`, `countTokens`, `batchGenerateContent` | models, Interactions, Files, cached contents, batches | `countTokens` and SDK `generateContent` succeeded; generated text was `OK` |
| `gemini-3.5-flash-lite` | `POST /v1/chat/completions` | `generateContent`, `streamGenerateContent`, `countTokens`, `batchGenerateContent` | models, Interactions, Files, cached contents, batches | `countTokens`, SDK generation, SDK streaming, raw native REST, and raw OpenAI-compatible chat succeeded; outputs included `OK`, `STREAM_OK`, and `OPENAI_OK` |
| `gemini-3.6-flash` | `POST /v1/chat/completions` | `generateContent`, `streamGenerateContent`, `countTokens`, `batchGenerateContent` | models, Interactions, Files, cached contents, batches | `countTokens`, generation, and streaming returned successful upstream responses; the deliberately small output cap was partly consumed by thinking tokens |
| `gemini-3.7-flash` | `POST /v1/chat/completions` | `generateContent`, `streamGenerateContent`, `countTokens`, `batchGenerateContent` | models, Interactions, Files, cached contents, batches | contract verified from the GA model page and latest-model guide; live upstream request not yet executed |
| `gemini-3-pro-image` | `POST /v1/images/generations` | `generateContent`, `streamGenerateContent`, `batchGenerateContent` | models, Interactions, Files, batches | model presence and advertised Google actions verified; image generation was not executed |
| `gemini-3.1-flash-image` | `POST /v1/images/generations` | `generateContent`, `streamGenerateContent`, `batchGenerateContent` | models, Interactions, Files, batches | model presence and advertised Google actions verified; image generation was not executed |
| `gemini-3.1-flash-lite-image` | `POST /v1/images/generations` | `generateContent`, `streamGenerateContent`, `batchGenerateContent` | models, Files, batches | model presence and advertised Google actions verified; image generation was not executed; Interactions is intentionally not enabled |
| `gemini-3.1-flash-tts-preview` | none; `/v1/audio/speech` is not mapped to Gemini | `generateContent`, `streamGenerateContent`, `batchGenerateContent` | models, Interactions, batches | model presence and advertised Google actions verified; speech generation was not executed |
| `gemini-embedding-2` | `POST /v1/embeddings` | `embedContent`, `batchEmbedContents` | models, Files, batches | SDK `embedContent` succeeded and returned a 3,072-dimensional vector; compatible embeddings and batch embedding were not executed |

### Endpoint paths exposed by GodarAI

| Surface | Method and path | Allowed models | Verification state |
| --- | --- | --- | --- |
| OpenAI-compatible chat | `POST /v1/chat/completions` | five text models | implemented and contract verified; direct Google compatibility request verified with `gemini-3.5-flash-lite` |
| OpenAI-compatible embeddings | `POST /v1/embeddings` | `gemini-embedding-2` | implemented and contract verified; native SDK embedding verified, compatible endpoint not live-tested |
| OpenAI-compatible image generation | `POST /v1/images/generations` | three image models | implemented and contract verified; not live-tested |
| GodarAI model catalog | `GET /v1/models` | configured deployments and virtual models visible to the tenant | implemented; this is a local GodarAI catalog response, not a pass-through Google model-list request |
| Native model list | `GET /v1beta/models` | response filtered to the ten-model allowlist | implemented and contract verified; direct Google model list verified for the previous nine-model boundary |
| Native model lookup | `GET /v1beta/models/{model}` | all ten allowlisted models | implemented and contract verified |
| Native generation | `POST /v1beta/models/{model}:generateContent` | all models except `gemini-embedding-2` | implemented and contract verified; live text generation verified |
| Native streaming generation | `POST /v1beta/models/{model}:streamGenerateContent` | all models except `gemini-embedding-2` | implemented over SSE and contract verified; live text streaming verified |
| Native token count | `POST /v1beta/models/{model}:countTokens` | five text models only | implemented and contract verified; live token counting verified for the four models that preceded `gemini-3.7-flash` |
| Native embedding | `POST /v1beta/models/gemini-embedding-2:embedContent` | `gemini-embedding-2` | implemented and contract verified; live SDK embedding verified |
| Native synchronous batch embedding | `POST /v1beta/models/gemini-embedding-2:batchEmbedContents` | `gemini-embedding-2` | implemented and contract verified; not live-tested |
| Interactions create | `POST /v1beta/interactions` | text models, `gemini-3-pro-image`, `gemini-3.1-flash-image`, and TTS | implemented and contract verified; not live-tested |
| Interactions lifecycle | `GET` or `DELETE /v1beta/interactions/{id}`; `POST /v1beta/interactions/{id}:cancel` | same Interactions-capable models; account selected by `x-godarai-provider-account` when required | implemented and contract verified; not live-tested |
| File upload | `POST /upload/v1beta/files` | provider-account resource; uploaded files are usable as model input only where the model profile includes Files | implemented as a length-preserving raw upload and contract verified; not live-tested |
| Files lifecycle | `GET /v1beta/files`, `GET` or `DELETE /v1beta/files/{id}` | provider-account resource; text, image, and embedding profiles advertise Files input | implemented and contract verified; not live-tested |
| Explicit context cache | `POST` or `GET /v1beta/cachedContents`; `GET`, `PATCH`, or `DELETE /v1beta/cachedContents/{id}` | five text models | implemented and contract verified; not live-tested |
| Batch submission | `POST /v1beta/models/{model}:batchGenerateContent` | generative text, image, and TTS models | implemented and contract verified; not live-tested |
| Batch lifecycle | `GET /v1beta/batches`, `GET /v1beta/batches/{id}`, `POST /v1beta/batches/{id}:cancel` | operations belonging to the selected Gemini provider account | implemented and contract verified; not live-tested |

### Explicitly unsupported combinations

- `gemini-embedding-2` cannot use chat, content generation, streaming generation, image generation, TTS, Interactions, or explicit context caching.
- Image and TTS models cannot use `/v1/chat/completions`; image models use `/v1/images/generations`, while TTS uses native `generateContent` or `streamGenerateContent`.
- Only the five text models can use `countTokens` and explicit context caching in the current GodarAI runtime profiles, even if Google advertises additional actions for another model.
- `gemini-3.1-flash-lite-image` cannot use Interactions.
- OpenAI-compatible `/v1/audio/speech`, Responses API, legacy Completions, image edits/variations, Realtime/Live WebSocket, Imagen, and Veo are not part of the Gemini adapter surface.
- Model aliases, `*-latest`, Gemma, agents, preview predecessors, and models outside the ten-ID allowlist are not routed.

## Live verification record

The temporary probe produced these non-secret results:

- Google SDK model list: 58 models returned; all nine GodarAI IDs configured at the time were present. `gemini-3.7-flash` was released later and is not covered by this probe.
- `gemini-3.5-flash-lite`: six input tokens counted; normal generation, streaming generation, raw REST token counting, and OpenAI-compatible chat succeeded.
- `gemini-3.5-flash`: token counting and generation succeeded.
- `gemini-3.6-flash`: token counting, generation, and streaming requests succeeded.
- `gemini-3.7-flash`: no live probe has been run; official specification and repository contracts were verified on 2026-08-15.
- `gemini-3.1-pro-preview`: token counting succeeded; generation could not be exercised with the Free Tier credential because Google reported quota limit `0` for the Pro model.
- `gemini-embedding-2`: embedding succeeded with 3,072 output dimensions.

The following repository suites also passed:

- `services/silk-road`: `go test ./pkg/server ./cmd/server`
- `services/control-plane`: proxy tests and the then-existing direct-provider
  adapter suite (historical verification; the unused adapter suite was later
  removed, and current provider identity coverage lives in
  `./internal/providerregistry`)
- `packages/contracts`: `go test ./platform`

Media generation, Interactions, Files, caches, batch submission/lifecycle, compatible embeddings, and an end-to-end request through a running GodarAI deployment remain unverified by this live credential run. Their status is contract verified, not live verified.

## Curated model boundary

The active catalog and runtime allowlist contains exactly these model IDs:

1. `gemini-3.1-pro-preview`
2. `gemini-3.5-flash`
3. `gemini-3.5-flash-lite`
4. `gemini-3.6-flash`
5. `gemini-3.7-flash`
6. `gemini-3-pro-image`
7. `gemini-3.1-flash-image`
8. `gemini-3.1-flash-lite-image`
9. `gemini-3.1-flash-tts-preview`
10. `gemini-embedding-2`

Aliases, agents, Gemma, Imagen, Veo, Live models, preview predecessors, `*-latest`, and `gemini-3.1-pro-preview-customtools` are outside the boundary and must not be advertised or routed.

## Model evidence matrix

Legend: **G** = documented by Google, **N** = native Gemini API surface, **O** = Google OpenAI-compatible surface, **—** = unsupported, **D** = account-specific/dynamic or not publicly specified.

| Model | Lifecycle | Modalities (input → output) | Limits | Applicable surfaces | Thinking | Google capabilities | Consumption |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `gemini-3.1-pro-preview` | Preview; released 2026-02-19; no shutdown announced | text, image, video, audio, PDF → text | 1,048,576 input; 65,536 output | N: Interactions, generate/stream content, count tokens, models, files/cache where accepted, Batch. O: chat completions and documented Batch compatibility | low, medium, high; default high in the current guide | caching, code execution, function calling, parallel calls, structured output, Search, Maps, URL context; File Search is documented as AI Studio only | Standard, Batch, Flex, Priority; no Free pricing |
| `gemini-3.5-flash` | GA; released 2026-05-19; no shutdown announced | text, image, video, audio, PDF → text | 1,048,576 input; 65,536 output | N: Interactions, generate/stream content, count tokens, models, files/cache, Batch. O: chat completions and documented Batch compatibility | low, medium, high; default medium | caching, code execution, Computer Use (preview), File Search, function calling, Search, Maps, structured output, URL context | Free availability plus Standard, Batch, Flex, Priority |
| `gemini-3.5-flash-lite` | GA; released 2026-07-21; no shutdown announced | text, image, video, audio, PDF → text | 1,048,576 input; 65,536 output | same text-model native/compatible surfaces as above | minimal and high are the documented tuning choices; default minimal | full built-in tool suite including Computer Use; function calling and structured output | Free availability plus Standard, Batch, Flex, Priority |
| `gemini-3.6-flash` | GA; released 2026-07-21; no shutdown announced | text, image, video, audio, PDF → text | 1,048,576 input; 65,536 output | same text-model native/compatible surfaces as above | medium and high; default medium | full built-in tool suite including Computer Use; function calling and structured output | Free availability plus Standard, Batch, Flex, Priority |
| `gemini-3.7-flash` | GA; released 2026-08-13; no shutdown announced | text, image, video, audio, PDF → text | 1,048,576 input; 65,536 output | same text-model native/compatible surfaces as above | low, medium, high; default medium; `minimal` is rejected | same built-in tool suite as 3.6 Flash, including Computer Use; function calling and structured output | Free availability plus Standard, Batch, Flex, Priority |
| `gemini-3-pro-image` | GA; released 2026-05-28; no shutdown announced | text, image → text, image | 65,536 input; 32,768 output | N: Interactions, generate content/image editing, models, Batch. O: image generation only where the compatibility guide accepts the stable ID | thinking is supported | image generation/editing and Search grounding; no caching, code execution, function calling, Maps, structured output, URL context, or Live | Standard, Batch, Flex, Priority |
| `gemini-3.1-flash-image` | GA; released 2026-05-28; no shutdown announced | text, image/PDF and model-specific video context → text, image | 131,072 input; 32,768 output | N: Interactions, generate content/image editing, models, Batch. O: image generation only where documented for the stable ID | supported; Search works with thinking on or off | image generation/editing, Web and Image Search grounding; no caching, code execution, function calling, Maps, structured output, URL context, or Live | Standard and Batch; no Flex/Priority |
| `gemini-3.1-flash-lite-image` | GA; latest update 2026-06; no shutdown announced | text, image → text, image | 65,536 input; 4,096 output | N: generate content/image editing and Batch; Interactions support is not listed in the current Interactions supported-model table and is therefore not advertised without live official confirmation | minimal and high | function calling, image generation/editing; no Search/Maps grounding, caching, code execution, File Search, structured output, URL context, or Live | Standard and Batch |
| `gemini-3.1-flash-tts-preview` | Preview; released 2026-04-13; no shutdown announced | text → audio | 8,192 input; 16,384 output | N: Interactions (including SSE) and legacy generateContent streaming; Batch. O: no documented `/audio/speech` compatibility contract | unsupported | steerable speech, expressive tags, single/multi-speaker options described in the speech guide; no tools/caching/thinking/Live | Free availability plus Standard and Batch |
| `gemini-embedding-2` | GA; released 2026-04-22; no shutdown announced | text, image, video, audio, PDF → embedding | 8,192 input; flexible dimension 128–3,072 (recommended 768/1,536/3,072) | N: `embedContent`, batch embedding and supporting files. O: embeddings where documented | unsupported | task-specific multimodal embeddings; no generative tools | Free availability plus Standard and Batch |

## Parameter decisions

- `gemini-3.7-flash`, `gemini-3.6-flash`, and `gemini-3.5-flash-lite` deprecate and ignore `temperature`, `top_p`, and `top_k`; the current latest-model guide says to remove them. They also reject a request whose last non-empty turn is a model/assistant prefill.
- Thinking is represented as `thinking_level`, with model-specific values. OpenAI-compatible `reasoning_effort` is only mapped where Google documents the mapping.
- Image-only response controls (resolution, aspect ratio, MIME type, reference images) are not modeled as text sampling controls.
- Embedding `output_dimensionality` is independent of generative max-output tokens.

## Image evidence and conflict resolution

The current image-generation guide is more specific than the generic model table for video-to-image generation. It documents video context exclusively for `gemini-3.1-flash-image`; that exception is retained even though the compact model table lists text and image/PDF inputs.

The prior local YAML values for `gemini-3.1-flash-image` (1,048,576 input and 65,536 output) conflict with the model page. The model page's 131,072 input and 32,768 output values win.

- `gemini-3.1-flash-image`: 0.5K, 1K, 2K, 4K; Web and Image Search grounding; up to 14 total reference images subject to documented object/character limits.
- `gemini-3-pro-image`: 1K, 2K, 4K; Search grounding; up to 14 references subject to documented limits.
- `gemini-3.1-flash-lite-image`: 1K only; 14 documented aspect ratios/reference objects; no Search grounding.
- Generated images carry Google's documented SynthID watermark; Lite Image also documents C2PA.

The image guide currently enumerates ten ratios (`1:1`, `3:2`, `2:3`, `3:4`, `4:3`, `4:5`, `5:4`, `9:16`, `16:9`, `21:9`) while the Lite Image model page says its discrete set contains fourteen. Per the conflict policy, GodarAI validates the ten enumerated values and does not invent the four unnamed values. Reference limits are: Lite, 14 object images; Flash, 10 object plus 4 character images; Pro, 6 object plus 5 character images within 14 total.

## TTS and embedding details

- TTS output examples define raw mono PCM at 24 kHz and 16-bit sample width; binary/base64 data is forwarded unchanged. Single-speaker and up-to-two-speaker configurations are supported. The current prebuilt voice set is `Zephyr`, `Puck`, `Charon`, `Kore`, `Fenrir`, `Leda`, `Orus`, `Aoede`, `Callirrhoe`, `Autonoe`, `Enceladus`, `Iapetus`, `Umbriel`, `Algieba`, `Despina`, `Erinome`, `Algenib`, `Rasalgethi`, `Laomedeia`, `Achernar`, `Alnilam`, `Schedar`, `Gacrux`, `Pulcherrima`, `Achird`, `Zubenelgenubi`, `Vindemiatrix`, `Sadachbia`, `Sadaltager`, and `Sulafat`. Style, accent, pace and tone are prompt-controlled. Version 3.1 supports streamed Interactions `step.delta` audio events.
- Embedding 2 accepts inline or Files-backed text, image, audio, video and PDF content. Output dimensions are 128–3,072; 768, 1,536 and 3,072 are recommended. Google automatically normalizes vectors truncated below 3,072 dimensions. Task types are represented only where the current embedding/API reference documents them; the importer does not transfer `gemini-embedding-001`-only task-type claims to Embedding 2.

## Pricing evidence (USD per one million tokens unless stated)

| Model/tier | Input | Output | Cache read | Cache storage | Other |
| --- | ---: | ---: | ---: | ---: | --- |
| 3.7 and 3.6 Flash promotional Standard / Batch / Flex / Priority, through 2026-12-31 | 0.75 / 0.375 / 0.375 / 1.35 | 3.75 / 1.875 / 1.875 / 6.75 | 0.075 / 0.0375 / 0.0375 / 0.135 | 0.50 per 1M token-hour | Free availability on Standard/Priority as documented |
| 3.7 and 3.6 Flash Standard / Batch / Flex / Priority, from 2027-01-01 | 1.50 / 0.75 / 0.75 / 2.70 | 7.50 / 3.75 / 3.75 / 13.50 | 0.15 / 0.075 / 0.075 / 0.27 | 1.00 per 1M token-hour | scheduled standard pricing |
| 3.5 Flash Standard / Batch / Flex / Priority | 1.50 / 0.75 / 0.75 / 2.70 | 9.00 / 4.50 / 4.50 / 16.20 | 0.15 / 0.075 / 0.08 / 0.27 | 1.00 per 1M token-hour | Free availability on Standard/Priority as documented |
| 3.5 Flash-Lite Standard / Batch / Flex / Priority | 0.30 / 0.15 / 0.15 / 0.54 | 2.50 / 1.25 / 1.25 / 4.50 | 0.03 / 0.02 / 0.02 / 0.05 | 1.00 per 1M token-hour | Same token rate for text/image/video/audio input |
| 3.1 Pro Preview Standard | 2.00 ≤200k; 4.00 >200k | 12.00 ≤200k; 18.00 >200k | 0.20 ≤200k; 0.40 >200k | 4.50 per 1M token-hour | no Free tier |
| 3.1 Pro Batch/Flex | 1.00 ≤200k; 2.00 >200k | 6.00 ≤200k; 9.00 >200k | 0.20 ≤200k; 0.40 >200k | 4.50 per 1M token-hour | exact boundary is first prompt token 200,001 |
| 3.1 Pro Priority | 3.60 ≤200k; 7.20 >200k | 21.60 ≤200k; 32.40 >200k | 0.36 ≤200k; 0.72 >200k | 8.10 per 1M token-hour | exact boundary is first prompt token 200,001 |
| 3.1 Flash Image Standard / Batch | 0.50 / 0.25 | text+thinking 3.00 / 1.50; image tokens 60 / 30 | — | — | standard images: $0.045 0.5K, $0.067 1K, $0.101 2K, $0.151 4K; Batch equivalents documented separately |
| 3.1 Flash Lite Image Standard / Batch | 0.25 / 0.125 | text+thinking 1.50 / 0.75; image tokens 30 / 15 | — | — | 1K image $0.0336 / $0.0168 |
| 3 Pro Image Standard | text/image token 2.00 | text+thinking 12.00; image tokens 120 | — | — | input image $0.0011; output $0.134 for 1K/2K, $0.24 for 4K |
| 3 Pro Image Batch/Flex | text 1.00 | text+thinking 6.00 | — | — | input image $0.0006; output $0.067 for 1K/2K, $0.12 for 4K; Priority must follow the current pricing page if retained |
| 3.1 Flash TTS Standard / Batch | text 1.00 / 0.50 | audio 20.00 / 10.00 | — | — | audio is 25 tokens/second |
| Embedding 2 Standard / Batch | text 0.20/0.10; image 0.45/0.225; audio 6.50/3.25; video 12.00/6.00 | — | — | — | per-image/per-second/per-frame equivalents are also published |

Search and Maps are billed per actual upstream query after the shared 5,000 monthly allowance: $14 per 1,000 queries. Retrieved grounding context is not separately charged as input tokens where the pricing note says so.

## Rate-limit evidence

- Quotas apply per Google Cloud project, not per API key.
- Primary dimensions are RPM, input TPM and RPD; model-specific IPM/TPD can also apply.
- Active limits are dynamic by project usage tier/account status and are visible in AI Studio. Static RPM/TPM/RPD values must not be fabricated.
- Spend limits use a rolling ten-minute window: Tier 1 $10, Tier 2 $200, Tier 3 $200; exhaustion returns `429 RESOURCE_EXHAUSTED`.
- Priority has separate capacity with a documented default of 0.3× the corresponding Standard limit while also counting toward interactive traffic.
- Batch: 100 concurrent jobs, 2 GB input file, 20 GB file storage. Published Tier-1 queue limits for the in-scope text models are Pro 5M, 3.5 Flash-Lite 10M, 3.6 Flash 3M and 3.5 Flash 3M tokens. The page currently names older preview labels for some image/embedding rows; those values are not transferred to stable IDs without explicit official confirmation.

## Authentication evidence

- Native API-key and authorization-key requests use `x-goog-api-key`.
- Google OpenAI compatibility documents `Authorization: Bearer <GEMINI_API_KEY>`.
- OAuth uses `Authorization: Bearer <access-token>` and requires lifecycle-aware acquisition/refresh; a short-lived access token is not a permanent secret.
- Authorization keys are bound to a service account but retain API-key wire semantics.
- Google has announced migration away from unrestricted Standard keys; the connection UI/docs must not promise that an unrestricted legacy key remains viable.

## Implemented GodarAI surface

GodarAI keeps the native Gemini and OpenAI-compatible surfaces explicit:

- **G/N** means Google native support routed through fixed `ProviderOperation` values for Interactions lifecycle, generate/stream content, count tokens, embeddings, models, Files, caches and Batch resources.
- **G/O** means the documented compatibility operations (`chat/completions`, `embeddings`, and image generation) are routed through Google's `/v1beta/openai` surface. GodarAI model discovery remains local as described below.
- The Control Plane and Silk Road share the exact Go allowlist; the TypeScript importer has an independently tested matching list. Unsupported combinations fail closed before dispatch.
- Every curated Gemini YAML carries an explicit immutable `runtimeProfile` version 2. Import preserves prior profile versions and inserts the declared version; endpoint, transport, modality, tool, tier, quota and limit parity is enforced by tests.
- Native API keys use `x-goog-api-key`; compatibility API keys use Bearer. An already-refreshed OAuth access token is carried only in the internal mTLS request and is never logged or persisted as a durable credential.
- The native `/v1beta/models` response is filtered to the allowlist, and native model lookup fails closed. The public OpenAI-compatible `GET /v1/models` endpoint is a tenant-scoped GodarAI catalog assembled from configured deployments and virtual models; it does not pass through Google's upstream list. The dedicated importer owns catalog reconciliation; `--clear-missing` removes database model rows that no longer have a YAML profile.
- Gemini file uploads use the length-prefixed raw tunnel rather than JSON wrapping. Native generate, streaming generate, Interactions create, and embedding operations reserve budget and finalize against returned `usageMetadata`, including cached, thought, image/audio and Search-query usage when Google returns it.
- An account-specific or undocumented limit remains discovery metadata, not a fabricated GodarAI enforcement ceiling.
- The data plane currently uses direct REST through `net/http`, not `google.golang.org/genai`. This is a supported partner-integration architecture, but Google recommends the official SDK for enterprise gateways and requires partner traffic to identify the client with `x-goog-api-client`.
- GodarAI does not currently add `x-goog-api-client` to Gemini native or compatibility requests. Adding a stable value such as `shahrah/<version>` to both request paths is tracked as integration debt; this absence does not invalidate the successful authentication and inference probes above.

## Operational and rollback notes

- Apply migration `202607310001_google_gemini_connection_schemas`, then run `pnpm --filter model-discovery import:provider:google-gemini -- --clear-missing`. The import is deterministic and the post-import invariant requires exactly ten models.
- Rollback requires restoring removed model profiles and importing them again; connection-schema changes should be superseded by a new forward migration rather than rewriting an applied migration.
- The `google_service_account` schema is intentionally `runtimeReady=false`: Google's OAuth guide supports access tokens, but a production token-source/refresh contract must be configured before this durable mode is advertised. The implemented runtime accepts only an ephemeral already-refreshed token over mTLS. This is an explicit credential-lifecycle limitation, not an API-key limitation.

## Primary official sources

- Model pages: <https://ai.google.dev/gemini-api/docs/models>
- API reference: <https://ai.google.dev/api>
- Interactions: <https://ai.google.dev/gemini-api/docs/interactions-overview>
- Latest model behavior: <https://ai.google.dev/gemini-api/docs/latest-model>
- OpenAI compatibility: <https://ai.google.dev/gemini-api/docs/openai>
- Image generation: <https://ai.google.dev/gemini-api/docs/image-generation>
- Speech generation: <https://ai.google.dev/gemini-api/docs/speech-generation>
- Embeddings: <https://ai.google.dev/gemini-api/docs/embeddings>
- Pricing: <https://ai.google.dev/gemini-api/docs/pricing>
- Rate limits: <https://ai.google.dev/gemini-api/docs/rate-limits>
- Authentication: <https://ai.google.dev/gemini-api/docs/api-key> and <https://ai.google.dev/gemini-api/docs/oauth>
- Official SDKs: <https://ai.google.dev/gemini-api/docs/libraries> and <https://pkg.go.dev/google.golang.org/genai>
- Partner and gateway integration paths: <https://ai.google.dev/gemini-api/docs/partner-integration>
- Lifecycle: <https://ai.google.dev/gemini-api/docs/deprecations> and <https://ai.google.dev/gemini-api/docs/changelog>

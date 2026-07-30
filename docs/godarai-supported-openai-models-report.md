# Cumulative OpenAI model support in GodarAI

**Review date:** 2026-07-30
**Provider:** OpenAI only
**Audited models:** 21
**Private pricing:** Out of scope
**Responses REST baseline:** `openai/openai-openapi@db14b6e1712aaf5265cf5a6871adff7a9c61d31c`

## Executive result

This report is cumulative. It includes the previously reviewed GPT-5.4,
GPT-5.5, and GPT-5.6 families together with the later GPT-4.1, embedding,
moderation, image, transcription, and speech models.

GodarAI publicly exposes all seven OpenAI Responses REST operations. Request,
response, error, query, header, and SSE contracts are relayed without narrowing
the provider schema. All 21 models are present in the OpenAI catalog and have
endpoint, modality, feature, tool, public-pricing, and reference-rate-limit
metadata appropriate to their model type.

The implementation is route-, schema-, catalog-, billing-, and
automated-contract-ready for the REST capabilities marked ready below. It is
not accurate to call the complete OpenAI surface production-ready yet:
Responses WebSocket mode, Realtime transcription, fine-tuning job management,
and several newer GPT-5.6 Responses features still have gaps. No credentialed
OpenAI staging canary was run.

## Cumulative model and endpoint inventory

| Model | Public GodarAI endpoints | Context / max output | Modalities | Status |
|---|---|---:|---|---|
| [`gpt-5.6-sol`](https://developers.openai.com/api/docs/models/gpt-5.6-sol) | Chat Completions, Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.6-terra`](https://developers.openai.com/api/docs/models/gpt-5.6-terra) | Chat Completions, Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.6-luna`](https://developers.openai.com/api/docs/models/gpt-5.6-luna) | Chat Completions, Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.5`](https://developers.openai.com/api/docs/models/gpt-5.5) | Chat Completions, Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.5-pro`](https://developers.openai.com/api/docs/models/gpt-5.5-pro) | Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready; no streaming |
| [`gpt-5.4`](https://developers.openai.com/api/docs/models/gpt-5.4) | Chat Completions, Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.4-pro`](https://developers.openai.com/api/docs/models/gpt-5.4-pro) | Responses REST, Batch | 1.05M / 128K | Text/image → text | REST ready |
| [`gpt-5.4-mini`](https://developers.openai.com/api/docs/models/gpt-5.4-mini) | Chat Completions, Responses REST, Batch | 400K / 128K | Text/image → text | REST ready |
| [`gpt-5.4-nano`](https://developers.openai.com/api/docs/models/gpt-5.4-nano) | Chat Completions, Responses REST, Batch | 400K / 128K | Text/image → text | REST ready |
| [`gpt-4.1`](https://developers.openai.com/api/docs/models/gpt-4.1) | Chat Completions, Responses REST, Batch | 1,047,576 / 32,768 | Text/image/PDF file → text | Inference REST ready |
| [`gpt-4.1-mini`](https://developers.openai.com/api/docs/models/gpt-4.1-mini) | Chat Completions, Responses REST, Batch | 1,047,576 / 32,768 | Text/image/PDF file → text | Inference REST ready |
| [`gpt-4o-mini`](https://developers.openai.com/api/docs/models/gpt-4o-mini) | Chat Completions, Responses REST, Batch | 128K / 16,384 | Text/image → text | Inference REST ready |
| [`text-embedding-3-small`](https://developers.openai.com/api/docs/models/text-embedding-3-small) | Embeddings, Batch | 8,192 input | Text → embedding | Ready |
| [`text-embedding-3-large`](https://developers.openai.com/api/docs/models/text-embedding-3-large) | Embeddings, Batch | 8,192 input | Text → embedding | Ready |
| [`omni-moderation-latest`](https://developers.openai.com/api/docs/models/omni-moderation-latest) | Moderations | Model-managed | Text/image → moderation result | Ready |
| [`gpt-image-2`](https://developers.openai.com/api/docs/models/gpt-image-2) | Image generations, image edits, Batch | Model-managed | Text/image → image | Ready |
| [`gpt-transcribe`](https://developers.openai.com/api/docs/models/gpt-transcribe) | Audio transcriptions, streamed file transcription | 25 MB upload | Audio/text hints → text | REST/SSE ready; Realtime unavailable |
| [`gpt-4o-mini-transcribe`](https://developers.openai.com/api/docs/models/gpt-4o-mini-transcribe) | Audio transcriptions, streamed file transcription | 25 MB upload | Audio/text hints → text | REST/SSE ready |
| [`gpt-4o-transcribe-diarize`](https://developers.openai.com/api/docs/models/gpt-4o-transcribe-diarize) | Audio transcriptions, streamed file transcription | 25 MB upload | Audio/speaker references → diarized text | REST/SSE ready |
| [`tts-1`](https://developers.openai.com/api/docs/models/tts-1) | Audio speech | 4,096 characters | Text → audio | Ready |
| [`tts-1-hd`](https://developers.openai.com/api/docs/models/tts-1-hd) | Audio speech | 4,096 characters | Text → audio | Ready |

Audio and video are not input or output modalities for the audited general
text models. `gpt-image-2` does not support image variations; OpenAI documents
`/v1/images/variations` for DALL·E 2 only, so GodarAI rejects that model-route
combination instead of forwarding an invalid request.

## Responses API contract coverage

### Public operations

All seven Responses REST operations are public:

| Method | Public route | Status |
|---|---|---|
| `POST` | `/v1/responses` | Implemented |
| `POST` | `/v1/responses/compact` | Implemented |
| `POST` | `/v1/responses/input_tokens` | Implemented |
| `GET` | `/v1/responses/{response_id}` | Implemented |
| `DELETE` | `/v1/responses/{response_id}` | Implemented |
| `POST` | `/v1/responses/{response_id}/cancel` | Implemented |
| `GET` | `/v1/responses/{response_id}/input_items` | Implemented |

The generated readiness audit reports `official_operation_count: 7`,
`implemented_operation_count: 7`, no missing or extra operations,
`route_ready: true`, and `contract_ready: true`.

### Request, response, streaming, and lifecycle contracts

GodarAI preserves:

- unknown and future request and response fields;
- OpenAI error status, headers, and JSON bodies;
- query parameters and pagination fields;
- SSE event names, bodies, ordering, and terminal events;
- `X-Client-Request-Id`, `X-Request-Id`, `Retry-After`, and provider
  rate-limit headers;
- background-response create, retrieve, cancel, and delete behavior;
- input-item listing and token-counting payloads;
- response compaction payloads;
- tool definitions and model-specific tool fields.

Stored response IDs are bound to the tenant and provider account that created
them. A continuation may omit `model`; GodarAI derives routing and billing from
the stored `previous_response_id` mapping without injecting a model into the
provider request.

OpenAI Batch support includes `POST/GET /v1/batches`,
`GET /v1/batches/{batch_id}`, and
`POST /v1/batches/{batch_id}/cancel`. Supporting Files operations are public
at `POST/GET /v1/files`, `GET/DELETE /v1/files/{file_id}`, and
`GET /v1/files/{file_id}/content`. Uploads and downloads use the raw streaming
tunnel. File and Batch IDs retain tenant and provider-account affinity.

## General-model capabilities and tools

All audited GPT-5 and GPT-4.1-family models have model-specific streaming,
reasoning, structured-output, function-calling, prompt-caching, and tool
capabilities in the runtime profile. GodarAI does not infer capabilities from
model size.

`gpt-4.1`, `gpt-4.1-mini`, and `gpt-4o-mini` advertise function calling,
parallel function calling, tool choice, prompt caching, JSON/structured output,
predicted outputs, and fine-tuning capability. Their Responses tool set is
function calling, web search, file search, image generation, code interpreter,
and MCP.

For GPT-5 models, `FC` means function calling. The remaining abbreviations are
`WS` web search, `FS` file search, `TS` tool search, `IG` image generation,
`CI` code interpreter, `HS` hosted shell, `AP` apply patch, `SK` skills,
`CU` computer use, and `MCP`.

| Model | FC | WS | FS | TS | IG | CI | HS | AP | SK | CU | MCP |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| GPT-5.6 Sol/Terra/Luna | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.5 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.5 Pro | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | — | — | — | ✓ |
| GPT-5.4 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.4 Pro | ✓ | ✓ | ✓ | ✓ | ✓ | — | — | ✓ | — | ✓ | ✓ |
| GPT-5.4 mini | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GPT-5.4 nano | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ |

The relay is schema-forward-compatible, but two capabilities newly documented
for GPT-5.6—Programmatic Tool Calling and the beta multi-agent tool—are not yet
explicit capability entries in the GodarAI model catalog or dedicated
contract tests. They are therefore not marked production-ready in this audit.

## Specialized endpoint contracts

### Embeddings and moderation

`/v1/embeddings` supports string, string-array, token-array, and token-matrix
input together with `dimensions`, `encoding_format`, and `user`. Request and
response payloads are relayed losslessly after model routing.

`/v1/moderations` supports text, image URL, and mixed text/image input for
`omni-moderation-latest`. The model is free in OpenAI's public catalog;
GodarAI still records usage and provider request metadata for observability.

### Images

`/v1/images/generations` and `/v1/images/edits` preserve JSON or multipart
fields and binary inputs. Billing reconciles text input, cached text input,
image input, cached image input, and image output as separate dimensions.
Standard and Batch rates are selected independently.

### Speech to text

`/v1/audio/transcriptions` preserves multipart data and current
model-specific fields, and supports streamed file-transcription events.

- `gpt-transcribe` accepts `languages[]` and `keywords[]`; `language` and
  `languages[]` cannot be combined, and unsafe keyword delimiters are rejected.
- `gpt-4o-transcribe-diarize` rejects unsupported `prompt`, accepts
  `chunking_strategy`, and requires equal known-speaker name/reference arrays
  with no more than four speakers.
- Final `usage.seconds` is converted to billable audio minutes, including the
  final event of a streamed transcription.

### Text to speech

`/v1/audio/speech` preserves the OpenAI request contract and provider content
type. Binary audio streams through the raw tunnel without buffering the full
response. `tts-1` and `tts-1-hd` are billed by exact Unicode input-character
count.

## Public pricing represented in GodarAI

Prices are OpenAI public list prices in USD. Token rates are per one million
tokens. Private tenant or provider-account price overrides are excluded.

### GPT-5 families

| Model | Standard input / cached / output | Batch and Flex input / cached / output | Fast input / cached / output |
|---|---:|---:|---:|
| `gpt-5.6-sol` | 5 / 0.5 / 30 | 2.5 / 0.25 / 15 | 10 / 1 / 60 |
| `gpt-5.6-terra` | 2 / 0.2 / 12 | 1 / 0.1 / 6 | 4 / 0.4 / 24 |
| `gpt-5.6-luna` | 0.2 / 0.02 / 1.2 | 0.1 / 0.01 / 0.6 | 0.4 / 0.04 / 2.4 |
| `gpt-5.5` | 5 / 0.5 / 30 | 2.5 / 0.25 / 15 | 12.5 / 1.25 / 75 |
| `gpt-5.5-pro` | 30 / — / 180 | 15 / — / 90 | — |
| `gpt-5.4` | 2.5 / 0.25 / 15 | 1.25 / 0.125 / 7.5 | 5 / 0.5 / 30 |
| `gpt-5.4-pro` | 30 / — / 180 | 15 / — / 90 | — |
| `gpt-5.4-mini` | 0.75 / 0.075 / 4.5 | 0.375 / 0.0375 / 2.25 | 1.5 / 0.15 / 9 |
| `gpt-5.4-nano` | 0.2 / 0.02 / 1.25 | 0.1 / 0.01 / 0.625 | — |

For 1.05M-context models, requests above 272K input tokens use the published
long-context rates for the full request. GodarAI has distinct Standard, Batch,
and Flex long-context dimensions. Fast/Priority is rejected for unsupported
long-context combinations. `service_tier: "fast"` and `"priority"` use the
Fast dimensions.

GPT-5.6 cache-write tokens are preserved from
`usage.input_tokens_details.cache_write_tokens` and billed separately. Pro
models fall back to ordinary input pricing for cached tokens because OpenAI
does not publish a cached-input discount.

The central [OpenAI pricing table](https://developers.openai.com/api/docs/pricing)
currently lists the Terra and Luna values used above and in GodarAI. Their
individual model cards currently display different Standard prices
(`gpt-5.6-terra`: 2.5 / 0.25 / 15; `gpt-5.6-luna`: 1 / 0.1 / 6). This is an
upstream documentation conflict. GodarAI follows the dedicated central pricing
table pending clarification.

### GPT-4.1, GPT-4o mini, and embeddings

| Model | Standard input / cached / output | Batch input / output | Fast input / cached / output |
|---|---:|---:|---:|
| `gpt-4.1` | 2 / 0.5 / 8 | 1 / 4 | 3.5 / 0.875 / 14 |
| `gpt-4.1-mini` | 0.4 / 0.1 / 1.6 | 0.2 / 0.8 | 0.7 / 0.175 / 2.8 |
| `gpt-4o-mini` | 0.15 / 0.075 / 0.6 | 0.075 / 0.3 | 0.25 / 0.125 / 1 |
| `text-embedding-3-small` | 0.02 input | 0.01 input | — |
| `text-embedding-3-large` | 0.13 input | 0.065 input | — |

For non-preview web search with `gpt-4.1-mini` or `gpt-4o-mini`, GodarAI adds
the documented fixed 8,000 input tokens per call before spend calculation.

### Images, transcription, speech, moderation, and tools

| Model or tool | Public pricing represented in GodarAI |
|---|---|
| `omni-moderation-latest` | Free |
| `gpt-image-2` Standard | Text input 5; cached text 1.25; image input 8; cached image 2; image output 30 per 1M tokens |
| `gpt-image-2` Batch | Text input 2.5; cached text 0.625; image input 4; cached image 1; image output 15 per 1M tokens |
| `gpt-transcribe` | $0.0045 per audio minute |
| `gpt-4o-mini-transcribe` | Audio input 1.25; text output 5 per 1M tokens |
| `gpt-4o-transcribe-diarize` | Audio input 2.5; text output 10 per 1M tokens |
| `tts-1` | $15 per 1M input characters |
| `tts-1-hd` | $30 per 1M input characters |
| Web search | $10 per 1,000 calls |
| File search | $2.50 per 1,000 calls |
| File-search storage | $0.10 per GB-day after the public free allowance |
| Containers | Public 1 GB, 4 GB, 16 GB, and 64 GB session prices are cataloged |

## Provider rate-limit references

These are catalog reference limits, not limits imposed by GodarAI. OpenAI
selects the account's actual usage tier, and GodarAI forwards live provider
rate-limit headers.

The common GPT-5 Standard profile used by GPT-5.6 Sol/Terra and GPT-5.5/GPT-5.4
at standard context is:

| Tier | RPM | TPM | Batch queue |
|---:|---:|---:|---:|
| 1 | 500 | 500,000 | 1,500,000 |
| 2 | 5,000 | 1,000,000 | 3,000,000 |
| 3 | 5,000 | 2,000,000 | 100,000,000 |
| 4 | 10,000 | 4,000,000 | 200,000,000 |
| 5 | 15,000 | 40,000,000 | 15,000,000,000 |

Luna, mini, nano, Pro, and long-context requests retain their model-page
specific tables. For example, the GPT-5.6 Luna TPM limits are 500K, 2M, 4M,
10M, and 180M for Tiers 1–5. GPT-5.5 Pro uses 50/50K, 500/200K, 500/500K,
1K/1M, and 2K/4M RPM/TPM across Tiers 1–5.

| Models | Tier 1 | Tier 2 | Tier 3 | Tier 4 | Tier 5 |
|---|---|---|---|---|---|
| `gpt-4.1` | 500 RPM / 30K TPM | 5K / 450K | 5K / 800K | 10K / 2M | 10K / 30M |
| `gpt-4.1-mini`, `gpt-4o-mini` | 500 RPM / 200K TPM | 5K / 2M | 5K / 4M | 10K / 10M | 30K / 150M |
| Embedding models | 3K RPM / 1M TPM | 5K / 1M | 5K / 5M | 10K / 5M | 10K / 10M |
| `omni-moderation-latest` | 500 RPM / 10K TPM | 500 / 20K | 1K / 50K | 2K / 250K | 5K / 500K |
| `gpt-image-2` | 100K TPM / 5 IPM | 250K / 20 | 800K / 50 | 3M / 150 | 8M / 250 |
| `gpt-4o-mini-transcribe` | 500 RPM / 50K TPM | 2K / 150K | 5K / 600K | 10K / 2M | 10K / 8M |
| `gpt-4o-transcribe-diarize` | 500 RPM / 10K TPM | 5K / 100K | 5K / 400K | 10K / 2M | 10K / 6M |
| `tts-1`, `tts-1-hd` | 500 RPM | 2.5K | 5K | 7.5K | 10K |

OpenAI does not publish a Tier 1–5 table on the current `gpt-transcribe`
model page, so GodarAI does not invent one. Batch queue limits and
request-per-day values are stored wherever OpenAI publishes them. The linked
model pages and live response headers remain authoritative.

## Remaining incompatibilities

1. **Responses WebSocket mode is not public.** GodarAI exposes all seven REST
   operations and SSE streaming, but not OpenAI's persistent `/v1/responses`
   WebSocket mode. The readiness artifact therefore has
   `websocket_ready: false`.
2. **Realtime transcription is not public.** `gpt-transcribe` file and SSE
   transcription are supported, but committed-turn Realtime WebSocket
   transcription is not available through the regional tunnel.
3. **Fine-tuning management is not public.** `gpt-4.1`, `gpt-4.1-mini`, and
   `gpt-4o-mini` support fine-tuning at OpenAI, but GodarAI does not expose the
   job, event, or checkpoint management routes. Inference with an existing
   fine-tuned model ID remains supported.
4. **New GPT-5.6 tools need explicit productization.** Programmatic Tool
   Calling and beta multi-agent requests may pass through the forward-compatible
   Responses relay, but they are not cataloged or covered by dedicated
   lifecycle/event contract tests.
5. **Some tool-cost reconciliation remains incomplete.** Web-search and
   file-search calls are reconciled, but file-search storage and container
   duration/memory cannot always be attributed from one Responses body and are
   not fully debited by a provider-usage reconciliation worker.
6. **No live canary was run.** Static and automated tests cannot establish
   account entitlement, current provider behavior, or regional availability.

Consequently, REST route and contract readiness are true, while
`websocket_ready`, `production_ready`, and `live_validated` remain false.

## Verification evidence

- Responses REST routing and account affinity:
  [`services/control-plane/internal/proxy/responses_operations_handler.go`](../services/control-plane/internal/proxy/responses_operations_handler.go)
- Request, response, error, and SSE relay:
  [`services/control-plane/internal/proxy/responses_handler.go`](../services/control-plane/internal/proxy/responses_handler.go)
- Specialized model handlers:
  [`services/control-plane/internal/proxy`](../services/control-plane/internal/proxy)
- OpenAI Batch and Files relay:
  [`services/control-plane/internal/proxy/openai_batch_files_handler.go`](../services/control-plane/internal/proxy/openai_batch_files_handler.go)
- Pricing and usage reconciliation:
  [`services/control-plane/internal/billing/service.go`](../services/control-plane/internal/billing/service.go)
- Runtime-profile schema and importer:
  [`apps/model-discovery/scripts/import-models-from-yaml.ts`](../apps/model-discovery/scripts/import-models-from-yaml.ts)
- Shared runtime contract:
  [`packages/contracts/platform/model_runtime_profile.go`](../packages/contracts/platform/model_runtime_profile.go)
- Pinned OpenAI Responses route baseline:
  [`packages/contracts/openapi/openai-scoped-routes-db14b6e.json`](../packages/contracts/openapi/openai-scoped-routes-db14b6e.json)
- Generated readiness artifact:
  [`services/control-plane/openai-readiness-report.json`](../services/control-plane/openai-readiness-report.json)

Automated verification recorded for the cumulative implementation:

- all seven Responses REST routes compared with the pinned OpenAI baseline;
- targeted Responses request, response, error, SSE, and lifecycle tests;
- targeted specialized proxy and image-pricing tests;
- billing and platform-contract tests;
- model-discovery production build;
- dry-run import of all 149 OpenAI catalog models;
- Go formatting and repository diff validation.

The monolithic proxy test process exceeded the Windows page-file limit during
the earlier implementation run. The affected suites passed when run in
targeted partitions.

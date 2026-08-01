# Deepgram full-support gap analysis and implementation plan

**Status:** Proposed  
**Scope:** The Deepgram models currently present in `apps/model-discovery/providers/deepgram/`  
**Analysis date:** 2026-08-01  
**Implementation in this change:** None

## 1. Executive summary

GodarAI does **not** currently provide full Deepgram support.

- `deepgram/nova-3` is **partially supported**. A synchronous uploaded-file transcription can reach Deepgram, but GodarAI forwards only `language`, discards almost the entire Deepgram response, does not expose streaming or remote-URL input, and cannot reliably charge the exact provider cost.
- `deepgram/flux-general-en` and `deepgram/flux-general-multi` are **unsupported**. Flux is a WebSocket-only `/v2/listen` model; GodarAI exposes no realtime WebSocket proxy route and its Deepgram adapter explicitly rejects streaming.
- `deepgram/aura-2-arcas-en`, `deepgram/aura-2-javier-es`, and `deepgram/aura-2-thalia-en` are **unsupported**. GodarAI's speech endpoint rejects Deepgram and Silk Road has no Deepgram TTS adapter.
- The current seven files in the Deepgram discovery folder are zero-byte working-tree files. The importer cannot turn them into active, routable, priced models. Historical tracked Deepgram files were only minimal `mode`, `model`, and `supportedModes` records and were therefore insufficient for production support too.
- The provider credential schema exists, but both platform-managed credential validation and tenant BYOK validation reject or do not implement Deepgram. Regional base URLs are not allowlisted.
- Deepgram is missing from runtime catalog pricing initialization and the billing provider mapping. The current fallback can produce zero or estimated charges instead of failing closed. Deepgram's per-second, multichannel, tier-, language-, MIP-, and add-on-sensitive pricing cannot be represented and selected correctly by the present one-rate-per-dimension path.

The recommended program is four delivery phases: establish complete catalog and credential contracts; make Nova-3 prerecorded transcription correct and billable; add Aura-2 REST TTS; then add the realtime WebSocket data plane required by Flux, Nova-3 live transcription, and Aura-2 streaming. Full support is complete only after all six model rows pass contract, billing, resilience, and end-to-end tests.

### Decision summary

| Decision | Recommendation |
|---|---|
| Public API compatibility | Keep OpenAI-compatible `/v1/audio/transcriptions` and `/v1/audio/speech`; add an explicit GodarAI realtime WebSocket contract and a namespaced `provider_options.deepgram` extension. |
| Response fidelity | Keep the existing normalized default, add a provider-neutral verbose response, and offer explicit Deepgram-native payload/event passthrough. Do not silently discard requested Deepgram fields. |
| Catalog scope | Support the six current model files. Treat `default.yaml` as provider defaults, not a model. Do not silently restore the deleted historical catalog. |
| Credentials | Store project API keys using the existing encrypted/versioned connection system; validate with the documented auth probe; mint short-lived JWTs only when a client-direct realtime design explicitly requires them. |
| Billing | Fail closed when a Deepgram price cannot be selected. Finalize STT from exact processed duration and channel count, and TTS from `dg-char-count`. Model base and add-on charges as additive components. |
| Async callbacks | Defer raw provider callbacks. Introduce a GodarAI-owned async job/webhook contract later, with callback allowlisting and signatures. |

## 2. Scope, sources, and repository state

### In scope

The current working tree contains these Deepgram entries:

| File | Catalog role |
|---|---|
| `default.yaml` | Provider defaults and metadata; not a user-selectable model |
| `nova-3.yaml` | Nova-3 speech to text |
| `flux-general-en.yaml` | Flux English realtime speech to text |
| `flux-general-multi.yaml` | Flux multilingual realtime speech to text |
| `aura-2-arcas-en.yaml` | Aura-2 English TTS voice |
| `aura-2-javier-es.yaml` | Aura-2 Mexican Spanish TTS voice |
| `aura-2-thalia-en.yaml` | Aura-2 English TTS voice |

All seven are zero bytes as of the analysis date. The three Aura files are modified tracked files; `default.yaml`, both Flux files, and `nova-3.yaml` are untracked. Approximately 140 historical Deepgram model files are deleted in the same uncommitted working tree. This report treats that state as user-owned work, does not modify it, and scopes the implementation plan to the six current model names. Whether the historical catalog should return is an explicit product decision, not an inference.

### Sources reviewed

- Current GodarAI model discovery, control-plane, Silk Road tunnel, billing, contracts, tests, Postman collections, documentation, and related plans.
- Repository LiteLLM notes in [`.llmLite_docs/providers/deepgram.md`](../.llmLite_docs/providers/deepgram.md), [audio transcription](../.llmLite_docs/audio_transcription.md), [text to speech](../.llmLite_docs/text_to_speech.md), [model discovery](../.llmLite_docs/proxy/model_discovery.md), and [credential UI](../.llmLite_docs/proxy/ui_credentials.md). These are comparative evidence only: the local LiteLLM material covers Deepgram prerecorded transcription but does not provide full Flux or Aura support.
- Existing GodarAI architecture and credential plans, especially [the repository plan index](README.md), [provider support](phase2-liteLLM-provider-support.md), [platform-managed credentials](platform-managed-provider-credentials-product-analysis.md), and [provider credential matrix](platform-managed-provider-credentials-provider-matrix.md).
- Official Deepgram documentation and pricing, verified 2026-08-01. Direct links appear beside the facts they support.

### Evidence convention

“Current” means directly observed in this repository. “Deepgram contract” means documented by Deepgram. “Recommendation” means the proposed GodarAI behavior. Prices are volatile external data and must carry source and effective timestamps rather than being copied as timeless constants.

The existing focused tests were also run during this review and passed: Silk Road's prerecorded Deepgram dispatch test, the control-plane native-provider routing and streaming-rejection tests, and the platform credential unsupported-probe test. These passes confirm the narrow current behavior; they are not evidence of full model support.

## 3. Current architecture and request flow

The existing prerecorded request path is:

```text
Client
  -> control-plane POST /v1/audio/transcriptions
  -> model/deployment routing and wallet reservation
  -> mTLS tunnel request
  -> Silk Road Deepgram adapter
  -> POST https://api.deepgram.com/v1/listen
  -> response collapsed to {"text":"..."}
  -> estimated-usage billing finalization
```

The control plane mounts only the audio transcription and speech HTTP endpoints; there is no model-facing realtime WebSocket route in the proxy registration ([`routes_proxy.go:68-74`](../services/control-plane/internal/server/routes_proxy.go#L68)). Although shared contracts define tunnel WebSocket interfaces ([`types.go:700-741`](../packages/contracts/platform/types.go#L700)), the readiness command reports WebSocket support as false ([`openai-readiness/main.go:115-129`](../services/control-plane/cmd/openai-readiness/main.go#L115)). A contract without an implementation does not make Flux routable.

The router recognizes a `deepgram/` model prefix and can resolve either tenant or platform credentials ([`router.go:129-173`](../services/control-plane/internal/router/router.go#L129), [`router.go:467-533`](../services/control-plane/internal/router/router.go#L467)). Deployment contracts already carry API base, credential source, billing mode, and resolved provider auth ([`types.go:260-289`](../packages/contracts/platform/types.go#L260)). These are useful primitives, but they do not compensate for the missing endpoint adapters, model profiles, and billing rules.

### Existing Nova-3 behavior

The control plane classifies Deepgram transcription as native but non-streaming ([`audio_transcriptions_handler.go:1298-1323`](../services/control-plane/internal/proxy/audio_transcriptions_handler.go#L1298)) and rejects a streaming request before dispatch ([`audio_transcriptions_handler.go:221-226`](../services/control-plane/internal/proxy/audio_transcriptions_handler.go#L221)). Silk Road registers Deepgram to one transcription-only adapter ([`adapters.go:31-66`](../services/silk-road/pkg/server/adapters.go#L31)); that adapter accepts only `/audio/transcriptions`, rejects streaming, and defaults to `api.deepgram.com` ([`audio_transcriptions.go:41-59`](../services/silk-road/pkg/server/audio_transcriptions.go#L41)).

The adapter reads the multipart upload into memory ([`audio_transcriptions.go:167-213`](../services/silk-road/pkg/server/audio_transcriptions.go#L167)), builds `/v1/listen`, and forwards only `model` and `language` ([`audio_transcriptions.go:215-235`](../services/silk-road/pkg/server/audio_transcriptions.go#L215)). It then extracts the first channel's first alternative and replaces the provider response with only `{text}` ([`audio_transcriptions.go:342-395`](../services/silk-road/pkg/server/audio_transcriptions.go#L342)). The existing unit test verifies that deliberately narrow contract ([`server_test.go:663-724`](../services/silk-road/pkg/server/server_test.go#L663)).

### Existing Aura-2 behavior

The control plane's TTS provider allowlist excludes Deepgram ([`audio_speech_handler.go:1005-1011`](../services/control-plane/internal/proxy/audio_speech_handler.go#L1005)), so requests are rejected at the provider gate ([`audio_speech_handler.go:210-215`](../services/control-plane/internal/proxy/audio_speech_handler.go#L210)). Silk Road maps the Deepgram provider exclusively to the transcription adapter, which rejects any other operation ([`adapters.go:55-60`](../services/silk-road/pkg/server/adapters.go#L55), [`audio_transcriptions.go:41-46`](../services/silk-road/pkg/server/audio_transcriptions.go#L41)).

## 4. Model-by-model support matrix

Deepgram documents Nova-3 as supporting prerecorded and streaming transcription, Flux as a streaming voice-agent model, and Aura-2 as text-to-speech ([models and languages](https://developers.deepgram.com/docs/models-languages-overview), verified 2026-08-01). Flux specifically requires WebSocket `/v2/listen`; `/v1/listen` does not support it ([Flux quickstart](https://developers.deepgram.com/docs/flux/quickstart), verified 2026-08-01).

| GodarAI model | Deepgram modality and transport | Current status | Evidence-backed gap | Target state |
|---|---|---|---|---|
| `deepgram/nova-3` | Audio or media URL → transcript; REST prerecorded and WebSocket streaming | **Partial** | Uploaded binary REST only; one parameter; normalized text only; no exact duration/channel usage; no URL, callback, intelligence fields, diarization, keyterm, or live stream | Complete REST file/URL transcription, rich/native responses, supported options, exact billing, and realtime WS |
| `deepgram/flux-general-en` | Live English audio → transcript and turn events; WebSocket `/v2/listen` only | **Unsupported** | No public or tunnel WS data plane; current adapter rejects streaming and builds `/v1/listen` | Bidirectional WS, binary audio frames, Flux configuration/events, backpressure, reconnect and billing |
| `deepgram/flux-general-multi` | Live multilingual audio → transcript and turn events; WebSocket `/v2/listen` only | **Unsupported** | Same transport gap; no multilingual/language-hint metadata or contract | Same as Flux English plus supported multilingual configuration and catalog metadata |
| `deepgram/aura-2-arcas-en` | English text → audio; REST chunked response and WS streaming | **Unsupported** | TTS allowlist rejects provider; no adapter; catalog file empty | REST and WS TTS, voice/model validation, output controls, header-based usage |
| `deepgram/aura-2-javier-es` | Mexican Spanish text → audio; REST and WS | **Unsupported** | Same as above; language metadata absent | Same as above with `es-MX` voice metadata |
| `deepgram/aura-2-thalia-en` | English text → audio; REST and WS | **Unsupported** | Same as above | Same as above with `en-US` voice metadata |

The voice metadata should be sourced from Deepgram's [Aura-2 model list](https://developers.deepgram.com/docs/tts-models), not inferred from the filename. Flux English currently supports English; Flux multilingual currently documents English, Spanish, French, German, Hindi, Russian, Portuguese, Japanese, Italian, and Dutch. Nova-3's documented language set includes Persian (`fa`), which is strategically relevant to GodarAI; language availability should be catalog data with a `verified_at` field rather than hard-coded UI copy.

## 5. Provider credentials, validation, and regions

### Current support

The discovery migration defines a Deepgram `api_key` credential schema, marks it write-only, maps it to `authorization_token`, and provides `https://api.deepgram.com` as the default base URL ([`migration.sql:90-93`](../apps/model-discovery/prisma/migrations/202607170001_platform_provider_connection_schemas/migration.sql#L90)). The general credential system already supports encrypted versioned material and resolution into a deployment ([`platformcredentials/types.go:17-75`](../services/control-plane/internal/platformcredentials/types.go#L17)). The Silk Road adapter correctly uses Deepgram's `Authorization: Token …` header for the outbound call ([`audio_transcriptions.go:64-82`](../services/silk-road/pkg/server/audio_transcriptions.go#L64)).

The credential is not operationally complete:

- The platform HTTP prober returns `unsupported` when no provider probe is implemented ([`platformcredentials/service.go:273-280`](../services/control-plane/internal/platformcredentials/service.go#L273)); its documented probe switch and safe-host allowlist do not include Deepgram ([`platformcredentials/service.go:338-454`](../services/control-plane/internal/platformcredentials/service.go#L338)).
- Tenant provider validation has no Deepgram branch and reports unsupported providers as invalid ([`provider_credential_validation.go:21-35`](../services/control-plane/internal/storage/provider_credential_validation.go#L21), [`provider_credential_validation.go:104-121`](../services/control-plane/internal/storage/provider_credential_validation.go#L104)).
- API base validation excludes Deepgram hosts ([`provider_api_base_validation.go:163-197`](../services/control-plane/internal/storage/provider_api_base_validation.go#L163)), despite the schema exposing a configurable base URL.
- The repository's contract test explicitly expects the Deepgram platform probe to be unsupported ([`provider_contracts_test.go:75-89`](../services/control-plane/internal/platformcredentials/provider_contracts_test.go#L75)).

### Required contract

Deepgram accepts a project API key in `Authorization: Token <key>` and short-lived JWTs in `Authorization: Bearer <token>` ([authentication reference](https://developers.deepgram.com/reference/authentication), verified 2026-08-01). GodarAI should:

1. Store project API keys only in the existing encrypted/versioned credential store. Treat them as opaque non-empty values; do not impose a guessed format.
2. Validate without a billable inference by calling `GET /v1/auth/token`, using the documented authentication test flow ([authentication guide](https://developers.deepgram.com/guides/fundamentals/authenticating), verified 2026-08-01). Distinguish invalid auth, forbidden scope, network failure, and unsupported region.
3. Allow the official regional hosts `api.deepgram.com`, `api.eu.deepgram.com`, and `api.au.deepgram.com`; require an operator-approved egress policy for self-hosted or enterprise domains. Never accept arbitrary tenant-supplied hosts.
4. Use API keys server-side by default. If a future client-direct WebSocket design is approved, mint a 1–3600 second temporary JWT and never expose the project key ([temporary token reference](https://developers.deepgram.com/reference/auth/tokens/grant), verified 2026-08-01).
5. Preserve explicit credential-source and billing-mode selection. Do not silently fall back from BYOK to a platform-funded credential; the router's shared-secret fallback path must be gated by the account's declared mode ([`router.go:796-889`](../services/control-plane/internal/router/router.go#L796)).

## 6. Catalog, modalities, capabilities, and lifecycle

### Why the present catalog cannot activate these models

The importer recognizes `audio_transcription`, `text_to_speech`, and `realtime` modes ([`import-models-from-yaml.ts:32-46`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L32)), but an absent or unknown status becomes unavailable ([`import-models-from-yaml.ts:292-303`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L292)). Provider-specific enrichment exists for several providers, not Deepgram ([`import-models-from-yaml.ts:513-529`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L513)). Runtime profiles are read from YAML ([`import-models-from-yaml.ts:533-600`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L533)), but Deepgram is not among the providers for which a missing profile is rejected at import time ([`import-models-from-yaml.ts:1127-1163`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L1127)).

Only models with runtime profiles enter the runtime snapshot; an empty result produces `no_eligible_models` ([`catalog.service.ts:29-63`](../apps/model-discovery/src/catalog/catalog.service.ts#L29)). The runtime profile schema already names audio speech, transcription, realtime, and realtime-transcription endpoints and supports HTTP, SSE, and WebSocket transports ([`runtime-profile.schema.ts:32-160`](../apps/model-discovery/src/catalog/runtime-profile.schema.ts#L32)). It should be extended rather than bypassed.

### Required metadata

Each of the six model files needs:

- Stable GodarAI model FQN, provider model ID, display name, family, version/alias behavior, lifecycle status, release/verification dates, and official documentation URL.
- Input/output modalities, batch/realtime support, endpoint and transport bindings, default timeout, retry class, request size limits, concurrency notes, regions, and billing dimensions.
- Supported languages and voice locale. Keep language lists data-driven because Deepgram changes them independently of GodarAI releases.
- Typed capability flags and parameter schemas. The current limited capability taxonomy ([`import-models-from-yaml.ts:123-181`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L123)) needs audio-specific additions instead of hiding Deepgram behavior in arbitrary strings.
- A complete runtime profile. The Go mirror supports endpoints, transports, usage accounting, and provider-parameter declarations ([`model_runtime_profile.go:9-125`](../packages/contracts/platform/model_runtime_profile.go#L9)), but its limits need audio duration, sample rate, channels, text characters, realtime session duration, and connection concurrency ([`model_runtime_profile.go:152-161`](../packages/contracts/platform/model_runtime_profile.go#L152)).

Recommended capability groups:

| Group | Examples |
|---|---|
| STT input | file, raw binary, remote URL, encoding, sample rate, channels, multichannel |
| STT transcript | language/detection, confidence, word timestamps, utterances, paragraphs, diarization, smart format, punctuation, numerals, profanity filtering |
| STT intelligence | keyterm, keywords, redaction, search/replace, sentiment, summarize, topics, intents, entities |
| Realtime | interim results, endpointing, turn events, configure event, keepalive, reconnect semantics |
| TTS | encoding, container, sample rate, bit rate, speed, pronunciation, chunked REST, WebSocket commands |
| Operations | callback support, tags, `mip_opt_out`, regional endpoint, provider request ID |

Capabilities must be model-scoped. For example, Flux supports its own end-of-turn configuration and event model, while Nova-3 supports keyterm prompting. Do not mark a feature at provider level if one listed model cannot execute it.

## 7. Endpoint and REST/WebSocket API contracts

### 7.1 Nova-3 prerecorded transcription

Deepgram's contract is `POST /v1/listen` with either audio bytes or `{ "url": "…" }`, plus query parameters for transcription, formatting, intelligence, callback, tagging, and MIP behavior. A direct response contains metadata such as request ID, duration, channels, models, alternatives, words, and optional intelligence results ([prerecorded listen reference](https://developers.deepgram.com/reference/speech-to-text/listen-pre-recorded), verified 2026-08-01).

Recommended GodarAI contract:

- Retain `POST /v1/audio/transcriptions`.
- Accept multipart `file` and JSON remote `url` as mutually exclusive inputs. Remote URL support needs a deliberate privacy policy, scheme validation, redacted logging, and a documented responsibility boundary because Deepgram—not GodarAI—fetches the media.
- Preserve OpenAI-compatible fields where meanings align. Put Deepgram-only controls under `provider_options.deepgram`, validate them against the selected model profile, and reject unknown or contradictory combinations before reserving cost.
- Keep `{text}` as the default backward-compatible response.
- Add `response_format=verbose_json` for normalized language, duration, channels, utterances, words, timestamps, confidence, and speaker data.
- Add an explicit native response mode for users who require every Deepgram field. Never change the default to an unstable provider payload and never silently discard requested fields.
- Return the provider request ID in a safe response header and trace metadata.
- Support synchronous prerecorded requests first. Defer provider-direct `callback` to a GodarAI async-job/webhook design with signed delivery and callback-domain policy.

The current public parser requires a file for non-Mistral providers and validates an OpenAI-centric set of response formats ([`audio_transcriptions_handler.go:640-748`](../services/control-plane/internal/proxy/audio_transcriptions_handler.go#L640)). The adapter must stop discarding metadata before exact usage and verbose/native responses can work.

### 7.2 Realtime transcription: Nova-3 and Flux

Recommendation: add `GET /v1/realtime?intent=transcription` as a public bidirectional WebSocket contract, with an optional compatibility alias only if an SDK requires it. The session configuration selects the model, format, sample rate, channel count, language, and `provider_options.deepgram`. Audio is sent as binary frames; control and result messages are JSON.

- Nova-3 maps to Deepgram `/v1/listen` WebSocket.
- Flux maps to `/v2/listen` WebSocket and exposes `eot_threshold`, `eager_eot_threshold`, and `eot_timeout_ms` plus Configure events. Flux-specific `EagerEndOfTurn`, `TurnResumed`, and `EndOfTurn` events must survive either as normalized event types with a `provider_event` field or explicit native passthrough ([Flux configuration](https://developers.deepgram.com/docs/flux/configuration), verified 2026-08-01).
- The gateway must implement bounded queues, backpressure, maximum session duration, idle timeout, ping/keepalive, close-code mapping, client cancellation, upstream reconnect rules, and tenant concurrency limits.
- Do not automatically replay audio after an ambiguous disconnect: duplicate transcription and duplicate billing are worse than an explicit resumability error unless the provider offers a proven idempotent resume mechanism.

### 7.3 Aura-2 speech

Deepgram's REST contract is `POST /v1/speak?model=…` with `{ "text": "…" }`; output is audio and response headers include model identity, request ID, and `dg-char-count` ([speak reference](https://developers.deepgram.com/reference/text-to-speech/speak-request), [TTS guide](https://developers.deepgram.com/docs/text-to-speech), verified 2026-08-01).

Recommended GodarAI contract:

- Retain `POST /v1/audio/speech` and map its input to Deepgram. The selected GodarAI model already identifies the voice. Accept `voice` for OpenAI SDK compatibility, but require it to be `default`, the exact provider voice, or a documented alias; reject mismatches.
- Map supported output settings to Deepgram `encoding`, `container`, `sample_rate`, `bit_rate`, and `speed`. Deepgram documents speed `0.7–1.5`; the current generic parser accepts `0.25–4.0`, so validation must be model-specific ([voice controls](https://developers.deepgram.com/docs/tts-voice-controls), verified 2026-08-01; current parser at [`audio_speech_handler.go:693-841`](../services/control-plane/internal/proxy/audio_speech_handler.go#L693)).
- Enforce Deepgram's 2,000-character input limit for Aura-2 instead of the current 4,096-character generic limit. Surface inline pronunciation controls only for supported languages and validate their limits.
- Stream the REST response through without buffering the complete audio. Preserve safe content headers and capture `dg-char-count`, model, and request ID for billing and observability.
- Add Aura WebSocket streaming in the realtime phase, including Speak/Flush/Clear/Close control semantics and bounded buffers. HTTP chunked audio is not equivalent to a bidirectional TTS WebSocket.

### 7.4 Unsupported/deferred contracts

- **Include now:** the six listed models, Nova-3 file and remote-URL REST, rich/native response options, Aura REST, realtime Nova/Flux/Aura, regions, MIP, tags, and model-valid controls.
- **Defer:** raw Deepgram callback URLs, provider conversational-agent APIs, batch management beyond the listed model contracts, self-hosted domains without operator egress configuration, and a public API for Deepgram project administration.
- **Exclude from this scope:** historical deleted model names and every unlisted Aura voice. They require an explicit catalog decision and the same evidence bar before restoration.

## 8. Tiering and pricing

### Official pricing facts

The official pricing page, verified 2026-08-01, publishes per-minute STT and per-1,000-character TTS rates, separate Pay As You Go and Growth tiers, add-on prices, and temporary/promotional price presentation ([Deepgram pricing](https://deepgram.com/pricing)). It also states per-second billing without rounding and charges multichannel audio as duration multiplied by channel count.

At verification time the page showed these active public rates; they are discovery inputs, not timeless constants:

| Product | Pay As You Go | Growth |
|---|---:|---:|
| Flux English | $0.0065/min | $0.0057/min |
| Flux multilingual | $0.0078/min | $0.0068/min |
| Nova-3 monolingual | $0.0048/min | $0.0042/min |
| Nova-3 multilingual | $0.0058/min | $0.0050/min |
| Aura-2 | $0.030/1K chars | $0.027/1K chars |
| Redaction add-on | $0.0020/min | $0.0017/min |
| Keyterm add-on | $0.0013/min | $0.0012/min |
| Diarization add-on | $0.0020/min | $0.0017/min |

The page displayed some higher struck/list rates beside active rates. GodarAI must store the chargeable active price with `source_url`, `observed_at`, `effective_from`, tier, currency, and promotional status; list price may be retained separately for UI comparison but must never be selected accidentally. Published Growth prices do not prove that a particular GodarAI provider account has Growth terms. Enterprise contract prices must be connection/deployment overrides.

### Current pricing gaps

The importer can map per-character and per-minute inputs to `speech_input_characters` and `audio_input_minutes` ([`import-models-from-yaml.ts:752-760`](../apps/model-discovery/scripts/import-models-from-yaml.ts#L752)), but exact Deepgram STT is per second and conditional. Runtime pricing rejects duplicate dimensions for most providers ([`catalog_pricing_snapshot.go:145-180`](../services/control-plane/internal/billing/catalog_pricing_snapshot.go#L145)), preventing multiple conditional Deepgram rates with the same usage dimension. Only Mistral currently has provider-specific completeness checks ([`catalog_pricing_snapshot.go:197-264`](../services/control-plane/internal/billing/catalog_pricing_snapshot.go#L197)).

Deepgram is not initialized as a catalog pricing provider in the server ([`cmd/server/main.go:164-260`](../services/control-plane/cmd/server/main.go#L164)), is absent from the billing provider mapping ([`billing/service.go:2331-2355`](../services/control-plane/internal/billing/service.go#L2331)), and has no static fallback registry entry ([`billing/pricing.go:24-255`](../services/control-plane/internal/billing/pricing.go#L24)). The metered path selects a single rate by dimension and only fails closed for a provider-specific Mistral case ([`billing/service.go:924-992`](../services/control-plane/internal/billing/service.go#L924)). This is unsafe for a provider whose cost varies by tier, language mode, MIP choice, channels, and additive features.

The transcription handler estimates usage from file bytes when provider usage is unavailable ([`audio_transcriptions_handler.go:1059-1104`](../services/control-plane/internal/proxy/audio_transcriptions_handler.go#L1059)) and parses only top-level `usage.seconds`-style fields ([`audio_transcriptions_handler.go:1106-1162`](../services/control-plane/internal/proxy/audio_transcriptions_handler.go#L1106)). Deepgram reports duration in `metadata.duration`, but the adapter deletes it. TTS already meters `speech_input_characters` ([`audio_speech_handler.go:140-146`](../services/control-plane/internal/proxy/audio_speech_handler.go#L140)), although it does not yet reconcile with Deepgram's exact `dg-char-count`.

### Required pricing model

1. Add first-class `audio_input_seconds` and retain `speech_input_characters`. Store provider unit prices in those exact units; display per-minute or per-1K-character prices separately.
2. Select STT base rates by provider connection pricing plan, concrete model, monolingual/multilingual mode, and MIP setting. Do not infer Growth from the model name.
3. Multiply billed STT seconds by channel count when multichannel is active.
4. Price redaction, keyterm, and diarization as additive components. A bill must be able to contain base transcription plus zero or more add-ons, each with quantity, rate snapshot, and subtotal.
5. Reserve from validated request settings using a conservative duration estimate and reconcile after the response. For REST, finalize from Deepgram metadata before normalization. For TTS, finalize from `dg-char-count`. For live raw PCM, track audio frames from encoding/sample-rate/channel facts and reconcile with provider usage when available.
6. Reject before upstream dispatch when no unambiguous rate exists. Never turn a catalog/pricing miss into zero-cost usage.
7. Snapshot selected price IDs, conditions, source timestamp, provider request ID, quantity, channel multiplier, add-ons, and provider-account tier in the ledger for audit and later reconciliation.
8. Add a scheduled pricing-freshness check. A stale or changed promotional rate should block platform-funded traffic according to policy; BYOK can continue only if GodarAI is not using the rate for wallet charging.

## 9. Error handling, retries, limits, and resiliency

Deepgram errors include provider fields such as `err_code`, `err_msg`, and `request_id` ([Deepgram errors](https://developers.deepgram.com/docs/errors), verified 2026-08-01). The current generic parser does not understand `err_msg`, so it can surface a serialized JSON body as the message ([`errors.go:85-121`](../services/control-plane/internal/proxy/errors.go#L85)). Its safe header allowlist also omits Deepgram request and character-count headers ([`errors.go:128-194`](../services/control-plane/internal/proxy/errors.go#L128)).

Required behavior:

- Normalize Deepgram errors into GodarAI's error envelope while retaining `provider=deepgram`, safe provider code, request ID, HTTP status, and retryability. Never return secrets or the original authorization header.
- Treat validation/auth/permission/unsupported media/payload-too-large errors as non-retryable. Retry transport failures, 429, and selected 5xx responses only when the operation is safe; honor `Retry-After` and use bounded jittered backoff.
- Do not blindly retry a partially consumed realtime or streaming TTS session. Emit a terminal event with provider request ID and billed/processed usage so the client can decide whether to start a new session.
- Enforce per-model limits before dispatch. Deepgram documents prerecorded files up to 2 GB and provider concurrency constraints ([prerecorded audio guide](https://developers.deepgram.com/docs/pre-recorded-audio), [API rate limits](https://developers.deepgram.com/reference/api-rate-limits), verified 2026-08-01), but accepting 2 GB into the current in-memory `[]byte` tunnel contract would be a denial-of-service risk. GodarAI should choose a lower configurable public limit and stream or spool content instead of `io.ReadAll`.
- Add tenant and credential-level concurrency controls for REST and WebSocket sessions. Feed 429/limit observations into health and circuit-breaker state without globally disabling every region or credential.
- Make timeouts phase-specific: upload, provider first byte, total prerecorded processing, WS handshake, idle, and maximum session lifetime.

## 10. Security, privacy, and observability

The trace builders already avoid recording raw transcription audio and TTS audio, and hash or summarize textual fields ([`trace_payloads.go:570-633`](../services/control-plane/internal/proxy/trace_payloads.go#L570), [`trace_payloads.go:662-709`](../services/control-plane/internal/proxy/trace_payloads.go#L662)). Preserve this baseline while adding provider metadata.

Required controls:

- Never log audio bytes, transcript text, TTS input, API keys, temporary JWTs, signed media URLs, pronunciation payloads, or raw Deepgram headers by default. Keep content logging opt-in, tenant-controlled, access-audited, and retention-limited.
- Record provider request ID, model, region, endpoint, latency phases, response status, retry count, audio seconds, channels, character count, selected capabilities, rate ID, and cost components. Hash tenant-provided tags if their contents are not explicitly classified as safe.
- Treat `mip_opt_out` as both a privacy and pricing choice. Its resolved value should be explicit in deployment policy, request audit metadata, and the selected price conditions.
- Enforce SSRF and data-egress policy for remote media URLs. Do not persist full signed URLs; redact query strings and fragments.
- Keep platform API keys server-side. If temporary JWTs are used, scope their TTL to the expected session, bind issuance to an authenticated tenant/model/deployment decision, rate-limit grants, and audit issuance without logging the token.
- For raw callbacks, require HTTPS, resolve and pin allowed destinations safely, block private/link-local/metadata networks, sign GodarAI webhooks, and prevent cross-tenant request-ID disclosure. This is why provider-direct callbacks are deferred.
- Validate content types and audio parameters; cap decompressed/decoded size, multipart fields, JSON depth, WS message size, queue depth, and session duration.

Observability should distinguish `route`, `credential`, `connect`, `upload`, `provider_processing`, `first_result`, and `download` latency. Realtime dashboards need active sessions, rejected sessions by limit, buffered bytes, dropped/backpressured frames, turn-event counts, close codes, audio seconds, and cost drift. Alerts should detect missing exact-usage metadata, unknown prices, response-normalization loss, and sustained differences between estimated and reconciled cost.

## 11. Implementation blueprint

### Ownership boundaries

| Component | Responsibility |
|---|---|
| Model discovery | Provider defaults, model lifecycle, modes, capabilities, parameter schemas, runtime profiles, public baseline prices |
| Control plane | Authentication, tenant/model authorization, credential resolution, routing, request validation, wallet reservation/finalization, normalized public contracts, traces |
| Silk Road | Provider protocol adaptation, regional upstream connection, streaming transport, safe header/metadata extraction, cancellation/backpressure |
| Billing | Conditional price selection, additive components, rate snapshots, exact usage reconciliation, fail-closed policy |
| Tenant admin/docs | Credential setup/validation, region/tier/MIP configuration, model/capability visibility, examples and limitations |

### Data and contract changes

1. Add a Deepgram discovery importer/validator and complete all seven YAML files. Reject any active Deepgram model without a runtime profile, endpoint binding, auth contract, usage rule, and price selection rule.
2. Extend capability and limit schemas with audio/realtime concepts. Add typed Deepgram parameter schemas per endpoint; unknown provider options should be rejected.
3. Add conditional and additive rate support, plus `audio_input_seconds`. Tie non-public plan pricing to a provider connection/deployment override.
4. Implement Deepgram platform/BYOK validation and regional host policy.
5. Refactor Silk Road's Deepgram registration into a provider adapter that dispatches prerecorded STT, realtime STT/Flux, REST TTS, and realtime TTS by operation.
6. Change tunnel request bodies from required in-memory byte slices to bounded streaming readers or safe spool handles. Implement the WebSocket tunnel interface end to end.
7. Preserve raw provider usage and response metadata internally before applying normalized/native response selection.
8. Add Deepgram-aware error parsing and safe response header extraction.

### Phased rollout

#### Phase 0 — Contracts, catalog, credentials, and pricing foundation

- Resolve whether the current six-model catalog intentionally replaces the deleted historical catalog.
- Complete and validate Deepgram YAML/runtime profiles and audio capability taxonomy.
- Add credential probes, regional host policy, provider-account pricing-plan and MIP configuration.
- Add Deepgram catalog pricing initialization, conditional/additive pricing, exact usage dimensions, and fail-closed tests.
- Publish API schemas and examples for normalized, verbose, native, and realtime event contracts.

**Exit criteria:** all six models appear correctly but remain feature-gated; credentials validate in each enabled region; ambiguous/missing price tests fail before provider dispatch.

#### Phase 1 — Correct Nova-3 prerecorded support

- Stream/spool multipart payloads safely; add remote URL input.
- Map and validate documented Nova-3 options.
- Preserve metadata/results, implement verbose/native responses, and normalize provider errors.
- Finalize duration × channels and additive feature cost from exact response facts.
- Keep callbacks disabled or implement only a GodarAI-owned async job contract.

**Exit criteria:** prerecorded file and URL golden tests pass against a Deepgram sandbox; requested timestamps/diarization/keyterm survive; cost reconciles exactly; no raw media/content appears in standard logs.

#### Phase 2 — Aura-2 REST TTS

- Enable Deepgram in the speech provider gate and add `/v1/speak` adaptation.
- Implement model/voice compatibility, text and voice-control limits, output format mapping, binary streaming, safe header capture, and exact character billing.
- Add all three listed voices to catalog/admin/docs.

**Exit criteria:** each voice returns playable audio in supported formats; invalid voice/model and speed combinations fail locally; `dg-char-count` drives final cost; cancellation releases buffers and reservations.

#### Phase 3 — Realtime data plane

- Implement authenticated public and tunnel WebSockets, connection limits, backpressure, cancellation, idle/session timeouts, and trace events.
- Add Nova-3 `/v1/listen`, Flux `/v2/listen`, and Aura realtime adapters.
- Preserve Flux turn events and Deepgram control/event semantics through normalized/native modes.
- Implement realtime exact-usage accounting and reconciliation.

**Exit criteria:** Flux English/multilingual, Nova-3 live, and all three Aura voices pass sustained, cancellation, slow-client, provider-close, and concurrency tests; readiness reports WebSocket support only after deployment health is proven.

#### Phase 4 — Production hardening and general availability

- Canary by region, provider credential, and model; compare provider and GodarAI usage/cost.
- Run soak, fault-injection, large-input, memory, connection-leak, and privacy tests.
- Regenerate docs and Postman/SDK examples; enable tenants progressively with rollback switches.

**Exit criteria:** no unknown-price traffic, cost drift within the agreed threshold, SLOs met, audit/security sign-off complete, and every model row below is green.

## 12. Testing, rollout, and acceptance criteria

The existing Deepgram Postman collection is a disabled placeholder rather than executable coverage ([`GodarAI - deepgram.postman_collection.json:1-28`](../tests/postman/collections/GodarAI%20-%20deepgram.postman_collection.json#L1)); its manifest records placeholder execution and no provider docs ([`provider-manifest.json:191-198`](../tests/postman/provider-manifest.json#L191)). Replace it with capability-focused tests for the six current models, not a model-count assertion based on the deleted historical catalog.

### Required test layers

- **Catalog validation:** status, modality, language, capabilities, runtime profile, parameter schema, usage dimensions, price conditions, source/effective dates, and model alias behavior.
- **Credential tests:** valid/invalid/revoked key; insufficient scope; NA/EU/AU hosts; blocked arbitrary host; rotation; BYOK versus platform-funded selection; no secret in errors/traces.
- **Adapter unit/golden tests:** every request mapping and response/error shape; binary and URL STT; rich fields; Aura formats/headers; Flux and WS event sequences.
- **Billing tests:** mono/multi, tier, MIP, multiple channels, redaction/keyterm/diarization combinations, exact seconds/chars, interrupted streams, stale/missing/ambiguous prices, and estimate-to-final reconciliation.
- **Transport tests:** large input without full memory buffering, slow upload/download, slow WS client, backpressure, cancellation, half-close, idle timeout, close codes, 429, 5xx, network reset, and tunnel loss.
- **Security tests:** multipart and JSON limits, malicious media URL, callback SSRF controls if introduced, oversized WS frame, zip/decode bombs, header injection, cross-tenant request IDs, and logging assertions.
- **Provider contract tests:** scheduled, credential-gated tests against Deepgram's live sandbox/account. Record provider request IDs, not media or transcript content.
- **UI/docs tests:** provider credential wizard, region/tier/MIP settings, model availability, language/voice labels, unsupported-feature messages, curl/OpenAI SDK examples, and RTL layout where UI changes occur.

### Per-model definition of done

| Model | Required green scenarios |
|---|---|
| `nova-3` | File REST, URL REST, verbose/native responses, language selection, timestamps, diarization, keyterm, multichannel billing, provider errors, live WS |
| `flux-general-en` | `/v2/listen` WS, English audio, configure, eager/end/resumed turn events, backpressure, close/retry, exact billing |
| `flux-general-multi` | Same as English plus every declared language/language-hint behavior and multilingual price selection |
| Each Aura-2 voice | REST binary streaming, WS streaming, correct locale/model, format/sample-rate/speed validation, 2,000-character boundary, `dg-char-count`, cancellation |

### Rollout and rollback

- Feature flags must be per endpoint/model and independent for catalog visibility, tenant selection, platform credentials, BYOK, REST, and realtime.
- Start with internal credentials and synthetic non-sensitive audio, then selected tenants, then general availability.
- Rollback disables new routing while allowing in-flight streams to close cleanly. Pricing/catalog snapshots used by completed ledger entries remain immutable.
- Do not advertise a model as available until its endpoint and pricing health checks are green in the tenant's selected region and credential source.

## 13. Risks, open questions, and recommended decisions

### Principal risks

| Risk | Consequence | Mitigation |
|---|---|---|
| Empty/deleted catalog state is imported accidentally | Models disappear or remain archived/unroutable | Resolve scope; add provider completeness validation; review generated diff before import |
| Realtime implemented as an HTTP-stream variant | Flux still cannot work and event semantics are lost | Implement true bidirectional WS through control plane and tunnel |
| Provider response normalized too early | Lost usage, timestamps, speakers, intelligence fields, request IDs | Capture immutable provider metadata first; normalize at the public boundary |
| Pricing falls back to zero or one unconditional rate | Wallet leakage and margin loss | Deepgram fail-closed selection, exact quantities, additive components, immutable snapshots |
| Large files remain in memory | Process exhaustion and tenant denial of service | Lower bounded limit plus streaming/spooling and backpressure |
| Raw callbacks or arbitrary base URLs are enabled | SSRF, data exfiltration, cross-tenant leakage | Defer callbacks; allowlist regions; operator approval for custom egress |
| Provider promotions change | Cost drift | Effective-dated rates, freshness automation, platform-funded circuit breaker |
| Retry after partial stream | Duplicate audio/output and charges | No automatic replay without proven idempotency; explicit terminal events |

### Decisions required before Phase 0 exit

1. Confirm that the six current model files—not the approximately 140 deleted historical entries—are the intended supported catalog.
2. Approve the public realtime path and event envelope. The recommendation is `/v1/realtime?intent=transcription|speech` with normalized events plus explicit provider-native payloads.
3. Decide whether provider-native REST responses are available to all tenants or gated as an advanced compatibility mode.
4. Define the maximum public prerecorded upload size and duration. It should be materially below Deepgram's 2 GB maximum until the full path is streaming and capacity-tested.
5. Decide whether remote URL input is generally available, allowlisted, or enterprise-only.
6. Define pricing ownership for platform Pay As You Go, published Growth, negotiated enterprise, MIP opt-out, markup, taxes, and currency conversion. Published Growth rates must not be assumed for an account.
7. Define the default MIP/retention posture and whether a tenant may override it per request.
8. Decide whether GodarAI-managed async jobs/webhooks are in the full-support definition. Provider-direct callbacks should remain excluded.
9. Define regional routing behavior and whether self-hosted Deepgram is a supported enterprise feature.
10. Establish cost-drift, latency, availability, and realtime concurrency SLOs required for general availability.

### Final recommendation

Proceed with the six-model scope, but do not market Deepgram as supported until Phases 0–3 are complete. Ship Nova-3 prerecorded and Aura-2 REST behind explicit beta flags as independently useful milestones. Flux must remain unavailable until the WebSocket data plane, exact realtime billing, and resiliency controls are production-ready. This sequencing delivers user value without representing a narrow `/v1/listen` adapter as full provider support.

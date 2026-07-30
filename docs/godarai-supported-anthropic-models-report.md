# GodarAI Anthropic model production-readiness audit

**Audit date:** 2026-07-31  
**Evidence retrieval:** 2026-07-31 (Asia/Tehran)  
**Scope:** `claude-sonnet-5`, `claude-haiku-4-5`, `claude-opus-5`,
`claude-fable-5` on the first-party Claude API through GodarAI  
**Final status:** **partially supported**

This is not a production-ready declaration. The catalog and local enforcement
paths are materially improved, but the credential-backed matrix, actual
organization/workspace limits, WIF replay, staging billing reconciliation, and
human release review are `BLOCKED` because the required credentials and
staging endpoint are not available in this environment.

## Executive decision

| Area | Status | Decision |
| --- | --- | --- |
| Exact identities and GodarAI canonical IDs | PASS | The four dateless IDs are current first-party IDs/aliases; GodarAI exposes them as `anthropic/<id>`. The dateless Claude 5 IDs are pinned snapshots, not moving aliases. Haiku's underlying dated ID is `claude-haiku-4-5-20251001`. |
| Native Messages API forwarding | CODE-COMPLETE, LIVE BLOCKED | Unknown provider-native fields and content blocks remain provider-native JSON; runtime-profile validation constrains known incompatible controls. |
| OpenAI-compatible chat | PARTIAL BY DESIGN | Text, images/PDF, custom tools, structured output, thinking, streaming, and the explicitly mapped web/code server tools are supported. Anthropic-only constructs that cannot be represented losslessly are rejected or are native-only. |
| Catalog/runtime profiles | PASS LOCALLY | All four models have explicit immutable version-2 profiles. Claude 5 prefill is disabled; model-specific thinking, server tools, limits, beta headers, retention, ZDR, speed, and 300K Batch metadata are explicit. |
| Pricing | PASS LOCALLY, RECONCILIATION BLOCKED | Effective-dated Sonnet 5 pricing, cache, Batch, US data-residency, Opus 5 fast-mode, and tool dimensions are cataloged. Provider invoice/Cost API reconciliation is not available. |
| Live availability and requested-vs-served identity | BLOCKED | No `ANTHROPIC_API_KEY`, `GODARAI_VIRTUAL_KEY`, WIF configuration, provider account, or staging endpoint was available. |
| Actual rate limits | BLOCKED | No Anthropic Admin API key/workspace ID was available. Published reference values are not treated as the tenant's actual limits. |
| Release | BLOCKED | Production enablement must wait for the gates in the companion plan. |

## Exact model identity and lifecycle

| GodarAI canonical ID | Claude API request ID | Model API identity expectation | Context / max output | Lifecycle note |
| --- | --- | --- | --- | --- |
| `anthropic/claude-sonnet-5` | `claude-sonnet-5` | exact `id=claude-sonnet-5` | 1,000,000 / 128,000 | Active; earliest retirement 2027-06-30 |
| `anthropic/claude-haiku-4-5` | `claude-haiku-4-5` | accepts the alias or its dated identity `claude-haiku-4-5-20251001` | 200,000 / 64,000 | Active; earliest retirement 2026-10-15 |
| `anthropic/claude-opus-5` | `claude-opus-5` | exact `id=claude-opus-5` | 1,000,000 / 128,000 | Active; earliest retirement 2027-07-24 |
| `anthropic/claude-fable-5` | `claude-fable-5` | exact `id=claude-fable-5` | 1,000,000 / 128,000 | Active; earliest retirement 2027-06-09; minimum 30-day retention, not ZDR eligible |

The readiness harness now rejects a Models API response that omits
`max_input_tokens`/`max_tokens` or returns a different identity. OpenAI
non-stream responses now report the upstream `response.model` rather than
blindly repeating the requested model, which is required for Fable fallback
auditability.

## Capability matrix

`Native` means provider-native JSON is preserved. `Compatible` means the
OpenAI-compatible adapter has a semantically lossless mapping. `Reject` means
the compatible route fails closed or the catalog profile rejects the request.
All live conclusions remain blocked until the release matrix is run.

| Capability | Sonnet 5 | Haiku 4.5 | Opus 5 | Fable 5 | Compatible-route policy |
| --- | --- | --- | --- | --- | --- |
| Thinking | Adaptive default; disable allowed; no manual | Manual budget or disabled; no adaptive/effort | Adaptive default; disable only through `high` effort | Adaptive always on; cannot disable | Mapped and profile validated |
| Sampling | Rejected | Allowed without thinking; rejected with thinking | Rejected | Rejected | Profile validated |
| Assistant prefill | Rejected | Allowed | Rejected | Rejected | Profile validated |
| Custom/parallel tools | Native | Native | Native | Native | Lossless mapping |
| Structured output | Native | Native | Native | Native | Lossless mapping |
| Web search | `web_search_20260318` | conservative `web_search_20250305` | `web_search_20260318` | `web_search_20260318` | Explicit mapping |
| Web fetch | `web_fetch_20260318` | Not cataloged | **Unsupported** | `web_fetch_20260318` | Explicit mapping or reject |
| Code execution/programmatic calling | Cataloged | Not cataloged | Cataloged | Cataloged | Explicit code-execution mapping; other programmatic semantics native-only |
| Tool search | Not cataloged pending live proof | Cataloged | Cataloged (`regex`/`bm25` 20251119) | Cataloged | Native-only |
| Memory/advisor/MCP/computer/skills | Cataloged where first-party docs list availability | Cataloged subset | Cataloged | Cataloged | Native-only unless an existing lossless type exists |
| Files | `files-api-2025-04-14` | same | same | same | Native resource API |
| Batch 300K output | `output-300k-2026-03-24` | No | `output-300k-2026-03-24` | No | Native Batch only |
| Fast mode | No | No | `fast-mode-2026-02-01`, `speed=fast` | No | Mapped request/usage/billing |
| Compaction | `compact-2026-01-12` | Not cataloged | cataloged | cataloged | Native-only because compaction blocks must round-trip |
| Mid-conversation system messages | Not supported | Not cataloged | Native | Native | Native-only; compatible conversion cannot promise lossless placement |
| Fable refusal/fallback | N/A | N/A | N/A | `stop_reason=refusal`; optional `server-side-fallback-2026-06-01` | Native-only; served model is surfaced |

The compatibility table is intentionally narrower than Anthropic's API.
Declaring an Anthropic-native feature in a generic `tools.server=true` field is
no longer sufficient: supported server-tool families are checked per model.

## API behavior and controls

- Native `/v1/messages`, `/v1/messages/count_tokens`, Message Batches, Files,
  and Models remain the source-of-truth surfaces.
- OpenAI-compatible `/v1/chat/completions` uses the native Messages API only
  for representable requests. Model-specific unsupported web fetch/search/code
  combinations fail before dispatch.
- Claude 5 sampling and assistant-prefill conflicts now fail locally.
- Opus 5 rejects `thinking.type=disabled` when either
  `reasoning_effort` or `output_config.effort` exceeds `high`.
- Fable 5's runtime profile allows only adaptive thinking and records its
  30-day retention floor and lack of ZDR eligibility.
- A model's `zdrEligible` flag describes the base model, not every feature.
  MCP connector and Agent Skills requests are not ZDR eligible and require
  their own retention-policy gate.
- Streaming reports the upstream model from `message_start`; synchronous
  compatible responses now do the same.
- Provider-native unknown fields are not stripped from the tunnel body.
  Unknown OpenAI-compatible fields are not advertised as supported.

## Pricing and billing matrix

All values are USD per million tokens unless stated otherwise. Effective times
are UTC and end times are exclusive.

| Model / period | Input | Output | 5m write | 1h write | Cache hit | Batch input/output |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sonnet 5, 2026-06-30 through 2026-08-31 | 2 | 10 | 2.5 | 4 | 0.2 | 1 / 5 |
| Sonnet 5, from 2026-09-01 | 3 | 15 | 3.75 | 6 | 0.3 | 1.5 / 7.5 |
| Haiku 4.5 | 1 | 5 | 1.25 | 2 | 0.1 | 0.5 / 2.5 |
| Opus 5 standard | 5 | 25 | 6.25 | 10 | 0.5 | 2.5 / 12.5 |
| Fable 5 | 10 | 50 | 12.5 | 20 | 1 | 5 / 25 |

Additional dimensions:

- US-only inference applies Anthropic's 1.1 multiplier to ordinary, cache,
  Batch, and fast-mode rates.
- Opus 5 fast mode is 10/50; its cache multipliers are applied to the fast
  input price. Fast mode and Batch are mutually exclusive.
- Web search is $10 per 1,000 searches; failed searches are not billable.
- Code execution is $0.05/container-hour with the published 1,550 free
  organization hours/month and five-minute minimum. GodarAI still fails closed
  when a billable request cannot be reconciled from provider cost evidence.
- Web fetch has no additional tool charge; token charges still apply.
- The importer now persists manifest `effective_from`/`effective_to` rather
  than replacing them with import time. Runtime-profile versions are
  immutable: changing a stored profile requires incrementing its version.

Invoice/Cost API reconciliation, rounding parity, failed/cancelled Batch
semantics, and promotional transition replay are `BLOCKED` pending credentials
and staging data.

## Rate-limit and resource behavior

Published reference ceilings are discovery evidence only and must not be used
as the tenant's actual limits. The retrieved first-party standard table
currently groups Claude 5 separately from prior generations and lists:

| Model group | RPM | Input TPM | Output TPM |
| --- | ---: | ---: | ---: |
| Fable 5 | 1,000 | 500,000 | 100,000 |
| Opus 5 | 1,000 | 2,000,000 | 400,000 |
| Sonnet 5 | 1,000 | 2,000,000 | 400,000 |
| Haiku 4.5 | 1,000 | 2,000,000 | 400,000 |

The actual organization and workspace limits are `BLOCKED`. The readiness
command queries the Admin Rate Limits API when given a separate Admin key. Its
report stores only response byte length, resource count, and SHA-256 digest;
it does not persist organization/workspace metadata or credentials.

The command covers sync/stream Messages, Count Tokens, structured output,
thinking, a profile-selected compatible web tool, Files lifecycle, Batch
lifecycle/cancel/results, optional 300K Batch, and Admin rate-limit evidence.
Resource checks are opt-in because they create billable provider resources.

## Requirement-by-requirement audit

| Requirement | Evidence/status | Severity if released now |
| --- | --- | --- |
| Four exact IDs and canonical GodarAI IDs | Implemented and documented | PASS |
| Native plus lossless compatible API | Core paths implemented; advanced provider-only constructs kept native-only | P1 residual |
| Current first-party ground truth with reproducible revision | URLs, retrieval date, and content SHA-256 below | PASS |
| Explicit versioned runtime profiles | Version 2 for all four; immutable import | PASS |
| Exact limits/thinking/prefill/sampling | Encoded and locally enforced | PASS |
| Model-specific tools/versions/betas | Core matrix encoded; full live acceptance still blocked | BLOCKED |
| Sync and streaming model identity | Upstream served model surfaced | PASS LOCALLY |
| Caching/Batch/fast/data-residency/tool pricing | Encoded with effective dates | PASS LOCALLY |
| Actual rate limits | Admin API credentials absent | BLOCKED |
| Files/Batch lifecycle | Harness exists; provider account absent | BLOCKED |
| API-key and WIF auth paths | Existing paths inspected; live replay absent | BLOCKED |
| Retries/timeouts/backpressure/load | Existing tests/harness present; 500 MB gate not run here | BLOCKED |
| Staging deployment/real calls/long context/abuse/observability | No staging authority or credentials | BLOCKED |
| Human review and release approval | Not performed | BLOCKED |
| Recurring source/model/pricing/deprecation refresh | Designed in release plan; no scheduled job deployed | P1 |

## Local verification

| Check | Result |
| --- | --- |
| Model runtime-profile schema tests | PASS (8 tests) |
| Contracts platform tests | PASS |
| Anthropic readiness command tests | PASS |
| Focused control-plane proxy and billing tests | PASS with `GOMAXPROCS=2`, `-parallel 1` |
| Silk Road focused adapter/server tests | PASS with `GOMAXPROCS=2`, `-parallel 1` |
| Model-discovery build and Anthropic YAML dry-run | PASS; 27 manifests parsed |
| Full control-plane package suite | BLOCKED by Windows commit-limit OOM in unrelated parallel cache-heavy tests |
| Credential-backed readiness matrix | BLOCKED: no credentials/staging |
| Admin Rate Limits snapshot | BLOCKED: no Admin key |

No live call was made and no secret or organization metadata was written.

## Source ledger

Anthropic did not return `ETag` or `Last-Modified` headers for these responses.
The revision identifier is therefore the SHA-256 of the UTF-8 HTTP 200 response
body retrieved on 2026-07-31.

| First-party source | SHA-256 |
| --- | --- |
| [Models overview](https://platform.claude.com/docs/en/about-claude/models/overview) | `8ef42d53a29b9f26ce7108e720a7bfc87351da5eae5a538874057e155358f282` |
| [Pricing](https://platform.claude.com/docs/en/about-claude/pricing) | `9796f469bd1338382d708533a621f638e156ba89c62f10caf333ef93f53b615b` |
| [Model deprecations](https://platform.claude.com/docs/en/about-claude/model-deprecations) | `062c685f83fe0a147b852a5291fdb262ec6deaf10d3ad5cd4448bd32a3a0457f` |
| [Migration guide](https://platform.claude.com/docs/en/about-claude/models/migration-guide) | `f118cab8785fbf81ec7fd7753de5daccf9410b815a2523d63ef98b3c498f70bb` |
| [What's new in Sonnet 5](https://platform.claude.com/docs/en/about-claude/models/whats-new-sonnet-5) | `6f5097677357d5c9d5b3938988e6393e1db382cf61bbb89bac9b4cc595802454` |
| [Models API](https://platform.claude.com/docs/en/api/models) | `91ba601fb497063f632766d9d24168e77102ce1b257bf89b98e260ee08a118bd` |
| [Rate limits](https://platform.claude.com/docs/en/api/rate-limits) | `2818f2166367fd6209c41ddd010ef7368a13d876e859d4033e1e701a745381b0` |
| [Service tiers](https://platform.claude.com/docs/en/api/service-tiers) | `7b1da33b06bf42a9c97e0d73cc726b91278ad456db333166a681132751af1e3c` |
| [Fast mode](https://platform.claude.com/docs/en/build-with-claude/fast-mode) | `90635a617fd5eb12c1247634ed9d1d1cacd0bfb44a1c1bcc462132872b3efbce` |
| [API/data retention](https://platform.claude.com/docs/en/manage-claude/api-and-data-retention) | `75674991fb64cd62b0c17a5fc4057c70e14a6a5fce58903007c365741549e004` |
| [Web search](https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-search-tool) | `78236ddb4d1467828ffe1e162b86d46f92ef8ba2657d7829764b043d2dafcbb0` |
| [Web fetch](https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-fetch-tool) | `043d4ecd5633be67635bb077699c6b57ca9949e6cac3327bf1a5bf83a3bbf0d9` |
| [Code execution](https://platform.claude.com/docs/en/agents-and-tools/tool-use/code-execution-tool) | `7d830efbea2c0e3037032719c34ac2e72469c9421a3868357cc73056e9ec3d24` |
| [Tool search](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool) | `e5bc252632a9de7209a9d8faf7b2946f26daadeaa099a9617e017c5dbc6a7481` |
| [Advisor](https://platform.claude.com/docs/en/agents-and-tools/tool-use/advisor-tool) | `705b2188132de0b3c55aa08b9cda4e384729df5bc61e38741d300ebbe70db82e` |
| [Memory](https://platform.claude.com/docs/en/agents-and-tools/tool-use/memory-tool) | `afe0f4c6a361b3fdce758ff1057bb26532a2536ff085a86f4eb0285ee204b61f` |
| [MCP connector](https://platform.claude.com/docs/en/agents-and-tools/mcp-connector) | `5951c38b034f13918bebb73e04c9d7761ce5195b1a958ba83dcf38d688935c5b` |
| [Computer use](https://platform.claude.com/docs/en/agents-and-tools/tool-use/computer-use-tool) | `c70e0769f53bb1b0052cd867c372096cb9e59e1130dc30fa54709d4e5d895206` |
| [Agent Skills](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview) | `d227456656921a3d97b877946602276cd4d78290fd99a9317ea6b046caebb691` |
| [Prompt caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) | `624ae6139ce5137b661bc48ddb10767960da6a5f9d23723ac41c3c156ba7003e` |
| [Batch processing](https://platform.claude.com/docs/en/build-with-claude/batch-processing) | `06202d546dfd034c5b79b71ff5d59afbcb7d77c1f626f72501cb3ec04b2b9ef8` |
| [Token counting](https://platform.claude.com/docs/en/build-with-claude/token-counting) | `51b6705c11ebeec6d8cc031ff57866ac111f6fad1c6f999aa3bea41533fb9f95` |
| [Files](https://platform.claude.com/docs/en/build-with-claude/files) | `c4567f8f519142188ad50454dfc84a2d0e76e78810b93fa5616e13f4023fd1f0` |
| [Fable 5 introduction](https://platform.claude.com/docs/en/about-claude/models/introducing-claude-fable-5-and-claude-mythos-5) | `667315ef59b9f3734f3e6febf9d6ecdaeb38d6a447932b0c486c13c7ee5cd8df` |

## Release conclusion

The implementation is not eligible for the label `production-ready` today.
It is **partially supported**: locally provable catalog, validation, adapter,
identity, and pricing defects are addressed, while release-critical live,
operational, billing-reconciliation, rate-limit, WIF, long-context, and human
approval gates remain blocked.

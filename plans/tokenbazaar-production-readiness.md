# TokenBazaar provider production-readiness plan

Status: **not production-ready**
Last verified against reachable sources: 2026-09-11 (this session, Anthropic Messages endpoint delivery)

## Feature-flag removal (2026-09-11, this delivery) -- read this before anything below

**Decision, dated and deliberate:** all five TokenBazaar kill-switch
environment variables described throughout this document --
`SHAHRAH_TOKENBAZAAR_ENABLED` (chat), `SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`,
`SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED`, `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED`,
and `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED` -- were **removed entirely** from
the codebase in this delivery, along with their `ProviderFeatureSettings`/
`ProviderFeatureConfig` struct fields, their dispatch-time gate checks, and
every test that existed solely to prove one of these flags defaults to
`false` or is isolated from the others. This was an explicit product
decision, not a bug fix or an oversight correction, for these reasons:

- GodarAI's admin-driven provider/model enable-disable mechanism
  (support-admin -> model-discovery catalog's `enabled` flag) is the intended
  control surface for turning providers and models on and off going forward
  -- not a per-provider, per-capability environment variable baked into
  control-plane at build/deploy time. Five independent env-var flags for one
  reseller provider, added across five separate delivery sessions, was not a
  pattern worth propagating to every future provider or capability.
- **Known, accepted trade-off, stated plainly so it is never silently
  "fixed" by a future delivery inventing a substitute gate:** the
  model-discovery catalog's `enabled` flag today only affects catalog
  visibility/selection in the admin UI and tenant-facing model picker -- it
  does **not** affect control-plane's actual dispatch decision.
  `providerregistry` (`services/control-plane/internal/providerregistry/registry.go`)
  is a static, compile-time Go registry with no runtime database read; it has
  no connection today to whatever an operator sets in the catalog UI. This is
  an architectural gap that predates this delivery and applies to every
  provider registered there, not just TokenBazaar.
- Removing these five flags means **TokenBazaar's five capabilities (chat,
  images, embeddings, responses, messages) now have no independent
  code-level kill switch at all**, until the catalog-to-dispatch connection
  above is designed and built as its own, general (all-providers) piece of
  work. That gap is accepted, explicitly, as a known and recorded risk in
  this decision -- not resolved by this delivery, and not to be quietly
  reintroduced as a new ad hoc flag "just in case."
- The trade-off is judged acceptable **right now** for one concrete,
  time-bound reason: **no tenant has a live TokenBazaar provider
  account/credential yet.** Every dispatch path below is fully wired and
  reachable via its explicit `tokenbazaar:<account>:<model-id>` FQN, but
  reaching it in practice still requires a tenant to hold a TokenBazaar
  provider account in the first place -- which today, nobody does. The
  moment that changes, the catalog-to-dispatch connection above stops being
  optional and becomes the actual blocking prerequisite for exposing
  TokenBazaar to real traffic.

**What this means for every flag-related claim below:** every mention in this
document, in sessions dated before 2026-09-11, of `SHAHRAH_TOKENBAZAAR_ENABLED`,
`SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`, `SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED`,
`SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED`, `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED`,
their "default false" state, or a flag being "independent" of another, is a
**historical description of code that existed at the time that session was
written and no longer exists today.** None of it should be read as a
current safety property. TokenBazaar's actual current safety posture is:
all five capabilities are wired and reachable via explicit FQN with **no
independent code-level kill switch**, gated in practice only by (a) no
tenant holding a real TokenBazaar provider account/credential yet, and (b)
the release gates further down this document (live-probe confirmation of
each wire contract, the pricing-anomaly product/finance sign-off) --
which remain the actual, unresolved reasons `production_ready` stays
`false` for every TokenBazaar capability. The "Release gates" section and
the `production_ready` summary later in this document have been updated to
drop the now-nonexistent flags and state this directly; treat those, not the
older per-delivery narrative above them, as the current status.

## Anthropic Messages endpoint delivery (this session, 2026-09-11)

This delivery adds real, tested dispatch for `POST /v1/messages`
(`platform.ProviderOperationAnthropicMessages`) for exactly the 4 Claude
models TokenBazaar carries: `claude-fable-5`, `claude-haiku-4-5`,
`claude-opus-5`, `claude-sonnet-5`. Create-only, mirroring the Responses
delivery's own scope discipline: `count_tokens`, `message_batches`, and
`files` are not implemented for TokenBazaar (nothing in TokenBazaar's docs
suggests it offers them). `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED` (default
`false`) is a new kill switch, independent of the four existing TokenBazaar
flags (chat/images/embeddings/responses, all still `false`).
`production_ready` for the messages capability stays explicitly `false` --
see the release-gate additions at the bottom of this document.

### Routing architecture -- the central decision of this delivery

Before writing any dispatch code, this delivery investigated whether
`AnthropicNativeHandler` (`services/control-plane/internal/proxy/
anthropic_native_handler.go`) is the sole entry point for `POST /v1/messages`,
or whether a more generic tunnel/proxy path (the one `chat_completions`/
`responses`/images/embeddings already use for TokenBazaar) could carry this
operation instead. The answer is unambiguous:
`services/control-plane/internal/server/routes_proxy.go` wires
`v1.Post("/messages", native)` where `native := deps.AnthropicNativeHandler.Handle`
-- there is no other handler registered for this path, and no generic
`ChatCompletionHandler`/`ResponsesHandler`/`EmbeddingsHandler`/`ImagesHandler`-
style handler exists for `platform.ProviderOperationAnthropicMessages`. Two
more findings confirmed the "single-provider-locked" characterization in this
delivery's own executable prompt:

1. `AnthropicNativeHandler` resolves deployments via its own
   `anthropicDeploymentLister`, filtering `strings.EqualFold(deployment.Provider,
   "anthropic")` and applying an `accountAllowlist` scoped to
   `AnthropicAccountIDs` -- a routing model with no FQN concept at all (the
   direct Anthropic provider is never reached via a
   `"anthropic:<account>:<model>"` string the way TokenBazaar's other four
   capabilities require an explicit `"tokenbazaar:<account>:<model>"` FQN).
   It also bundles WIF credential issuance, native file/batch resource
   mappings, and `filesEnabled`/`batchesEnabled` flags that have no
   TokenBazaar equivalent.
2. `services/control-plane/internal/providerregistry/registry.go`'s
   `anthropic` entry does not declare `platform.ProviderOperationAnthropicMessages`
   at all -- only `CapabilityChatCompletions`/`platform.ProviderOperationChatCompletions`.
   This means the executable prompt's own instruction to add
   `CapabilityMessages`/`platform.ProviderOperationAnthropicMessages` to
   TokenBazaar's registry entry "mirroring exactly how anthropic's own
   registry entry declares it" could not be followed literally: there is no
   existing precedent to mirror, because the direct Anthropic provider
   bypasses this registry's FQN-based capability system entirely for its
   native Messages/Batches/Files surface. `CapabilityMessages` was added as a
   **new** capability constant, following the same pattern
   `CapabilityResponses`/`CapabilityEmbeddings` established in the Responses/
   Embeddings deliveries when no prior constant existed for those operations
   either -- not a mirror of an anthropic precedent that turned out not to
   exist.

**Decision: a new, narrow, TokenBazaar-scoped sibling handler
(`proxy.TokenBazaarMessagesHandler`), not an extension of
`AnthropicNativeHandler`'s provider filter.** This handler reuses the
standard FQN -> deployment (`router.Router`) -> providerregistry capability ->
Silk Road adapter flow already used for TokenBazaar's other four
capabilities (mirroring `EmbeddingsHandler`'s non-virtual-model routing
style, the simplest existing precedent for a first delivery of a new
capability), rather than the account-allowlisted, WIF-laden
`anthropicDeploymentLister` path. A thin routing shim,
`proxy.NewAnthropicMessagesRouter(tokenBazaar, native)`, is wired in
`routes_proxy.go` in place of the bare `native` handler for
`v1.Post("/messages", ...)`: it inspects the request body's `"model"` field
(the only signal, matching every other TokenBazaar capability's "explicit FQN
only" rule) and dispatches a `"tokenbazaar:<account>:<model-id>"`-FQN'd
request to `TokenBazaarMessagesHandler`; every other request (including every
direct-Anthropic request, FQN'd or not) falls through to `native`
unconditionally, unchanged.

**`AnthropicNativeHandler`'s dispatch/routing logic, account allowlist, and
WIF/files/batches machinery are completely untouched by this delivery.** The
one exception, made for lint hygiene rather than behavior: once a second,
structurally-identical Anthropic SSE streaming relay existed
(`TokenBazaarMessagesHandler.handleStream`), this repository's `dupl`
linter flagged the duplication against `AnthropicNativeHandler.proxyStream`.
Per this project's own reuse discipline, the shared streaming-relay logic
(dispatch the tunnel stream, forward `content_block_delta`/`message_stop`
chunks raw, track usage via `parseAnthropicSSEUsage`, detect error frames via
`anthropicSSEIsError`, finalize billing on close) was extracted into one new
package-level function, `streamAnthropicShapedResponse`
(`anthropic_native_handler.go`), which both handlers now call.
`AnthropicNativeHandler.proxyStream` is now a one-line wrapper around this
function with byte-for-byte identical behavior to before the refactor
(confirmed by the full pre-existing `anthropic_native_handler_test.go` suite
passing unchanged) -- this is a pure extract-method refactor, not a change to
`AnthropicNativeHandler`'s account allowlist, WIF, files/batches gating, or
any dispatch decision.

This satisfies the executable prompt's own preference ordering: a parallel
path was feasible (unlike, say, widening `AnthropicNativeHandler`'s account
allowlist itself would have required to serve TokenBazaar's non-Anthropic-
account deployments), so this delivery did not need to fall back to
"surgically extending `AnthropicNativeHandler`'s provider check."

### Reuse-vs-new-code accounting

- **Reused unmodified:**
  - `services/silk-road/pkg/server/openai_compatible.go`'s
    `openAICompatibleAdapter.Dispatch` -- already generic raw-byte-forwarding
    HTTP construction (method/URL/auth-header/Content-Type/Accept-for-streaming),
    not actually OpenAI-shape-specific despite its name. TokenBazaar's
    `/v1/messages` request body is already Anthropic-native JSON by the time
    it reaches this adapter (built and forwarded unchanged by
    `TokenBazaarMessagesHandler`, mirroring how `AnthropicNativeHandler`
    forwards a direct Anthropic dispatch unchanged) -- so no OpenAI<->Anthropic
    conversion was ever needed here, and this adapter required **zero**
    changes.
  - `services/silk-road/pkg/server/tokenbazaar.go`'s `tokenBazaarAdapter.Dispatch`,
    `validateTokenBazaarBaseURL` (host allowlist), and `normalizeTokenBazaarResponse`
    (429-passthrough/4xx-normalization, response-size ceiling) -- reused
    exactly as-is, same as every prior TokenBazaar capability.
  - `services/control-plane/internal/proxy/anthropic_native_request.go`'s
    `anthropicMaximumOutputTokens` (billing-estimate max-tokens extraction) and
    `anthropic_native_handler.go`'s `parseAnthropicNativeUsage`,
    `parseAnthropicSSEUsage`, `anthropicSSEIsError`, `writeAnthropicError`,
    `writeAnthropicBillingError`, `closeStreamBestEffort`,
    `writeTunnelStreamHTTPError` -- all already provider-agnostic,
    package-level functions; none needed a TokenBazaar-specific branch.
  - `services/control-plane/internal/proxy/handler.go`'s
    `runtimeProfilesFromEstimate`, `rewriteOpenAIJSONModel`,
    `toContractDeployment`, `applyDeploymentPricingOverride`,
    `deploymentBillingMode`, `providerModelID`, `billingEstimateModel` --
    the same generic routing/billing helpers `EmbeddingsHandler`/
    `ImagesHandler`/`ResponsesHandler` already use.
  - `services/control-plane/internal/billing/provider_pricing_strategy.go`'s
    `tokenBazaarPricingStrategy` -- its embedded `standardPricingStrategy`
    already bills any usage (`InputTokens`/`OutputTokens`/`TotalTokens`, the
    same `contractsopenai.Usage` shape `parseAnthropicNativeUsage` already
    maps Anthropic's `input_tokens`/`output_tokens` into) via the generic
    exact-decimal `costBreakdown`, keyed by model + token counts, not
    operation type. **No `FinalizeUsage` override was needed or added.** The
    4 in-scope models' catalog YAMLs
    (`apps/model-discovery/providers/tokenbazaar/{claude-fable-5,
    claude-haiku-4-5,claude-opus-5,claude-sonnet-5}.yaml`) already declared
    `runtimeProfile.endpoints` including `messages` **before this delivery**
    (they were pre-scaffolded with the full Anthropic profile) -- confirmed
    by inspection, no YAML change was needed for any of the 4 models.
- **New code written:**
  - `services/control-plane/internal/proxy/tokenbazaar_messages_handler.go`
    (new): `TokenBazaarMessagesHandler` (mirroring `EmbeddingsHandler`'s
    non-virtual-model routing/billing/logging structure), its
    `Handle`/`handleSync`/`handleStream` methods, `requestTargetsTokenBazaarMessages`
    and `NewAnthropicMessagesRouter` (the routing shim), and
    `supportsAnthropicMessagesProvider` (a fail-closed defense-in-depth check
    that only "tokenbazaar" may reach this generic path -- the direct
    "anthropic" provider is structurally never routed here in the first
    place, since it is never reached via a bare-model FQN).
  - `services/control-plane/internal/proxy/anthropic_native_handler.go`: the
    `streamAnthropicShapedResponse` extraction described above (behavior-
    preserving refactor, not new logic).
  - `services/silk-road/pkg/server/tokenbazaar.go`: `tokenBazaarMessagesModels`
    (the 4-model allowlist), a new `case platform.ProviderOperationAnthropicMessages`
    in `validateTokenBazaarDispatch`'s switch, and
    `validateTokenBazaarMessagesDispatch` plus its helpers
    (`validateTokenBazaarMessagesMaxTokens`, `validateTokenBazaarMessagesToolsUnsupported`,
    `validateTokenBazaarMessagesPromptCachingUnsupported`,
    `tokenBazaarBlocksContainCacheControl`, `tokenBazaarPromptCachingUnsupportedError`)
    -- these implement the fail-closed field-level validation described below.
  - `services/silk-road/pkg/server/openai_operations.go`: one new `case
    platform.ProviderOperationAnthropicMessages: return http.MethodPost,
    "/messages", nil` in `resolveCoreOpenAIOperation` -- a pure method/path
    mapping, exactly like every other case in this function; no request-body
    transformation happens here (the body is already Anthropic-native JSON
    by construction). Only TokenBazaar's adapter is ever dispatched this
    operation via this shared path today.
  - `services/control-plane/internal/config/provider_features.go` +
    `internal/proxy/dependencies.go` + `internal/bootstrap/build_gateway.go`:
    added `TokenBazaarMessagesEnabled` / `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED`,
    wired end to end exactly like the four existing TokenBazaar flags.
  - `services/control-plane/internal/server/server.go` +
    `internal/server/routes_proxy.go`: added `Dependencies.TokenBazaarMessagesHandler`
    and wired `NewAnthropicMessagesRouter` in place of the bare `native`
    handler for `POST /v1/messages` (the only line-level change to how this
    route is registered; every other `/v1/messages*` sub-route -- count_tokens,
    batches -- is untouched, still routed straight to `native`, since Messages
    create-only is this delivery's explicit scope).
  - `services/control-plane/internal/providerregistry/registry.go`: added
    `CapabilityMessages` (new constant, added to `validCapability`'s
    allowlist switch too) and `platform.ProviderOperationAnthropicMessages`
    to TokenBazaar's registered capabilities/operations.
    `TestTokenBazaarSupportsMessagesCapability` and
    `TestTokenBazaarMessagesModelsNeverCollideWithDirectProviders` added
    (mirroring the existing responses/images/embeddings collision tests);
    `TestDefinitionsSatisfyProviderContracts`'s hardcoded tokenbazaar contract
    updated.
  - `packages/contracts/platform/tenant_provider_account.go`: added
    `ModelTypeMessages` -- a new `platform.ModelType` constant, mirroring how
    `ModelTypeResponses` already exists as its own capability-scoped access
    type distinct from `ModelTypeChat`, even though the underlying models are
    shared with chat. `TokenBazaarMessagesHandler` checks
    `modelAccessForType(authInfo, requestedModel, platform.ModelTypeMessages)`,
    matching `ResponsesHandler`'s own use of `platform.ModelTypeResponses`
    rather than `platform.ModelTypeChat`. `services/control-plane/internal/admin/provider_catalog.go`'s
    `mapModeToModelType` (the tenant-admin-configured "mode" string -> `ModelType`
    mapping) gained a `"messages"` case, mirroring its existing `"responses"`
    case, so operators can actually grant this access type through the
    existing admin configuration surface.

### Tool-use and prompt-caching decision

TokenBazaar's documented `/v1/messages` example does not show a `tools`
parameter (unlike its `/v1/chat/completions` section, which explicitly
documents `tools`) and does not show `cache_control` prompt-caching fields.
Per this delivery's own fail-closed discipline (and the executable prompt's
explicit preference), **both are rejected before dispatch, each with its own
distinguishable error, rather than silently passed through:**

- Any non-empty `tools` array ->
  `tokenbazaar_messages_tools_unsupported`. An absent, `null`, or empty
  `tools` array is allowed through unchanged.
- A `cache_control` block on a system block or a message content block ->
  `tokenbazaar_messages_prompt_caching_unsupported`. (The Anthropic schema's
  third `cache_control` location, a tool definition, is not checked as a
  separate case: since any non-empty `tools` array is already rejected
  outright by the tools check above, a request with `cache_control` on a
  tool definition is already rejected by that check before this one ever
  runs -- scanning tools for `cache_control` here would be unreachable dead
  code, and was deliberately not written.)

`thinking: {type: "enabled", budget_tokens: N}` is documented as supported on
this exact route and is **not** validated or stripped anywhere in this
delivery: since the request body is forwarded to TokenBazaar byte-for-byte
(aside from the model-id rewrite), it passes through unmodified by
construction -- confirmed by
`TestTokenBazaarMessagesThinkingPassesThroughUnmodified` (Silk Road, all 4
models, fixture-based) and
`TestTokenBazaarMessagesHandlerThinkingPassesThroughUnmodified` (control
plane, end-to-end through the handler).

Multimodal image input (`{"type":"image","source":{"type":"base64",...}}`) is
documented on this route but is **not explicitly gated or validated by this
delivery** -- it is neither accepted-and-tested nor explicitly rejected; it
simply passes through as part of the forwarded message content, the same as
any other content-block shape this delivery does not specifically inspect.
This is a deliberate, narrower scope than a "supported and tested" claim
would require: **`FOLLOW-UP`** -- a future delivery should either add
fixture-based coverage confirming image content blocks dispatch and bill
correctly, or add an explicit gate if a reason to block them is found.

### Streaming evidence-tier caveat (same discipline as the Responses delivery)

TokenBazaar's `/v1/messages` streaming (`content_block_delta`, `message_stop`,
etc.) is documented (the "Confirmed wire contract" section above already
states TokenBazaar mirrors Anthropic's own `content_block_delta`/
`message_stop` SSE shape) but has **never been independently live-verified
end to end** by this or any prior session -- the same evidence tier as
`/v1/chat/completions` and `/v1/responses` streaming. This delivery
implements against the documented shape only:

- The Silk Road adapter passes the raw SSE bytes through unmodified for
  `/v1/messages`, exactly as it already does for `/v1/chat/completions` and
  `/v1/responses` -- confirmed by
  `TestTokenBazaarMessagesSSEStreamingPassesThroughUntilTerminalEvent` and
  `TestTokenBazaarMessagesMalformedSSEChunkPassesThroughWithoutError`.
- Typed-event parsing (usage extraction via `parseAnthropicSSEUsage`, error-frame
  detection via `anthropicSSEIsError`) is the shared, provider-agnostic
  responsibility of `streamAnthropicShapedResponse` (see above) -- exercised
  against malformed/unexpected events by the new
  `TestParseAnthropicSSEUsageHandlesMalformedAndUnexpectedEvents` (control
  plane), which is itself a **new** test filling a gap: no prior delivery had
  added unit coverage for `parseAnthropicSSEUsage`/`anthropicSSEIsError`
  against malformed input at the control-plane layer, even though
  `AnthropicNativeHandler` has relied on them since the direct Anthropic
  delivery. This is now closed for both the direct Anthropic and TokenBazaar
  Messages paths, since they share the exact same function.
- `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED` defaults `false` and stays off until
  the release gates below are met.

### Fail-closed validation summary (before any HTTP call)

- `max_tokens` absent, `null`, non-integer, zero, or negative ->
  `tokenbazaar_messages_max_tokens_required` (or `tokenbazaar_invalid_json`
  for a wrong JSON type) -- TokenBazaar's own docs mark it required, matching
  Claude's own requirement; this delivery does not rely on upstream to reject
  its absence.
- `messages` absent -> `tokenbazaar_messages_required`.
- Non-empty `tools` -> `tokenbazaar_messages_tools_unsupported`.
- `cache_control` on a system block or message content block ->
  `tokenbazaar_messages_prompt_caching_unsupported`.
- Model not one of the 4 in-scope models -> generic
  `tokenbazaar_unsupported_model` (the other 22 TokenBazaar models).
- Runtime profile missing the `messages` endpoint/transport ->
  `tokenbazaar_runtime_profile_missing` / `tokenbazaar_unsupported_transport`.

`TestTokenBazaarMessagesThreeRejectionsHappenBeforeAnyHTTPCall` proves the
max_tokens/tools/prompt-caching rejections all happen before any upstream
call via an end-to-end mocked-HTTP `Dispatch`, asserting the upstream call
counter stays at exactly 1 (the one valid control request) after all three
rejected variants are attempted.

### Pricing

Reuses the existing chat-completions billing path unmodified, as described
above -- no second pricing computation path was introduced. The 4 in-scope
models' catalog costs were already pinned (shared with the chat capability's
21-model catalog); no pricing anomaly investigation was needed since these 4
Claude models were not among the 3 previously-flagged anomalies (DeepSeek,
Gemini 3.6/3.7 Flash, GPT-5.6 family).

Per this delivery's own mandatory pre-implementation gate, base-URL
reachability was re-checked once more from this session's own execution
environment before any code was written:

| Probe | Result |
|---|---|
| `curl -sS -m 8 https://zx1.deepwl.net/v1/models` | `curl: (28) Resolving timed out after 8015 milliseconds` |
| `nslookup zx1.deepwl.net` | `DNS request timed out` (`Server: UnKnown`) |

Same DNS-timeout failure signature as every prior session's attempt -- no new
evidence the host is wrong, and no third host was invented. `https://zx1.deepwl.net`
remains the authoritative value baked into `validateTokenBazaarBaseURL`,
which the Messages dispatch path reuses unmodified (it targets
`https://zx1.deepwl.net/v1/messages`, confirmed by
`TestTokenBazaarMessagesDispatchExactRequestConstruction`).

## Responses endpoint delivery (this session, 2026-09-11)

This delivery adds real, tested dispatch for `POST /v1/responses`
(`platform.ProviderOperationOpenAIResponsesCreate`) for exactly 6 of the 26
TokenBazaar models: `gpt-5.5`, `gpt-5.6-luna`, `gpt-5.6-sol`, `gpt-5.6-terra`
(OpenAI family) and `deepseek-v4-flash`, `deepseek-v4-pro` (DeepSeek family).
This is a deliberately narrower scope than the 21 chat-capable models, per
this delivery's own executable prompt -- the remaining 15 chat-capable models
(Claude/Gemini/GLM/Kimi) do not gain this operation and continue to reject a
`/v1/responses` request with the same generic `tokenbazaar_unsupported_model`
error the 4 image models and the 1 embeddings model already carry for this
operation. `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED` (default `false`) is a new
kill switch, independent of the existing chat/images/embeddings flags (all
also still `false`). `production_ready` for the responses capability stays
explicitly `false` -- see the release-gate additions at the bottom of this
document.

**This delivery's central discipline is reuse, not new construction** -- per
its own executable prompt, its success is measured partly by how small the
"genuinely new code" category is relative to "existing code extended by one
allowlist entry." A precise accounting:

- **Reused unmodified, only extended with a `tokenbazaar` case/entry:**
  - `services/control-plane/internal/proxy/responses_handler.go`'s
    `supportsResponsesCreateProvider` -- a 4-line `switch` -- gained one case
    (`"tokenbazaar"`). `supportsResponseResourcesProvider` in the same file
    was **deliberately left untouched** (`"openai", "xai"` only): that switch
    gates the stateful GET/DELETE/CANCEL/InputItems/Compact/InputTokens
    surface, which TokenBazaar's own docs do not offer ("previous_response_id
    ... not supported" already rules out retrieval-by-ID as a concept for
    this provider).
  - `services/silk-road/pkg/server/openai_operations.go`'s
    `resolveOpenAIOperation` -- already resolved
    `platform.ProviderOperationOpenAIResponsesCreate` to `POST /responses`
    before this delivery (shared infrastructure used by every OpenAI-compatible
    adapter, including TokenBazaar's own existing chat path). **No path
    mapping was added or changed here.**
  - `services/control-plane/internal/billing/provider_pricing_strategy.go`'s
    `tokenBazaarPricingStrategy` -- its embedded `standardPricingStrategy`
    already bills `/v1/responses` usage (`input_tokens`/`output_tokens`/
    `total_tokens`) via the same generic exact-decimal `costBreakdown` used
    for chat completions and embeddings, because it keys off model + token
    counts, not operation type. **No `FinalizeUsage` override was needed or
    added** -- confirmed by inspection and by the fact that the 6 in-scope
    models' catalog YAMLs (`apps/model-discovery/providers/tokenbazaar/{gpt-5.5,
    gpt-5.6-luna,gpt-5.6-sol,gpt-5.6-terra,deepseek-v4-flash,deepseek-v4-pro}.yaml`)
    already declare a `responses` runtime-profile endpoint alongside
    `chat_completions`, sharing the same `costs` block.
  - The 6 in-scope models' catalog YAMLs already declared
    `runtimeProfile.endpoints: [..., responses, ...]` (OpenAI models) or
    `endpoints: [models, chat_completions, responses]` (DeepSeek models)
    **before this delivery** -- no YAML change was needed for any of the 6
    models.
- **New code written (deliberately kept small):**
  - `services/silk-road/pkg/server/tokenbazaar.go`: `tokenBazaarResponsesModels`
    (the 6-model allowlist, mirroring the existing
    `tokenBazaarImageModels`/`tokenBazaarEmbeddingModels` style), a new `case
    platform.ProviderOperationOpenAIResponsesCreate` in
    `validateTokenBazaarDispatch`'s switch, and
    `validateTokenBazaarResponsesDispatch` plus its 4 small helpers
    (`validateTokenBazaarResponsesPreviousResponseID`,
    `validateTokenBazaarResponsesBackground`,
    `validateTokenBazaarResponsesTools`,
    `validateTokenBazaarResponsesInputPresence`, and the shared
    `tokenBazaarJSONIsNull`) -- these implement the fail-closed rejection of
    the three documented unsupported fields plus the same model-scoping/
    runtime-profile/transport enforcement the chat/images/embeddings paths
    already use. No new HTTP-construction, host-allowlist, retry, or
    response-normalization code was needed: the existing
    `openAICompatibleAdapter.Dispatch`, `validateTokenBazaarBaseURL`, and
    `normalizeTokenBazaarResponse` are reused exactly as-is.
  - `services/control-plane/internal/config/provider_features.go`,
    `internal/proxy/dependencies.go`, `internal/bootstrap/build_gateway.go`:
    added `TokenBazaarResponsesEnabled` /
    `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED`, wired end to end exactly like the
    three existing TokenBazaar flags.
  - `services/control-plane/internal/proxy/responses_handler.go`: **a
    genuinely new structural gap was found and fixed here** (see "Assumption
    found in `responses_handler.go`" below) -- `ResponsesHandler` had no
    `providerFeatures ProviderFeatureConfig` field at all (unlike
    `ChatCompletionHandler`/`ImagesHandler`/`EmbeddingsHandler`, which all
    already carry one), even though `BaseHandlerDependencies.ProviderFeatures`
    was already being passed into `NewResponsesHandlerWithDependencies` and
    silently discarded. This delivery adds the field, wires it in the
    constructor, and adds `rejectDisabledDirectResponsesProvider` (extracted
    into its own method to keep `Handle`'s cyclomatic complexity under the
    project's `gocyclo` threshold), called for non-virtual dispatch, plus an
    equivalent inline check in `estimateVirtualResponseTarget` for
    virtual-model-routed TokenBazaar targets (defense in depth: virtual models
    can in principle route to a TokenBazaar deployment for these 6 models).
  - `services/control-plane/internal/providerregistry/registry.go`: added
    `CapabilityResponses` and `platform.ProviderOperationOpenAIResponsesCreate`
    to TokenBazaar's registered capabilities/operations list.
    `registry_test.go` extended with `TestTokenBazaarSupportsResponsesCapability`
    and `TestTokenBazaarResponsesModelsNeverCollideWithDirectProviders`
    (mirroring the existing images/embeddings collision tests), and the
    exhaustive `TestDefinitionsSatisfyProviderContracts` contract map updated.

**Assumption found in `responses_handler.go` (reported per this delivery's own
instruction, not silently worked around):** two provider-specific assumptions
exist in this file, neither of which needed a TokenBazaar-specific branch, but
both are worth naming explicitly:
1. `prepareDirectResponseRuntimeProfiles` (and its virtual-model twin,
   `prepareVirtualResponseRequest`) special-cases `isDeepSeekProvider(...)` to
   run `validateDeepSeekResponsesRuntimeProfile` -- a deep, DeepSeek-specific
   per-parameter allowlist check (`deepseek_runtime_profile.go`) that
   TokenBazaar does not get, because TokenBazaar is not DeepSeek. This is
   correct and expected: TokenBazaar's own parameter-level validation
   (`validateTokenBazaarResponsesDispatch` in Silk Road) is a materially
   lighter check than DeepSeek's own direct-dispatch validator. This is not a
   bug to fix in this delivery -- DeepSeek's deep validation exists because
   DeepSeek is dispatched to *directly* (its own first-party API), whereas
   TokenBazaar is a reseller whose own docs already promise OpenAI-Responses
   compatibility; but it does mean TokenBazaar Responses requests get less
   control-plane-level parameter scrutiny than DeepSeek's own direct requests
   do. Future TokenBazaar deliveries should not assume this gap is closed
   without deciding whether TokenBazaar needs its own analogous
   `validateTokenBazaarResponsesRuntimeProfile` at the control-plane layer (a
   deliberate scoping decision, not an oversight of this delivery).
2. `resolveResponseCreateDeployment`'s `previous_response_id` continuation
   machinery (`responses_operations_handler.go`) resolves a mapping via
   `responseResourceMapping`, which only ever looks up `"openai"`/`"xai"`
   provider rows. Since `recordResponseMapping` also gates on
   `supportsResponseResourcesProvider` (which TokenBazaar is deliberately
   excluded from), no mapping is ever recorded for a TokenBazaar response, so
   a caller-supplied `previous_response_id` targeting a TokenBazaar deployment
   would never be found, would not be rewritten, and would reach the Silk Road
   adapter completely unvalidated by this generic machinery. This is exactly
   why the fail-closed rejection of `previous_response_id` had to be
   implemented at the TokenBazaar adapter level in Silk Road
   (`validateTokenBazaarResponsesPreviousResponseID`) rather than relied upon
   to be caught generically -- the generic path does not, and structurally
   cannot without further changes, catch it for a provider outside the
   openai/xai resource-mapping allowlist.

**Streaming evidence-tier caveat (do not claim a higher confidence than
chat_completions streaming):** per conflict-register item #4, TokenBazaar's
`/v1/responses` streaming (typed events such as `response.output_text.delta`,
terminating in `response.completed`) is documented with a worked example but
has **never been independently live-verified end to end** by this or any
prior session -- the same evidence tier as `/v1/chat/completions` streaming,
which also isn't live-verified but at least has more redundant documentation
(a full SSE example plus the `data: [DONE]` terminator convention shared with
every other OpenAI-compatible chat endpoint). This delivery implements
against the documented shape only:
- The Silk Road adapter (`tokenBazaarAdapter.Dispatch`) passes the raw SSE
  bytes through unmodified for `/v1/responses`, exactly as it already does for
  `/v1/chat/completions` -- confirmed by
  `TestTokenBazaarResponsesSSEStreamingPassesThroughUntilTerminalEvent`.
- Typed-event parsing, terminal-event detection, and malformed/unexpected-event
  handling are **not** reimplemented per-provider: they are the existing,
  provider-agnostic responsibility of
  `services/control-plane/internal/proxy/stream_relay.go`'s
  `openAIResponseStreamState`, already exhaustively unit-tested independent of
  provider in `stream_relay_test.go` (fragmented frames, oversized terminal
  frames, every SSE line-ending convention, every terminal event type). Once
  `tokenbazaar` was added to `supportsResponsesCreateProvider`, this machinery
  applies to TokenBazaar for free -- no TokenBazaar-specific streaming
  parser was written or was needed.
- `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED` defaults `false` and stays off until
  the release gates below are met, same discipline as every other TokenBazaar
  capability.

**`reasoning.effort` passthrough -- verified with a test, not assumed:** per
the readiness doc, `reasoning.effort` is documented only on `/v1/responses`,
never on chat completions. Since `responses_handler.go`'s request/response
translation is provider-agnostic and this file's only provider-specific
branches are the two named above (neither of which touches `reasoning`),
`reasoning.effort` was expected to pass through for free once TokenBazaar was
in the allowlist. This is now confirmed, not assumed, by two tests:
`TestTokenBazaarResponsesReasoningEffortPassesThroughUnmodified` (Silk Road,
all 6 models, fixture-based) and
`TestResponsesHandlerSupportsTokenBazaarWithReasoningEffortPassthrough`
(control plane, mirroring the existing xAI coverage) -- both confirm the exact
byte sequence `"reasoning":{"effort":"low"}` reaches the upstream request body
unchanged.

**The three documented "not supported" fields are rejected fail-closed,
before any HTTP call, each with its own distinguishable error:**
- `previous_response_id` (non-empty) →
  `tokenbazaar_responses_previous_response_id_unsupported`
- Any built-in/hosted tool type in `tools` (any `type` other than
  `"function"` -- covers `web_search`, `file_search`, `image_generation`,
  `code_interpreter`, and any other current or future OpenAI-hosted tool
  type) → `tokenbazaar_responses_builtin_tool_unsupported`
- `background: true` → `tokenbazaar_responses_background_unsupported`

`TestTokenBazaarResponsesThreeRejectionsHappenBeforeAnyHTTPCall` proves all
three via an end-to-end mocked-HTTP `Dispatch` call (not just the standalone
validator), asserting the upstream call counter stays at exactly 1 (only the
one valid control request) after all three rejected variants are attempted.

**Pricing**: reuses the existing chat-completions billing path unmodified, as
described above -- no second pricing computation path was introduced.

Per this delivery's own mandatory pre-implementation gate, base-URL
reachability was re-checked once more from this session's own execution
environment before any code was written:

| Probe | Result |
|---|---|
| `curl -sS -m 8 https://zx1.deepwl.net/v1/models` | `curl: (28) Resolving timed out after 8007 milliseconds` |
| `nslookup zx1.deepwl.net` | `DNS request timed out` (`Server: UnKnown`) |

Same DNS-timeout failure signature as every prior session's attempt -- no new
evidence the host is wrong, and no third host was invented. `https://zx1.deepwl.net`
remains the authoritative value baked into `validateTokenBazaarBaseURL`, which
the Responses dispatch path reuses unmodified (it targets
`https://zx1.deepwl.net/v1/responses`, confirmed by
`TestTokenBazaarResponsesDispatchExactRequestConstruction`).

## Embeddings endpoint delivery (this session, 2026-09-11)

This delivery removes `text-embedding-3-large` — the last of the originally-5
out-of-scope TokenBazaar models — from `tokenBazaarUnsupportedModels` in
`services/silk-road/pkg/server/tokenbazaar.go`, and implements real, tested
dispatch for it via `POST /v1/embeddings` (`tokenBazaarEmbeddingModels`,
`validateTokenBazaarEmbeddingsDispatch`, `validateTokenBazaarEmbeddingsResponseShape`).
`SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED` (default `false`) is a new kill
switch, independent of both the chat flag (`SHAHRAH_TOKENBAZAAR_ENABLED`) and
the images flag (`SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`), both also still
`false`. `production_ready` for the embeddings capability stays explicitly
`false` — see the release-gate additions at the bottom of this document.

**Evidence tier for this delivery is lower than chat and images, and this is
deliberate, not an oversight.** Unlike `/v1/chat/completions` and
`/v1/images/generations`/`edits` (both of which had a full worked
request/response example on TokenBazaar's own rendered `/docs` page), the
embeddings contract has **no worked example anywhere on that page** — this
session re-read the entire rendered page again, including its nav sidebar
(Quick Start … API Reference), and confirmed no "Embeddings" section exists.
No stronger first-party confirmation (a missed docs section, an OpenAPI/schema
document, a changelog entry) was found before implementation; a fuller
targeted re-read of `tokenbazaar.ai`'s `/docs` page and a search for a public
OpenAPI/schema document both turned up nothing beyond what conflict #11
(below) already recorded. Two facts raise this from "totally unknown" to
"strongly implied, not first-party confirmed":

1. The docs' own product description explicitly claims embedding support:
   "TokenBazaar is a unified AI API service with OpenAI-style compatibility.
   It supports text, image, video, and embedding capabilities across multiple
   models."
2. The authenticated `GET /v1/models` catalog lists `text-embedding-3-large`
   (and `text-embedding-ada-002`, `qwen3-embedding-8b`) with
   `supported_endpoint_types: embeddings`.

Given every other TokenBazaar capability (chat, responses, messages, images,
video) follows an exact OpenAI/Anthropic-compatible path and shape with no
deviation, this delivery implements against the inferred contract
(`POST /v1/embeddings`, standard OpenAI request `{model, input,
encoding_format, dimensions}` / response `{object:"list", data:[{object:
"embedding", embedding:[...], index}], model, usage:{prompt_tokens,
total_tokens}}`), but treats it explicitly as a lower evidence tier:

- The feature flag defaults false and is independent of chat/images (same
  "one model's readiness never unlocks another" rule).
- A successful (< 400) response is shape-checked in the adapter itself
  (`validateTokenBazaarEmbeddingsResponseShape`) before being handed back to
  the caller — a response that does not match `{object:"list", data:[...]}`
  with a non-empty embedding vector per entry is rejected with a distinct
  `tokenbazaar_embeddings_response_shape_unexpected` error rather than
  forwarded blind or causing a panic downstream, precisely because this
  contract is inferred and not first-party-confirmed.
- `production_ready` for embeddings stays `false` until a live, approved
  credential probe against `https://zx1.deepwl.net/v1/embeddings` confirms
  the actual path and response envelope match what is implemented here — see
  release gates below.

Pricing reuses the already-pinned official OpenAI cost in
`apps/model-discovery/providers/tokenbazaar/text-embedding-3-large.yaml`
(`input_cost_per_token: 1.3e-7`, `output_cost_per_token: 0`, from the
2026-09-10 pricing-override delivery) as-is; `tokenBazaarPricingStrategy`
needed no `FinalizeUsage` override to cover this — the embedded
`standardPricingStrategy`'s generic exact-decimal `costBreakdown` already
covers it, the same as every other TokenBazaar chat model's non-cached usage
billing (see `provider_pricing_strategy.go`'s updated comment and
`tokenbazaar_embeddings_pricing_test.go`).

Per this delivery's own mandatory pre-implementation gate, base-URL
reachability was re-checked once more from this session's own execution
environment before any code was written:

| Probe | Result |
|---|---|
| `curl -sS -m 8 https://zx1.deepwl.net/v1/models` | `curl: (28) Resolving timed out after 8003 milliseconds` |
| `nslookup zx1.deepwl.net` | `DNS request timed out` (`Server: UnKnown`) |

Same DNS-timeout failure signature as every prior session's attempt — no new
evidence the host is wrong, and no third host was invented. `https://zx1.deepwl.net`
remains the authoritative value baked into `validateTokenBazaarBaseURL`,
which the embeddings dispatch path reuses unmodified.

## Image endpoints delivery (this session, 2026-09-10)

This delivery implements real, tested dispatch for the 4 TokenBazaar image
models (`gpt-image-2`, `gemini-3-pro-image`, `gemini-3.1-flash-image`,
`gemini-3.1-flash-lite-image`) via `POST /v1/images/generations` and
`POST /v1/images/edits`, replacing the hard `tokenbazaar_model_not_yet_supported`
rejection those 4 models previously carried. `text-embedding-3-large` is
untouched and keeps rejecting exactly as before (conflict #11 remains open —
unrelated to this delivery). See "Implemented and tested in the Image
Endpoints delivery" and the updated conflict #12 entry below for full detail.
`SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED` (default `false`) is a new kill switch,
independent of the existing chat flag (`SHAHRAH_TOKENBAZAAR_ENABLED`, also
still `false`). `production_ready` for the image capability stays explicitly
`false` — code implemented and unit-tested is not the same claim as
production-ready; see the release-gate additions at the bottom of this
document.

Per this delivery's own mandatory pre-implementation gate, base-URL
reachability was re-checked once more from this session's own execution
environment before any code was written:

| Probe | Result |
|---|---|
| `curl -sS -m 8 https://zx1.deepwl.net/v1/models` | `curl: (28) Resolving timed out after 8009 milliseconds` |
| `nslookup zx1.deepwl.net` | `DNS request timed out` (`Server: UnKnown`) |

Same DNS-timeout failure signature as every prior session's attempt — no new
evidence the host is wrong, and no third host was invented. `https://zx1.deepwl.net`
remains the authoritative value baked into `validateTokenBazaarBaseURL`,
which the image dispatch path reuses unmodified (see "Confirmed base-URL
caveat is unchanged" below).

## Base-URL reachability re-check (Phases 2-5 delivery, 2026-09-10)

Per the executable prompt's mandatory pre-implementation gate, this session independently
re-attempted to resolve/connect to `https://zx1.deepwl.net` from its own execution
environment, using three separate tools:

| Probe | Result |
|---|---|
| `nslookup zx1.deepwl.net` | `DNS request timed out` (4 retries, `Server: UnKnown`) |
| `curl -sS -m 10 https://zx1.deepwl.net/` | `curl: (28) Resolving timed out after 10007 milliseconds` |
| `curl -sS -m 10 https://zx1.deepwl.net/v1/models` | `curl: (6) Could not resolve host: zx1.deepwl.net` |

**Case 3 from the executable prompt applies: this environment also cannot resolve
`zx1.deepwl.net`.** This is the identical failure signature already recorded above for
`docs.deepwl.cn` and for the prior session's own `zx1.deepwl.net` attempts (DNS
timeout/`NXDOMAIN`-class, not a TLS/HTTP-level rejection) — consistent with a
regional/sandbox DNS-filtering artifact rather than new evidence the host is wrong.

Per direct user instruction, `https://zx1.deepwl.net` is still used as the configured
value baked into `services/silk-road/pkg/server/tokenbazaar.go`'s host allowlist
(`validateTokenBazaarBaseURL`) — `api.tokenbazaar.ai` was NOT substituted and no third
host was invented. Host validation there is implemented to **fail closed**: a request
whose `Deployment.APIBase` does not exactly match the allowlisted
`https://zx1.deepwl.net` host is rejected before any HTTP call is attempted, and DNS/
connect failures at dispatch time surface as a clear, actionable
`tokenbazaar_host_unreachable`-class error (never a silent retry, never a silent
fallback to `api.tokenbazaar.ai`, never a swallowed error). No live dispatch to
`zx1.deepwl.net` has been exercised end-to-end by this or any prior session — that
remains a release-gate item requiring a real, user-approved credential (see release
gates below).

## Scope

TokenBazaar (`tokenbazaar`) is a third-party model aggregator/reseller. It
fronts models originally trained and hosted by at least six different
upstream vendors (Anthropic, DeepSeek, Google, OpenAI, Moonshot, Z.ai) behind
its own gateway, billing, and request/response contract, and also resells
models from vendors GodarAI does not otherwise carry directly (MiniMax,
Alibaba/Qwen, ByteDance/Doubao, Black Forest Labs, Kuaishou, Baidu, xAI-style
Grok is not present, etc.).

This delivery is scoped to exactly the 26 model IDs already scaffolded under
`apps/model-discovery/providers/tokenbazaar/`:

```
claude-fable-5, claude-haiku-4-5, claude-opus-5, claude-sonnet-5,
deepseek-v4-flash, deepseek-v4-pro,
gemini-2.5-flash, gemini-2.5-pro, gemini-3-pro-image, gemini-3.1-flash-image,
gemini-3.1-flash-lite-image, gemini-3.1-pro-preview, gemini-3.5-flash,
gemini-3.6-flash, gemini-3.7-flash,
glm-5.2, glm-5v-turbo,
gpt-5.5, gpt-5.6-luna, gpt-5.6-sol, gpt-5.6-terra, gpt-image-2,
kimi-k2.6, kimi-k2.7-code, kimi-k3,
text-embedding-3-large
```

No model is added to or removed from this list in this delivery. See
"Model-list reconciliation" below for the verification evidence.

## Source manifest and observation log (2026-09-10, this session)

| Source | Result | Evidence |
|---|---|---|
| `https://docs.deepwl.cn/zx1/en` | **Unreachable** — DNS resolution fails (`NXDOMAIN`/`getaddrinfo ENOTFOUND`) | Re-attempted via `WebFetch` (`getaddrinfo ENOTFOUND docs.deepwl.cn`) and via direct `curl` with an OS-level `nslookup` (`Server failed` / `DNS request timed out`). This is the third consecutive failure of this exact host (first two: 2026-09-10 earlier in the day, per the task's own briefing; this session's retry: also 2026-09-10). Treat this domain as unreliable/stale, not as TokenBazaar's real documentation host. |
| `https://docs.deepwl.cn/zx1/en/texts/model-matrix` | **Unreachable**, same failure mode (same host) | Not independently re-probed; DNS failure is host-level, so the specific path is moot until the host resolves. |
| `https://tokenbazaar.ai` | **Reachable**, HTTP 200 | `curl` confirms `200`, ~2s response. |
| `https://tokenbazaar.ai/api/catalog/public` | **Reachable**, HTTP 200, valid JSON, 139 models | This is the same public/unauthenticated catalog endpoint the existing YAML scaffold was priced from. Confirmed still live and returning the same `catalog_version: "v2"` shape used previously. |
| `https://tokenbazaar.ai/llms.txt`, `/openapi.json`, `/swagger.json`, `/robots.txt`, `/sitemap.xml`, `/docs`, `/developers`, `/shelf` | HTTP 200 for all, but **all are the same client-rendered SPA shell** (`Content-Type: text/html`, identical `<!doctype html>` boilerplate) | `curl` cannot see real content because these are not static files — the site is a client-side React/TanStack SPA with a catch-all route. **This is why the automated/`WebFetch` fetch of `tokenbazaar.ai` earlier returned only the bare page title** — a plain HTTP fetch cannot execute the client-side router. |
| TokenBazaar's own in-app **`Docs` navigation link** (client-rendered, reached via the Browser tool, not a static file) | **Reachable and substantive** — a genuine first-party developer-documentation page | See "Confirmed wire contract" below. This is the "linked developer/API-docs section" the task instructed us to look for on `tokenbazaar.ai` itself, and it materially changes this delivery's risk posture versus the task's initial premise. |
| `GET https://tokenbazaar.ai/api/user/self` (unauthenticated) | `401`, as expected | Confirms this is a real authenticated endpoint, not probed further (no live credential in this session; no destructive/undocumented probing performed). |
| Authenticated `GET /v1/models` (documented, not called) | Not called — no live API key available in this session | The docs themselves describe this endpoint's exact response shape (see below); it was not invoked live because doing so requires a real TokenBazaar account/API key, which this session does not have and did not create (credential creation/spend requires explicit user approval per this task's own rules). |

**Key finding that revises the task's starting premise:** the briefing's
claim that "TokenBazaar's actual first-party API documentation ... could not
be reached" was true only for the `docs.deepwl.cn` domain (re-confirmed
unreachable again in this session). It was not true for TokenBazaar's own
site. `tokenbazaar.ai` hosts a complete first-party developer-documentation
page behind its own client-side router (`Docs` nav item), which a plain HTTP
fetch/curl cannot see but a rendered browser can. This is exactly the kind of
gap Phase 0 exists to find, and it substantially unblocks this delivery
compared to the "fail closed on everything" default.

## Confirmed wire contract (first-party, from `tokenbazaar.ai`'s own rendered `/docs` page, observed 2026-09-10)

These facts are now **first-party confirmed**, not guessed:

- **Base URL**: `https://zx1.deepwl.net` — **corrected 2026-09-10, later in this engagement, per direct user-provided information** (the user's own access to `docs.deepwl.cn` and/or their TokenBazaar account panel, which this session's own tooling could not reach). The value `https://api.tokenbazaar.ai`, recorded below in the "Confirmed wire contract" prose and still present in the conflict-register table until this correction, was this session's own independent finding from the storefront site's rendered `/docs` page — plausibly a customer-facing alias/CNAME in front of the real gateway, or simply a documentation error on the storefront's part. `tokenbazaar.ai` (storefront) and `deepwl.net`/`deepwl.cn` (the underlying "deepwl" platform TokenBazaar appears to be a branded reseller of, per the original `docs.deepwl.cn/zx1/...` URL given at the start of this engagement) are plausibly two domains for the same backend. **This session independently attempted to resolve `zx1.deepwl.net` via `curl`, the in-app browser, and `WebFetch` and got DNS timeout/`NXDOMAIN`-class failures on all three** — the same failure signature as `docs.deepwl.cn`, most likely a regional/network DNS-filtering artifact of this sandbox rather than evidence the host is wrong, but it means **no independent first-party confirmation of `zx1.deepwl.net` exists from this session's own tools — only the user's direct assertion.** Treat `zx1.deepwl.net` as authoritative for implementation (per direct user instruction), but any code that hardcodes it into a security-sensitive host allowlist must independently re-verify reachability from its own execution environment first, and must fail closed with a clear, actionable error — never a silent fallback to the unconfirmed `api.tokenbazaar.ai` — if it also cannot resolve or connect. The remaining contract details below (auth, the three envelopes, error table, rate-limit posture) came from the storefront's `/docs` page and are assumed to still apply to the `zx1.deepwl.net` host, since a white-label storefront's documentation content commonly mirrors the underlying platform's real contract even when a displayed convenience hostname differs — but this assumption itself is unverified and should be reconfirmed against `docs.deepwl.cn` content directly if that becomes accessible.
- **Auth (primary)**: `Authorization: Bearer <API_KEY>` (standard OpenAI-style), on every endpoint.
- **Auth (compatibility aliases, endpoint-scoped)**: `x-api-key: <API_KEY>` is honored only on Anthropic-style endpoints (`/v1/messages`, `/v1/models`); `x-goog-api-key` (header or `?key=` query) is honored only on Gemini-style endpoints (`/v1beta/...`, `/v1/models/...`). GodarAI must not accept or forward these as generic auth for non-matching endpoints.
- **Text generation — three parallel API styles, same model names, same underlying models**:
  - `POST /v1/chat/completions` — OpenAI Chat Completions-compatible request/response, SSE streaming with `data: [DONE]` terminator, standard `usage.prompt_tokens/completion_tokens/total_tokens`.
  - `POST /v1/responses` — OpenAI Responses-API-compatible; `reasoning: {effort: "low"|"medium"|"high"}` is documented **only on this route**, not on Chat Completions ("rely on the model default on Chat Completions"). `previous_response_id`, built-in tools, and `background: true` are explicitly **not supported**.
  - `POST /v1/messages` — Anthropic Messages-API-compatible; requires `max_tokens`; `thinking: {type: "enabled", budget_tokens: N}` for extended thinking.
  - The docs explicitly state: "All accept the same model names and return compatible responses" — i.e. TokenBazaar itself normalizes any of the 21 chat-capable models in our 26-model scope to whichever of these three envelopes the caller selects. This means GodarAI does **not** need a per-upstream-vendor adapter family; it needs one TokenBazaar adapter targeting the documented envelope(s), which is a materially different (and simpler) conclusion than the task's initial assumption.
- **Multimodal input**: `image_url` content parts (URL or base64) on the OpenAI-style route; `{"type":"image","source":{"type":"base64",...}}` content parts on the Anthropic-style route. Documented as gated by the Shelf's per-model vision flag.
- **Function calling**: standard OpenAI `tools`/`tool_choice` (`auto`/`none`/`required`/named) on `/v1/chat/completions`; the model returns `tool_calls`, caller appends a `role: "tool"` message to continue.
- **Image generation** (`gpt-image-2`, `gemini-3-pro-image`, `gemini-3.1-flash-image`, `gemini-3.1-flash-lite-image` in our scope): `POST /v1/images/generations`, `POST /v1/images/edits`, OpenAI-image-API-compatible. `gpt-image-2` returns base64 only (`response_format` must be `b64_json` or omitted; `url` is rejected). Per-model supported sizes/quality tiers come from the authenticated `/v1/models` `supported_parameters`, not a static table — the docs' own "Supported Image Models" table does not enumerate `gemini-3-pro-image`/`gemini-3.1-flash-image`/`gemini-3.1-flash-lite-image` by name, even though the full `/v1/models` catalog lists them with `endpoint_types: ["images"]`. Treat per-model image parameter support as **must be read live from `/v1/models`**, not hardcoded.
- **Embeddings** (`text-embedding-3-large` in our scope) — **re-confirmed 2026-09-11 with a full read of the rendered docs page (nav sidebar included): the gap is real, not a missed section.** The docs' own nav lists exactly: Quick Start, Create an API Key, Make Your First Request, Authentication, Model Names, Optional Parameters, Text Generation (LLM), Image Generation, Video Generation, SDK Examples, Response Format, Error Codes, Billing, Limits, API Reference — there is no "Embeddings" section anywhere, and no `POST /v1/embeddings` request/response example exists on the page, unlike chat/images/video which each have one. Two things strengthen the evidence versus the prior read, without closing the gap: (1) the page's own product description states "TokenBazaar is a unified AI API service with OpenAI-style compatibility. It supports text, image, video, **and embedding** capabilities across multiple models" — an explicit capability claim; (2) the authenticated `/v1/models` "Available Model IDs" table lists `text-embedding-3-large`, `text-embedding-ada-002`, and `qwen3-embedding-8b` all with `supported_endpoint_types: embeddings`. Given every other capability on this site (chat, responses, messages, images, video) follows an exact OpenAI/Anthropic-compatible path and shape, the highly likely (but still **unconfirmed**) contract is `POST /v1/embeddings` with a standard OpenAI-compatible `{model, input, encoding_format, dimensions}` request and `{object:"list", data:[{object:"embedding", embedding:[...], index}], model, usage:{prompt_tokens, total_tokens}}` response — matching GodarAI's own existing `platform.ProviderOperationOpenAIEmbeddings` operation already used for direct OpenAI dispatch in `services/silk-road/pkg/server/openai_operations.go`. This is now a **"strongly implied, not first-party confirmed"** gap (a step up from "unknown") — implementation may proceed using this shape, but must still be treated as a lower evidence tier than chat/images (which had worked examples) and must remain gated behind its own kill switch until a live probe confirms the exact path and response shape.
- **Errors**: detailed HTTP-status + `code` table is documented (`invalid_request`, `model_not_found`, `insufficient_user_quota`, `model_unavailable`, etc.), with an explicit "attribution rule" (a raw 4xx from the underlying model service is not passed through unless locally allow-listed; otherwise it becomes `503 model_unavailable` without leaking the upstream body). Every response carries `X-Tokenbazaar-Request-Id`.
- **Rate limits**: explicitly **not publicly numeric** — "TokenBazaar does not currently publish RPM, TPM, RPD, or concurrency figures." `429` with optional `Retry-After` is the only documented signal.
- **Pricing discovery**: the authenticated `GET /v1/models` endpoint (Bearer auth) returns, per model, `supported_endpoint_types`, `supported_parameters`, `pricing_kind`, `pricing_defaults`, `price_rows` (decimal-string `usd_amount` per `meter`/`unit`/`dimensions`/tier boundary), and optional `compare_price_rows` (official-price comparison). The docs say this reflects **"effective TokenBazaar USD prices for this API key"** — i.e. pricing may be account/tier-specific, not a single global number. This is an important nuance for Phase 1's source-drift job: a single global snapshot from the public `/api/catalog/public` endpoint (unauthenticated) may not represent every tenant's actual billed price. Recommend a follow-up product decision on whether GodarAI re-checks price via a representative platform-managed key rather than the public endpoint alone.

None of the above required a billable call or a live account — it came from reading TokenBazaar's own published documentation page in a rendered browser.

## Model-list reconciliation (Phase 1 exit evidence)

All 26 scaffolded model IDs were checked against the **live** TokenBazaar public catalog (`GET https://tokenbazaar.ai/api/catalog/public`, fetched 2026-09-10) by exact `api_name` match:

- All 26 are present, `"availability": "online"`.
- No model in the 26-item list is missing, renamed, or delisted.
- **No drift found.** The scaffold is current as of this observation.

Endpoint-type cross-check (from the same catalog payload and the docs' `/v1/models` "Available Model IDs" table) confirms the 26 models split as:
- 21 chat-capable (`chat/completions`, `generateContent`, `messages`, `responses`): all Claude/DeepSeek/Gemini-text/GLM/GPT-text/Kimi models in the list.
- 4 image models (`images`): `gemini-3-pro-image`, `gemini-3.1-flash-image`, `gemini-3.1-flash-lite-image`, `gpt-image-2`.
- 1 embeddings model (`embeddings`): `text-embedding-3-large`.

`apps/model-discovery/scripts/import-models-from-yaml.ts --providers tokenbazaar --dry-run` (run in this session) imports **exactly 26 models**, with zero errors, confirming every model has a resolvable `costs` block with a valid pricing-source URL under the importer's existing "must contain a `pricing` URL" rule. A full `--dry-run` across all 13 provider directories (including `tokenbazaar`) also succeeds with **133 total models and zero collisions** — `canonicalId = tokenbazaar/<model-id>` is distinct from every direct provider's `canonicalId` by construction (different provider-slug prefix), even where the underlying `<model-id>` string is shared (e.g. `claude-sonnet-5` exists both as `anthropic/claude-sonnet-5` and `tokenbazaar/claude-sonnet-5`, which is intentional and correct).

## Conflict / gap register

| # | Question | Status | Current safe behavior |
|---|---|---|---|
| 1 | Real base URL for inference | **Resolved, corrected** | `https://zx1.deepwl.net`, per direct user-provided information (their own access to `docs.deepwl.cn`/account panel) — supersedes this session's own independent finding of `https://api.tokenbazaar.ai` from the storefront's `/docs` page, which is downgraded to "unconfirmed alias, do not use." This session's own tools cannot independently resolve `zx1.deepwl.net` (DNS timeout, same signature as `docs.deepwl.cn`) — implementation must re-verify reachability from its own environment and fail closed if it also cannot connect. |
| 2 | Auth scheme | **Resolved** | `Authorization: Bearer <key>` primary; `x-api-key`/`x-goog-api-key` are endpoint-scoped compatibility aliases, not to be treated as generic. |
| 3 | Whether TokenBazaar re-exposes each vendor's native shape or normalizes to one shape | **Resolved, and reverses the task's initial assumption** | TokenBazaar normalizes all chat-capable models (21 of the 26) to whichever of `/v1/chat/completions` (OpenAI), `/v1/responses` (OpenAI Responses), or `/v1/messages` (Anthropic) the caller selects — one contract per API style, not one per upstream vendor. GodarAI's dispatch decision is "which of these three envelopes to speak," not "which of six vendor-native shapes to reimplement." |
| 4 | Streaming/SSE behavior | **Resolved for `/v1/chat/completions`** (`data: {...}` chunks, `finish_reason` on the terminal chunk, `data: [DONE]`); **documented but not independently verified end-to-end for `/v1/responses` and `/v1/messages`** (typed-event streaming, e.g. `response.output_text.delta`, `content_block_delta`/`message_stop`) — no live call was made in this session. | Gate `/v1/responses` and `/v1/messages` dispatch (including reasoning-effort/thinking pass-through) until a live/bounded probe confirms streaming end-to-end; `/v1/chat/completions` non-reasoning text dispatch is the best-evidenced path. |
| 5 | Error-code shape | **Resolved** (documented table above), **not live-verified** | Map documented codes into GodarAI's normalized error envelope; treat any undocumented status/code combination as `model_unavailable`-class rather than inventing a mapping. |
| 6 | Public catalog/pricing endpoint stability | **Partially resolved** | `GET /api/catalog/public` is unauthenticated, undocumented (not mentioned on the `/docs` page — it is an internal endpoint the Shelf UI itself calls), and returns `catalog_version: "v2"`, implying it has already changed shape at least once historically. Treat as **not a stable contract** — add source-drift detection (Phase 1) comparing it against checked-in YAML on a schedule, and prefer the **documented, authenticated `GET /v1/models`** as the long-term source of truth for both capability discovery and per-key pricing once a platform-managed credential exists. |
| 7 | Rate limits | **Resolved as "none published"** | Do not invent RPM/TPM/concurrency figures; rely on `429` + optional `Retry-After`; apply GodarAI's own tenant-plan limits ahead of dispatch as the effective control. |
| 8 | DeepSeek V4 Flash/Pro pricing anomaly | **Confirmed, quantified** (see "Pricing anomalies" below) | Not an implementation blocker; requires product/finance sign-off before enabling live traffic on these two models. |
| 9 | Gemini 3.6/3.7 Flash pricing anomaly | **Partially confirmed** — see caveat below; needs a full re-check against the *current* effective-dated direct-provider price row, not the one captured in this session's abbreviated read. | Same as above — surfaced, not resolved, not blocking cataloging. |
| 10 | GPT-5.6 family output-price premium anomaly | **Confirmed, quantified precisely** (see below) | Same as above. |
| 11 | `text-embedding-3-large` embeddings request/response shape | **Mechanism implemented and unit-tested against an *inferred*, not first-party-confirmed, contract (Embeddings Endpoint delivery, 2026-09-11).** A full re-read of the rendered `/docs` page (including nav sidebar) confirmed no "Embeddings" section and no worked request/response example exist there; no OpenAPI/schema document was found either. The contract is inferred from the docs' own product description ("...text, image, video, and embedding capabilities...") and the authenticated `GET /v1/models` catalog listing this model with `supported_endpoint_types: embeddings`, combined with every other confirmed TokenBazaar capability following an exact OpenAI/Anthropic-compatible shape with no deviation. | `POST /v1/embeddings` dispatch exists (`services/silk-road/pkg/server/tokenbazaar.go`: `tokenBazaarEmbeddingModels`, `validateTokenBazaarEmbeddingsDispatch`) and is fully reachable via explicit FQN (no code-level kill switch since 2026-09-11 -- see "Feature-flag removal" at the top of this document). A successful response is shape-checked (`validateTokenBazaarEmbeddingsResponseShape`) before being returned; a mismatch is rejected with a distinct `tokenbazaar_embeddings_response_shape_unexpected` error rather than forwarded blind. **Still open**: a live, approved-credential probe against `https://zx1.deepwl.net/v1/embeddings` to confirm the exact path and response envelope — not just that *some* response comes back. Until that probe succeeds, `production_ready` for embeddings stays `false` regardless of how well the code/tests behave. |
| 12 | Per-model image size/quality parameter sets for the 4 image models | **Mechanism implemented and tested; live population still open.** `/v1/images/generations` and `/v1/images/edits` dispatch now exists (`services/silk-road/pkg/server/tokenbazaar.go`), with an in-memory, TTL-expiring cache (`tokenBazaarImageParameterCache`, 6h TTL) keyed by model, carrying `SourceURL` provenance, and a decode/install path (`services/silk-road/pkg/server/tokenbazaar_image_discovery.go`: `parseTokenBazaarImageParameterSnapshots` / `refreshTokenBazaarImageParameterCache`) that parses TokenBazaar's documented authenticated `GET /v1/models` `supported_parameters` shape. **No code path calls that endpoint live yet** — there is no platform-managed TokenBazaar credential to call it with (same gap as the base-URL reachability item), so the cache stays empty in production today. No size/quality enum is hardcoded anywhere in Go or YAML. | A caller-requested `size`/`quality` is rejected with a distinct `tokenbazaar_image_parameters_unconfirmed` error unless a live snapshot exists for that exact model (never silently passed through, never confused with the generic `tokenbazaar_model_not_yet_supported`/`tokenbazaar_unsupported_model` errors). A request that omits `size`/`quality` entirely is allowed to proceed even with no snapshot, letting TokenBazaar apply its own model default rather than GodarAI guessing one — this is a deliberate, documented scoping of "validate the caller's requested size/quality," not a loophole: GodarAI never asserts a specific value is supported without live confirmation. **Follow-up before this can be closed**: wire `refreshTokenBazaarImageParameterCache` to an actual call against a live, user-approved TokenBazaar credential (at credential-test time and/or on a periodic discovery job), per Phase 1's existing source-drift pattern. |

## Pricing anomalies — status

All three previously flagged anomalies were checked against this session's live `GET /api/catalog/public` pull and the checked-in TokenBazaar YAML `costs` blocks vs. the corresponding direct-provider YAML `costs` blocks in this same repository.

1. **DeepSeek V4 Flash/Pro priced far below GodarAI's current upstream cost — CONFIRMED.** Live TokenBazaar `deepseek-v4-flash`: input $0.098 / output $0.196 per million tokens. The direct `apps/model-discovery/providers/deepseek/deepseek-v4-flash.yaml` currently carries a **more recent** cost row (`effective_from: 2026-08-16T16:00:00Z`) of input $0.44 / output $1.32 per million tokens. TokenBazaar's resale price is roughly **85% below** GodarAI's current direct-provider cost for this model. `deepseek-v4-pro` shows the same directional pattern. **This is a margin question for product/finance, not an implementation defect** — TokenBazaar may have its own volume pricing with DeepSeek; GodarAI must not enable this model for live traffic while selling it below cost without an explicit decision.
2. **Gemini 3.6/3.7 Flash priced above GodarAI's current upstream cost — NOW FULLY RECONCILED (corrected 2026-09-10, later in this engagement).** A fresh, direct pull of `GET https://tokenbazaar.ai/api/catalog/public` gives: `gemini-3.6-flash` price rows input $1.05 / output $5.25 (compare/"official" rows: input $1.50 / output $7.50); `gemini-3.7-flash` price rows input $1.20 / output $6.00 (compare/"official" rows: also $1.50 / $7.50). The *currently effective* direct-provider cost row (confirmed by `effective_from`/`effective_to` in `apps/model-discovery/providers/google-gemini/{gemini-3.6-flash,gemini-3.7-flash}.yaml`, both `2026-08-13T00:00:00Z`–`2027-01-01T00:00:00Z`, i.e. active today) is input $0.75 / output $3.75 for **both** models. Reconciled deltas: **Gemini 3.6 Flash is +40.0% above GodarAI's current cost (both input and output); Gemini 3.7 Flash is +60.0% above (both input and output)** — not the same premium for both models, and not the ~90% figure earlier sessions flagged (that number came from an earlier, less precise live read and should be discarded). TokenBazaar's own displayed "compare/official" price ($1.50/$7.50) matches Google's *pre-2026-08-13* rate exactly — i.e. TokenBazaar's own UI reference is stale relative to Google's most recent price cut — but TokenBazaar's actual charged price is not simply the old rate either; it sits at an intermediate markup GodarAI does not currently understand the origin of (a genuine TokenBazaar/Google commercial arrangement, a partially-applied price update, or something else). This is now a well-quantified, stable finding as of this correction, not an open reconciliation question — the remaining action is a product/finance decision, same as the other two anomalies.
3. **GPT-5.6 family output-price premium — CONFIRMED, PRECISELY QUANTIFIED.** TokenBazaar's live `gpt-5.6-sol` output price is $42.75 per million tokens; the direct `apps/model-discovery/providers/openai/gpt-5.6-sol.yaml` standard output rate is $30.00 per million tokens. $42.75 / $30.00 = 1.425 — **exactly the flagged +42.5% premium.** `gpt-5.6-luna` and `gpt-5.6-terra` were not independently re-derived to the same precision in this session and should be checked with the same method before the finance conversation.

### Interim mitigation applied 2026-09-10 (later in this engagement) — Google and OpenAI pinned to official price

Per an explicit product decision, anomalies #2 (Gemini 3.6/3.7 Flash) and #3 (GPT-5.6 family) are **no longer priced from TokenBazaar's live resale rate**. As an interim measure — pending a full commercial decision on whether/how to pass TokenBazaar's premium through — **every Google Gemini and every OpenAI model** in `apps/model-discovery/providers/tokenbazaar/` (all 9 Gemini models and all 6 OpenAI models, not just the 2+3 flagged ones) has had its `costs` block replaced verbatim with the exact, currently-effective official cost block from the corresponding direct-provider file (`apps/model-discovery/providers/google-gemini/*.yaml` / `apps/model-discovery/providers/openai/*.yaml`), including tiered/time-effective pricing rows where applicable. Each edited file carries an inline YAML comment above its `costs:` block documenting this override and pointing back to this section. Re-verified: `import-models-from-yaml.ts --providers tokenbazaar --dry-run` still imports exactly 26 models with zero errors, and `tokenbazaar-provider-importer.spec.ts` still passes (4/4) after the change.

This does **not** resolve anomaly #1 (DeepSeek) — DeepSeek was not touched, since TokenBazaar's price there is *below* GodarAI's cost (a margin opportunity, not a customer-facing overpricing risk) and the product decision only covered Google/OpenAI. DeepSeek's live TokenBazaar price is still used as-is.

This is a pricing-catalog-level mitigation only — it does not by itself change the (currently disabled) Phases 2–5 dispatch code's behavior beyond what it already reads from these YAML files at import time, and does not substitute for the explicit product/finance sign-off still required before `production_ready` can become `true`.

**Anomaly #1 (DeepSeek) remains unresolved and requires a decision.** Anomalies #2 and #3 have an interim mitigation (official price used) but still require a final decision on whether TokenBazaar should ever be sold at its own live rate for these models. Per this task's own exit rule, the readiness status stays `not production-ready` until that decision is made, regardless of how well the catalog/importer layer behaves.

## What is implemented in this delivery vs. explicitly gated

### Implemented and verified

- **Phase 0 (this document)**: full evidence-gathering, conflict register, and pricing-anomaly quantification above. This is a hard, honest gate — not a formality.
- **Phase 1, catalog layer**:
  - Confirmed the existing 26-file scaffold under `apps/model-discovery/providers/tokenbazaar/` needs **no changes** — it matches the live catalog exactly (no added/removed/renamed models).
  - Confirmed the **generic import path already handles TokenBazaar correctly** — no dedicated `apps/model-discovery/scripts/model-import/providers/tokenbazaar.ts` importer was added, per this task's own instruction not to add one speculatively. Evidence: `npx tsx scripts/import-models-from-yaml.ts --providers tokenbazaar --dry-run` imports exactly 26 models with zero errors; a full cross-provider `--dry-run` imports 133 models across 13 providers (including `tokenbazaar`) with zero collisions.
  - Added `apps/model-discovery/src/catalog/tokenbazaar-provider-importer.spec.ts`, mirroring the intent of `zai-provider-importer.spec.ts`: asserts exactly 26 models, `canonicalId` non-collision with the 6 direct providers, complete provenance-tagged pricing (both an upstream-vendor source URL and a `tokenbazaar.ai` pricing-page source URL per model, and at least one strictly positive price dimension per cost row), and runtime-profile schema validity where a profile is declared. All 4 tests pass; `tsc --noEmit` and `eslint` are clean on the new file; the full `apps/model-discovery` Jest suite (120 tests / 26 suites) still passes with this file added.

### Implemented and tested in the Phases 2–5 delivery (this session), still gated

This session implemented and tested real, working dispatch for the 21 chat-capable
in-scope models, deliberately narrower than the full 26-model surface, per its own
executable prompt:

- **Contracts** (`packages/contracts/platform`): no changes needed. TokenBazaar's
  `api_key`/Bearer auth against a fixed host is already representable with the
  existing `ProviderOperationChatCompletions` operation,
  `CredentialSourcePlatformManaged`, and `ResolvedProviderAuth{Mode: api_key}` —
  the same shapes DeepSeek/Moonshot/Z.ai already use.
- **`services/control-plane/internal/providerregistry/registry.go`**: registered
  `tokenbazaar` (capability `chat_completions` only; `credential_probe` +
  `chat_completions` operations; `DispatchRuntimeProfileRequired: true`; no
  `OperationNamespaces`/`LegacyModelPrefixes`, so a bare model id such as
  `claude-sonnet-5` still resolves to `anthropic`, never to `tokenbazaar` — only
  the explicit `tokenbazaar:<account>:<model-id>` FQN reaches this provider).
  `registry_test.go` extended with `TestTokenBazaarNeverCollidesWithDirectProviders`
  and updated coverage counts.
- **`services/control-plane/internal/config/provider_features.go` +
  `internal/proxy/dependencies.go` + `internal/bootstrap/build_gateway.go`**: added
  `TokenBazaarEnabled`/`SHAHRAH_TOKENBAZAAR_ENABLED`, defaulting to **false**
  (independent kill switch, unlike Z.ai's default-on flag) and wired end to end.
- **`services/control-plane/internal/proxy/handler.go` + `model_fqn.go`**: added
  `isTokenBazaarProvider` and a dispatch-time rejection
  (`503 provider_disabled`) when the flag is off — checked unconditionally (not
  gated behind `providerFeatures.Configured`, unlike Z.ai) so an unconfigured
  snapshot still fails closed to "disabled".
- **`services/control-plane/internal/billing/provider_pricing_strategy.go`**:
  added `tokenBazaarPricingStrategy` (embeds the shared `standardPricingStrategy`
  exact-decimal cost calculator unmodified — TokenBazaar's confirmed contract has
  no documented cache/conditional-rate dimensions).
- **`services/control-plane/internal/admin/provider_catalog_test.go`**: added
  slug-isolation tests (`TestNormalizeProviderSlug_KeepsTokenBazaarDistinct`,
  `TestMapDiscoveryProviders_KeepsTokenBazaarSeparateFromDirectProviders`,
  `TestDiscoveryProviderModels_TokenBazaarSharedModelIDStaysProviderIsolated`).
- **`services/silk-road/pkg/server/tokenbazaar.go` (new) + `tokenbazaar_test.go`
  (new)**: the adapter. Exact method/URL/header construction against
  `https://zx1.deepwl.net/v1/chat/completions`; an exact-host allowlist
  (`validateTokenBazaarBaseURL`) that fails closed on any lookalike, subdomain,
  port, userinfo, query, or path variation, and explicitly rejects the
  now-downgraded `api.tokenbazaar.ai` alias; a 21-model chat allowlist plus a
  1-model explicit-rejection map (`tokenbazaar_model_not_yet_supported: ...`)
  for the embeddings model (originally 5 models — the 4 image models were
  moved out of this rejection map into real dispatch by the Image Endpoints
  delivery above); the documented "attribution
  rule" (429 passes through raw with `Retry-After` intact, every other 4xx is
  normalized to `503 model_unavailable` with a fixed body — never leaking the
  upstream body — while `X-Tokenbazaar-Request-Id` is preserved on both paths);
  a response-size ceiling that fails closed rather than silently truncating;
  no adapter-level retry (single POST attempt only — documented as a deliberate
  retry-safety choice, matching the DeepSeek/Z.ai precedent of never retrying a
  non-idempotent completions POST); and a test confirming the caller's GodarAI
  `Authorization` header is never forwarded upstream. 13 tests, including SSE
  `[DONE]` passthrough and a two-call tool-call round trip.
- **`services/silk-road/pkg/server/adapters.go`**: registered the `tokenbazaar`
  adapter.
- **`services/silk-road/pkg/server/credential_probe.go` + `credential_probe_test.go`**:
  added a `tokenbazaar` case (POST `/v1/chat/completions`,
  `model: deepseek-v4-flash`, `max_tokens: 1` — the cheapest in-scope model per
  the pricing-anomaly evidence above) and an exact-host entry in the credential
  probe's own `safeCredentialProbeURL` allowlist.
- **`apps/model-discovery/prisma/migrations/202609100001_tokenbazaar_chat_completions_support/migration.sql`**
  (new): registers the `tokenbazaar` `Provider` row (enabled/selectable, so
  account creation and credential testing can be exercised end to end) and a
  write-only, `api_key`-only `ProviderConnectionSchema` — deliberately **without**
  a user-editable `base_url` field (unlike Z.ai's schema), so a tenant can never
  point TokenBazaar dispatch at a different host than the code-level allowlist.
- **`apps/tenant-admin/features/provider-accounts/{utils.ts,provider-docs.ts}`**:
  added TokenBazaar display name and docs link. The provider picker, wizard, and
  credential-schema rendering are already fully backend-schema-driven and needed
  no TokenBazaar-specific code once the migration above exists.
- **`apps/tenant-admin/features/provider-accounts/provider-connection-test-dialog.tsx`**
  (+ new `provider-connection-test-dialog.test.tsx`): added a TokenBazaar-specific
  branch showing a billable-test warning with an explicit displayed spend ceiling
  (~$0.001 USD, `deepseek-v4-flash`, `max_tokens: 1`) and an approval checkbox that
  must be checked before "Run" is enabled — on top of the dialog's existing
  requirement that a human open it and click Run (never on save/load/health
  check). 4 new tests confirm the checkbox gates both the button's `disabled`
  state and the actual submit handler (a synthetic form-level submit event is
  also blocked pre-approval).
- **`apps/tenant-admin/messages/modules/providerAccounts/{en,fa}.json`**: added
  `tokenbazaarBillableWarning`/`tokenbazaarApprovalCheckbox` in both locales;
  `pnpm run validate:i18n` passes (30/30 namespaces).

None of the above required, and none of the above performs, a live/billable call
to `https://zx1.deepwl.net` — all adapter tests use recorded/mocked HTTP round
trippers, per this delivery's own rule against live TokenBazaar calls in CI.

### Implemented and tested in the Image Endpoints delivery (this session, 2026-09-10)

This session implemented and tested real dispatch for the 4 in-scope image
models, on top of (and without regressing) the Phases 2–5 chat-completions
delivery above:

- **Evidence tier — do not conflate `gpt-image-2` with the 3 Gemini image
  models.** Per conflict #12's own source finding, TokenBazaar's "Supported
  Image Models" documentation table names `gpt-image-2` but does **not** name
  `gemini-3-pro-image`, `gemini-3.1-flash-image`, or
  `gemini-3.1-flash-lite-image`, even though the full `/v1/models` catalog
  lists all 4 with `endpoint_types: ["images"]`. All 4 share the identical
  code path (`validateTokenBazaarImageDispatch` in
  `services/silk-road/pkg/server/tokenbazaar.go`), but `gpt-image-2` carries
  first-party documentation-table evidence the 3 Gemini models do not. **Each
  of the 3 Gemini image models needs its own live probe before being marked
  ready, independent of `gpt-image-2`'s status** — `gpt-image-2` reaching
  `production_ready` does not imply the 3 Gemini models are ready, and vice
  versa, per the "one model's readiness never unlocks another" rule.
- **`services/control-plane/internal/providerregistry/registry.go`**: added
  `CapabilityImages`, `platform.ProviderOperationOpenAIImagesGenerations`, and
  `platform.ProviderOperationOpenAIImagesEdits` to TokenBazaar's registered
  capabilities/operations (previously chat-completions only).
  `registry_test.go` extended with `TestTokenBazaarSupportsImagesCapability`
  and `TestTokenBazaarImageModelsNeverCollideWithDirectProviders` (a bare
  `gpt-image-2`/`gemini-3-*-image` model id must never resolve to
  `tokenbazaar` — only the explicit
  `tokenbazaar:<account>:<model-id>` FQN does), plus the existing
  `TestDefinitionsSatisfyProviderContracts` exhaustive-contract test updated
  for the new capability/operations.
- **`services/silk-road/pkg/server/tokenbazaar.go`**: removed the 4 image
  models from the out-of-scope rejection map (now named
  `tokenBazaarUnsupportedModels`, holding only `text-embedding-3-large`);
  added `tokenBazaarImageModels` (the 4-model allowlist) and
  `validateTokenBazaarImageDispatch`, mirroring the chat path's own
  host-validation, runtime-profile enforcement, and error-shape conventions.
  Enforces:
  - `gpt-image-2`'s base64-only constraint
    (`validateTokenBazaarGPTImage2ResponseFormat`): `response_format: "url"`
    (or anything other than `b64_json`) is rejected before dispatch with
    `tokenbazaar_gpt_image_2_response_format_unsupported`; `b64_json` and an
    omitted `response_format` both dispatch.
  - Live-confirmed size/quality only (`validateTokenBazaarImageParameters`,
    see conflict #12 above): an unconfirmed requested size/quality is
    rejected with `tokenbazaar_image_parameters_unconfirmed` (no snapshot
    yet) or `tokenbazaar_image_parameter_not_supported` (snapshot exists but
    doesn't include the requested value) — both distinct from the generic
    `tokenbazaar_unsupported_model`/`tokenbazaar_model_not_yet_supported`
    errors.
  - A `multipart/form-data` `images/edits` request (an uploaded source image)
    is forwarded without this delivery parsing its fields — model-scoping and
    runtime-profile checks still apply, but `response_format`/size/quality
    validation only covers JSON-encoded requests (**follow-up**: multipart
    field validation is unimplemented).
  Reused, not duplicated: the same `openAICompatibleAdapter.Dispatch`, host
  allowlist (`validateTokenBazaarBaseURL`), 429-passthrough/4xx-normalization
  (`normalizeTokenBazaarResponse`), and response-size ceiling the chat path
  already uses — `resolveOpenAIOperation` already resolved
  `openai_images_generations`/`openai_images_edits` to the correct paths
  before this delivery (used by `xai`/`google-gemini` already), so no new
  path-resolution logic was needed there.
- **`services/silk-road/pkg/server/tokenbazaar_image_discovery.go`** (new):
  `parseTokenBazaarImageParameterSnapshots` / `refreshTokenBazaarImageParameterCache`
  — the live-discovery parser/installer described in conflict #12. Not yet
  wired to an automatic trigger (no live credential to call with).
- **`services/control-plane/internal/config/provider_features.go` +
  `internal/proxy/dependencies.go` + `internal/bootstrap/build_gateway.go`**:
  added `TokenBazaarImagesEnabled` / `SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`,
  defaulting to **false**, independent of `TokenBazaarEnabled` /
  `SHAHRAH_TOKENBAZAAR_ENABLED` (chat). Verified independent in both
  directions by `TestTokenBazaarImagesFlagIsIndependentOfChatFlag`.
- **`services/control-plane/internal/proxy/images_handler.go`**: added a
  dispatch-time rejection (`503 provider_disabled`) when
  `TokenBazaarImagesEnabled` is off, mirroring the chat path's own
  `isTokenBazaarProvider` check in `handler.go` — checked unconditionally
  (not gated behind `providerFeatures.Configured`) so an unconfigured/
  zero-value snapshot still fails closed to "disabled". Verified by
  `TestImagesHandlerTokenBazaarImagesDisabledByDefault` and
  `TestImagesHandlerTokenBazaarImagesFlagIsolatedFromChatFlag`.
- **`services/control-plane/internal/billing/provider_pricing_strategy.go`**:
  extended `tokenBazaarPricingStrategy` with a `FinalizeImage` override
  (`tokenBazaarImageModelFamily` classifies each of the 4 image models to the
  cost-dimension shape its catalog `costs` block actually uses — OpenAI's
  per-token/per-image-token shape for `gpt-image-2`, reusing
  `openAIImageCostBreakdown` unmodified; Google Gemini's
  per-token/per-image-size-tier shape for the 3 Gemini models, reusing
  `googleGeminiImageCostBreakdown` unmodified). This sources cost dimensions
  from the already-pinned, verbatim-copied official catalog `costs` blocks in
  `apps/model-discovery/providers/tokenbazaar/{gpt-image-2,gemini-3-pro-image,
  gemini-3.1-flash-image,gemini-3.1-flash-lite-image}.yaml` (see the 2026-09-10
  pinning note at the top of each file) — no TokenBazaar-specific pricing
  logic was written; the delivery only routes to the correct existing
  exact-decimal (`shopspring/decimal`) calculator per model family. Verified
  by `TestTokenBazaarPricingStrategyRoutesGPTImage2ToOpenAIExactDecimalBreakdown`
  and `TestTokenBazaarPricingStrategyRoutesGeminiImageToGoogleGeminiExactDecimalBreakdown`
  (both assert an independently-computed exact-decimal expected total, not a
  float epsilon guess).
- **`apps/tenant-admin` / Playground**: **not extended in this delivery.**
  Adding a size/quality control would either (a) require a static
  guessed-size dropdown — exactly the invented compatibility this delivery
  exists to avoid — or (b) require the live discovery mechanism to already be
  wired to a real account, which it is not yet (see conflict #12). Left
  untouched; `FOLLOW-UP` once a live credential exists.
- **Tests**: `services/silk-road/pkg/server/tokenbazaar_test.go` and the new
  `tokenbazaar_image_discovery_test.go` add ~20 test cases covering: routing
  for all 4 models × both operations; the `text-embedding-3-large` regression
  (still rejects exactly as before); `gpt-image-2` response_format
  acceptance/rejection (including an end-to-end mocked-HTTP dispatch proving
  the rejection happens before any upstream call); size/quality confirmation
  and expiry; runtime-profile enforcement; multipart pass-through; no
  secret/key leakage on the image path (mirroring the chat path's own
  standard); and the `/v1/models` response parser. No live or billable
  TokenBazaar call is made anywhere in this test suite.

None of the above required, and none of the above performs, a live/billable
call to `https://zx1.deepwl.net` — the image adapter tests use recorded/mocked
HTTP round trippers, same rule as the chat path.

### Implemented and tested in the Embeddings Endpoint delivery (this session, 2026-09-11)

- **Registry**: `services/control-plane/internal/providerregistry/registry.go`
  adds `CapabilityEmbeddings` and `platform.ProviderOperationOpenAIEmbeddings`
  to TokenBazaar's registered capabilities/operations, alongside the existing
  chat and images ones. `TestTokenBazaarSupportsEmbeddingsCapability` and
  `TestTokenBazaarEmbeddingsModelNeverCollidesWithDirectProviders` confirm
  `text-embedding-3-large` routes to the embeddings dispatch path and never
  collides with the direct `openai` provider's own embeddings dispatch (a
  bare, unqualified `text-embedding-3-large` id never resolves to
  `tokenbazaar`; only the explicit `tokenbazaar:<account>:text-embedding-3-large`
  FQN does).
- **Adapter dispatch**: `services/silk-road/pkg/server/tokenbazaar.go` removes
  `text-embedding-3-large` from `tokenBazaarUnsupportedModels` (now empty) and
  adds `tokenBazaarEmbeddingModels` / `validateTokenBazaarEmbeddingsDispatch`,
  mirroring the structure of `validateTokenBazaarDispatch`/
  `validateTokenBazaarImageDispatch`: model-scoping, runtime-profile
  (`embeddings` endpoint, `http` transport) enforcement, a required `input`
  field, `encoding_format` validated against `{float, base64}`, and
  `dimensions` validated against the `[1, 3072]` bound already pinned in
  `apps/model-discovery/providers/tokenbazaar/text-embedding-3-large.yaml`'s
  own `params` entry (not a newly-invented bound). Dispatch targets
  `POST /v1/embeddings` via the existing generic `openAICompatibleAdapter` and
  `platform.ProviderOperationOpenAIEmbeddings` — no new HTTP-construction code
  was needed there, since that generic path already resolves this operation
  to `POST {base}/embeddings`.
- **Response-shape defense-in-depth**: because this contract is inferred, not
  confirmed (see conflict #11), a successful response is additionally
  shape-checked in the adapter itself
  (`validateTokenBazaarEmbeddingsResponseShape`): it must decode as
  `{object:"list", data:[...]}` with a non-empty `embedding` value on every
  entry, or dispatch fails with a distinct
  `tokenbazaar_embeddings_response_shape_unexpected` error instead of
  forwarding an unrecognized shape blind (which could otherwise be silently
  mis-parsed or mis-billed further up the stack) or panicking in a downstream
  fixed-struct unmarshal.
- **Feature flag**: `SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED` (default `false`)
  added to `config/provider_features.go`, `proxy/dependencies.go`,
  `bootstrap/build_gateway.go`, and enforced fail-closed in
  `proxy/embeddings_handler.go` (mirroring the images flag's fail-closed
  wiring in `proxy/images_handler.go`), independent of both
  `SHAHRAH_TOKENBAZAAR_ENABLED` and `SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED` in
  both directions. `tokenbazaar` was also added to
  `supportsEmbeddingsProvider` in that same file (it was previously absent,
  since the model had no dispatch at all).
- **Billing**: no `FinalizeUsage` override was needed —
  `tokenBazaarPricingStrategy` already inherits
  `standardPricingStrategy.FinalizeUsage`, which bills embeddings usage
  (`prompt_tokens`/`total_tokens`, `completion_tokens` always 0) via the same
  generic exact-decimal `costBreakdown` already used for every other
  TokenBazaar chat model, using the already-pinned official OpenAI rate
  (`input_cost_per_token: 1.3e-7`, `output_cost_per_token: 0`). Confirmed by
  `services/control-plane/internal/billing/tokenbazaar_embeddings_pricing_test.go`
  with exact-decimal assertions (via `shopspring/decimal`, not a float
  epsilon) for a representative input-token count, and a dedicated assertion
  that output cost is always exactly 0.
- **Tests**: `services/silk-road/pkg/server/tokenbazaar_test.go` adds routing
  coverage (embeddings operation accepted, chat/images operations still
  rejected as a generic unsupported model), runtime-profile enforcement, the
  required-`input` check, `encoding_format`/`dimensions` bound validation
  (including the inclusive 1/3072 boundaries and out-of-range/non-integer
  cases), an end-to-end mocked-HTTP dispatch asserting the exact method/URL
  (`POST https://zx1.deepwl.net/v1/embeddings`)/Authorization
  header/Content-Type/forwarded body and successful response parsing, a
  dedicated malformed-response test (not-JSON, wrong `object` field, empty
  `data`, missing embedding vector, and a chat-shaped response all rejected
  with a distinguishable error), 4xx normalization/no-upstream-body-leak
  parity with chat/images, no API-key leakage in validation errors, and no
  forwarding of the caller's own Authorization header. `registry_test.go`,
  `provider_features_test.go`, `embeddings_handler_test.go`, and
  `tokenbazaar_embeddings_pricing_test.go` add the control-plane-side routing,
  flag-isolation, handler-flag-enforcement, and billing coverage respectively.
  No live or billable TokenBazaar call is made anywhere in this test suite.
- **Regression**: the full existing `tokenbazaar_test.go` and
  `tokenbazaar_image_discovery_test.go` suites (chat and image dispatch for
  the other 25 models) were re-run unchanged and still pass; the one
  intentional behavior change to an existing test was
  `TestTokenBazaarRejectsOutOfScopeModelsWithSpecificError`, renamed to
  `TestTokenBazaarUnsupportedModelsMapIsNowEmpty` because its entire premise
  (text-embedding-3-large stays rejected) is what this delivery deliberately
  reverses; `TestTokenBazaarChatOperationRejectsImageModels` and the 4 image
  models' own dispatch tests were not touched and still pass as before.

### Implemented and tested in the Responses Endpoint delivery (this session, 2026-09-11)

See the top-of-document "Responses endpoint delivery" section for the full
reuse-vs-new-code accounting, the two assumptions found in
`responses_handler.go`, and the streaming evidence-tier caveat. Summary:

- **Registry**: `services/control-plane/internal/providerregistry/registry.go`
  adds `CapabilityResponses` and `platform.ProviderOperationOpenAIResponsesCreate`
  to TokenBazaar's registered capabilities/operations, alongside the existing
  chat/images/embeddings ones. `TestTokenBazaarSupportsResponsesCapability` and
  `TestTokenBazaarResponsesModelsNeverCollideWithDirectProviders` confirm the 6
  in-scope models route to the Responses dispatch path and never collide with
  the corresponding direct `openai`/`deepseek` providers' own dispatch (a
  bare, unqualified model id never resolves to `tokenbazaar`; only the
  explicit `tokenbazaar:<account>:<model-id>` FQN does), and that the
  create-only scope is enforced (the 6 stateful Responses resource operations
  are explicitly asserted absent from TokenBazaar's registered operations).
- **Adapter dispatch**: `services/silk-road/pkg/server/tokenbazaar.go` adds
  `tokenBazaarResponsesModels` (the 6-model allowlist) and
  `validateTokenBazaarResponsesDispatch`, mirroring the structure of
  `validateTokenBazaarChatDispatch`/`validateTokenBazaarEmbeddingsDispatch`:
  model-scoping, runtime-profile (`responses` endpoint, `http`/`sse`
  transport) enforcement, a required `input`-or-`instructions` field, and the
  three documented fail-closed rejections
  (`validateTokenBazaarResponsesPreviousResponseID`,
  `validateTokenBazaarResponsesTools`,
  `validateTokenBazaarResponsesBackground`). Dispatch targets
  `POST /v1/responses` via the existing generic `openAICompatibleAdapter` and
  `platform.ProviderOperationOpenAIResponsesCreate` — no new HTTP-construction
  or path-resolution code was needed there, since
  `resolveOpenAIOperation` already resolved this operation to
  `POST {base}/responses` before this delivery (used by every
  OpenAI-compatible adapter already).
- **Control plane**: `services/control-plane/internal/proxy/responses_handler.go`
  adds `"tokenbazaar"` to `supportsResponsesCreateProvider` (deliberately NOT
  to `supportsResponseResourcesProvider`), a `providerFeatures
  ProviderFeatureConfig` field on `ResponsesHandler` (a genuine pre-existing
  gap this delivery closed — see the top-of-document assumption note), and
  `rejectDisabledDirectResponsesProvider` enforcing
  `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED` fail-closed for both direct and
  virtual-model-routed TokenBazaar Responses dispatch.
- **Feature flag**: `SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED` (default `false`)
  added to `config/provider_features.go`, `proxy/dependencies.go`,
  `bootstrap/build_gateway.go`, independent of
  `SHAHRAH_TOKENBAZAAR_ENABLED`/`SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`/
  `SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED` in all directions — confirmed by
  `TestResponsesHandlerTokenBazaarResponsesDisabledByDefault` and
  `TestResponsesHandlerTokenBazaarResponsesFlagIsolatedFromOtherThreeFlags`.
- **Billing**: no `FinalizeUsage` override was needed or added — see the
  top-of-document reuse accounting.
- **Tests**:
  - `services/silk-road/pkg/server/tokenbazaar_test.go` adds routing coverage
    for all 6 in-scope models and a negative table for the other 20 models
    (`TestTokenBazaarResponsesDispatchRoutesForAllSixInScopeModels`,
    `TestTokenBazaarResponsesOperationRejectsOutOfScopeModels`), runtime-profile
    enforcement, the required-input-or-instructions check, dedicated
    table-driven tests for each of the three fail-closed rejections
    (`TestTokenBazaarResponsesRejectsPreviousResponseID`,
    `TestTokenBazaarResponsesRejectsBuiltinTools`,
    `TestTokenBazaarResponsesRejectsBackgroundTrue`) plus an end-to-end
    mocked-HTTP proof all three happen before any upstream call
    (`TestTokenBazaarResponsesThreeRejectionsHappenBeforeAnyHTTPCall`), an
    exact method/URL/header/body construction test against
    `https://zx1.deepwl.net/v1/responses`
    (`TestTokenBazaarResponsesDispatchExactRequestConstruction`), a
    fixture-based `reasoning.effort` passthrough test across all 6 models
    (`TestTokenBazaarResponsesReasoningEffortPassesThroughUnmodified`), an SSE
    typed-event passthrough test
    (`TestTokenBazaarResponsesSSEStreamingPassesThroughUntilTerminalEvent`),
    4xx normalization/no-upstream-body-leak parity, no API-key leakage in
    validation errors, and no forwarding of the caller's own Authorization
    header.
  - `services/control-plane/internal/providerregistry/registry_test.go`,
    `services/control-plane/internal/proxy/responses_tokenbazaar_test.go` (new)
    add the control-plane-side routing, create-only capability, flag-isolation,
    and reasoning-effort-passthrough coverage respectively (the last mirroring
    the existing xAI Responses coverage pattern).
  - No live or billable TokenBazaar call is made anywhere in this test suite.
- **Regression**: the full existing `tokenbazaar_test.go` and
  `tokenbazaar_image_discovery_test.go` suites (chat, image, and embeddings
  dispatch for the other 20 models) were re-run unchanged and still pass; no
  existing test was renamed or had its assertions changed. The full
  `services/control-plane/internal/proxy`, `internal/config`,
  `internal/providerregistry`, `internal/billing`, and `internal/bootstrap`
  suites, plus the full `services/control-plane` and `services/silk-road`
  module test suites, were re-run and pass with no regressions.

### Implemented and tested in the Anthropic Messages endpoint delivery (this session, 2026-09-11)

See the top-of-document "Anthropic Messages endpoint delivery" section for
the full routing-architecture decision, reuse-vs-new-code accounting, the
tool-use/prompt-caching decision, and the streaming evidence-tier caveat.
Summary:

- **Routing**: a new, narrow sibling handler
  (`proxy.TokenBazaarMessagesHandler`) reuses the standard FQN -> deployment
  -> providerregistry capability -> Silk Road adapter flow, wired via a thin
  shim (`proxy.NewAnthropicMessagesRouter`) in `routes_proxy.go` in place of
  the bare `AnthropicNativeHandler.Handle` registration for
  `POST /v1/messages`. `AnthropicNativeHandler`'s own dispatch/routing/
  account-allowlist/WIF/files/batches logic is completely untouched; the only
  change to that file is a behavior-preserving extract-method refactor
  (`streamAnthropicShapedResponse`) needed to satisfy this repo's `dupl`
  lint rule once a second, structurally-identical Anthropic SSE relay
  existed.
- **Registry**: `services/control-plane/internal/providerregistry/registry.go`
  adds a new `CapabilityMessages` constant (no prior "anthropic" registry
  precedent existed to mirror — see the routing-architecture section above)
  and `platform.ProviderOperationAnthropicMessages` to TokenBazaar's
  registered capabilities/operations. `TestTokenBazaarSupportsMessagesCapability`
  and `TestTokenBazaarMessagesModelsNeverCollideWithDirectProviders` confirm
  the 4 in-scope models route to the Messages dispatch path and never
  collide with the direct `anthropic` provider's own dispatch (a bare,
  unqualified model id never resolves to `tokenbazaar`; only the explicit
  `tokenbazaar:<account>:<model-id>` FQN does).
- **Adapter dispatch**: `services/silk-road/pkg/server/tokenbazaar.go` adds
  `tokenBazaarMessagesModels` (the 4-model allowlist) and
  `validateTokenBazaarMessagesDispatch`, mirroring the structure of
  `validateTokenBazaarChatDispatch`/`validateTokenBazaarResponsesDispatch`:
  model-scoping, runtime-profile (`messages` endpoint, `http`/`sse`
  transport) enforcement, a required `max_tokens` (positive integer) field,
  a required `messages` field, and the two fail-closed rejections
  (`validateTokenBazaarMessagesToolsUnsupported`,
  `validateTokenBazaarMessagesPromptCachingUnsupported`). Dispatch targets
  `POST /v1/messages` via the existing generic `openAICompatibleAdapter` and
  `platform.ProviderOperationAnthropicMessages` — the only new
  path-resolution code needed was one `case` in
  `services/silk-road/pkg/server/openai_operations.go`'s
  `resolveCoreOpenAIOperation` (a pure method/path mapping, since the
  request body is already Anthropic-native JSON by construction).
- **Feature flag**: `SHAHRAH_TOKENBAZAAR_MESSAGES_ENABLED` (default `false`)
  added to `config/provider_features.go`, `proxy/dependencies.go`,
  `bootstrap/build_gateway.go`, independent of
  `SHAHRAH_TOKENBAZAAR_ENABLED`/`SHAHRAH_TOKENBAZAAR_IMAGES_ENABLED`/
  `SHAHRAH_TOKENBAZAAR_EMBEDDINGS_ENABLED`/`SHAHRAH_TOKENBAZAAR_RESPONSES_ENABLED`
  in all directions — confirmed by
  `TestTokenBazaarMessagesHandlerDisabledByDefault` and
  `TestTokenBazaarMessagesHandlerFlagIsolatedFromOtherFourFlags`.
- **Model access**: a new `platform.ModelTypeMessages` constant
  (`packages/contracts/platform/tenant_provider_account.go`) gates access to
  this route independently of `ModelTypeChat`, mirroring how
  `ModelTypeResponses` already gates the Responses route independently —
  wired into `admin/provider_catalog.go`'s `mapModeToModelType` so operators
  can grant it through the existing tenant-admin configuration surface.
- **Billing**: no `FinalizeUsage` override was needed or added — see the
  top-of-document reuse accounting.
- **Tests**:
  - `services/silk-road/pkg/server/tokenbazaar_test.go` adds routing
    coverage for all 4 in-scope models and a negative table for the other 22
    models (`TestTokenBazaarMessagesDispatchRoutesForAllFourInScopeModels`,
    `TestTokenBazaarMessagesOperationRejectsOutOfScopeModels`), runtime-profile
    enforcement, `max_tokens` bound validation (absent/null/zero/negative/
    non-integer/wrong-type), the required-`messages` check, dedicated
    table-driven tests for both fail-closed rejections
    (`TestTokenBazaarMessagesRejectsToolUse`,
    `TestTokenBazaarMessagesRejectsPromptCaching`) plus an end-to-end
    mocked-HTTP proof both happen before any upstream call
    (`TestTokenBazaarMessagesThreeRejectionsHappenBeforeAnyHTTPCall`), an
    exact method/URL/header/body construction test against
    `https://zx1.deepwl.net/v1/messages`
    (`TestTokenBazaarMessagesDispatchExactRequestConstruction`), a
    fixture-based `thinking.budget_tokens` passthrough test across all 4
    models (`TestTokenBazaarMessagesThinkingPassesThroughUnmodified`), an SSE
    content_block_delta/message_stop passthrough test
    (`TestTokenBazaarMessagesSSEStreamingPassesThroughUntilTerminalEvent`)
    plus a malformed-SSE-chunk passthrough test
    (`TestTokenBazaarMessagesMalformedSSEChunkPassesThroughWithoutError`),
    4xx normalization/no-upstream-body-leak parity, no API-key leakage in
    validation errors, and no forwarding of the caller's own Authorization
    header.
  - `services/control-plane/internal/providerregistry/registry_test.go`
    adds the control-plane-side capability/collision coverage;
    `services/control-plane/internal/proxy/tokenbazaar_messages_handler_test.go`
    (new) adds routing-shim coverage
    (`TestRequestTargetsTokenBazaarMessages`,
    `TestAnthropicMessagesRouterFallsBackToNativeForNonTokenBazaarModels`),
    flag-isolation coverage
    (`TestTokenBazaarMessagesHandlerDisabledByDefault`,
    `TestTokenBazaarMessagesHandlerFlagIsolatedFromOtherFourFlags`),
    end-to-end dispatch/billing coverage
    (`TestTokenBazaarMessagesHandlerDispatchesSuccessfully`,
    `TestTokenBazaarMessagesHandlerThinkingPassesThroughUnmodified`,
    `TestTokenBazaarMessagesHandlerNeverForwardsGodarAIAuthorizationHeader`),
    and new malformed/unexpected-SSE-event coverage for the shared
    `parseAnthropicSSEUsage`/`anthropicSSEIsError` helpers
    (`TestParseAnthropicSSEUsageHandlesMalformedAndUnexpectedEvents` — a gap
    that existed even for the direct Anthropic path before this delivery,
    now closed for both).
  - No live or billable TokenBazaar call is made anywhere in this test suite.
- **Regression**: the full existing `tokenbazaar_test.go` and
  `tokenbazaar_image_discovery_test.go` suites (chat, image, embeddings, and
  responses dispatch for the other 22 models) were re-run unchanged and
  still pass; no existing test was renamed or had its assertions changed.
  The full existing `anthropic_native_handler_test.go` suite (the direct
  `anthropic` provider's account allowlist, code-execution pricing policy,
  and feature-flag independence) was re-run unchanged and still passes,
  confirming the `streamAnthropicShapedResponse` extraction preserved
  `AnthropicNativeHandler`'s behavior exactly. The full
  `services/control-plane/internal/proxy`, `internal/config`,
  `internal/providerregistry`, `internal/billing`, `internal/admin`, and
  `internal/bootstrap` suites, plus the full `services/control-plane`,
  `services/silk-road`, and `packages/contracts/platform` module test
  suites, were re-run and pass with no regressions.

### Explicitly gated (still not implemented) — and why

- **`/v1/responses` for the other 15 chat-capable models**: not wired into
  dispatch. The 6-model Responses scope (4 OpenAI family + 2 DeepSeek) is
  that delivery's own deliberate boundary, not an evidence-based exclusion of
  the other 15 models — TokenBazaar's own docs claim `/v1/responses` works
  identically for all 21 chat-capable models.
- **`/v1/messages` for models other than the 4 Claude models**: no other
  TokenBazaar model gains this operation (Anthropic Messages is a Claude-
  native envelope; TokenBazaar's own docs describe the three text envelopes —
  chat/responses/messages — as normalizing to the same underlying model, so
  a non-Claude model speaking `/v1/messages` is plausible in principle, but
  this delivery's own executable prompt scoped it to exactly the 4 Claude
  models TokenBazaar carries, mirroring the Responses delivery's own
  narrower-than-full-surface scoping). Requesting `/v1/messages` for any
  other TokenBazaar model still rejects with the generic
  `tokenbazaar_unsupported_model` error. See "Anthropic Messages endpoint
  delivery" above for the full accounting — dispatch for the 4 Claude models
  is implemented and unit-tested and fully reachable (no code-level kill
  switch since 2026-09-11, see "Feature-flag removal" at the top of this
  document); a live probe confirming streaming/thinking end to end is still
  the open release gate.
- **`text-embedding-3-large`**: no longer gated as out-of-scope — see the
  "Embeddings endpoint delivery" section above and conflict-register item #11.
  Dispatch is implemented and unit-tested against an *inferred* contract and
  is fully reachable (no code-level kill switch since 2026-09-11); a live
  probe confirming the actual path/response envelope is still the open
  release gate. `FOLLOW-UP`.
- **The 4 image models' live per-model parameter discovery** (conflict #12):
  mechanism built and unit-tested; not yet wired to an automatic live trigger.
  A caller-requested size/quality that cannot be live-confirmed is rejected,
  not guessed. `FOLLOW-UP`.
- **Phase 6 (full wallet/settlement integration beyond cost computation)**: the
  exact-decimal cost *computation* now exists
  (`tokenBazaarPricingStrategy`/`standardPricingStrategy`), but end-to-end wallet
  settlement, invoicing, and reconciliation against a live account are
  unimplemented. `FOLLOW-UP`.
- **Phase 7 (Playground UX)**: not implemented beyond the provider-accounts
  wizard/connection-test work above. `FOLLOW-UP`.
- **Phase 8 (Persian product docs)**: `apps/docs/docs/providers/tokenbazaar.mdx`
  not created (tenant-admin i18n strings for the credential UI were added, which
  is a distinct, narrower thing). `FOLLOW-UP`.
- **Phase 9 (readiness CLI, canary/kill-switch operational wiring)**:
  `services/control-plane/cmd/tokenbazaar-readiness` not created. As of
  2026-09-11 (see "Feature-flag removal" above), there is no code-level kill
  switch at all for any TokenBazaar capability — the build-time/config-time
  flags that previously existed here were removed by explicit product
  decision, and the fuller canary/rollback operational tooling the Z.ai
  reference has was never built either. `FOLLOW-UP`.
- **`production_ready`**: **stays false, for chat, images, embeddings,
  responses, and messages.** The provider is registered and fully wired end
  to end in code for all five capabilities, reachable via explicit
  `tokenbazaar:<account>:<model-id>` FQN with **no independent code-level
  kill switch** (see "Feature-flag removal" at the top of this document for
  why, and the accepted trade-off that decision carries) — but no live
  TokenBazaar account has ever exercised any path (no live call was made in
  Phase 0, the Phases 2–5 delivery, the Image Endpoints delivery, the
  Embeddings Endpoint delivery, the Responses Endpoint delivery, the
  Anthropic Messages endpoint delivery, or any prior session), the three
  chat pricing anomalies remain undecided by product/finance, the image
  per-model parameter live-discovery cache (conflict #12) has never been
  populated from a real account, the embeddings contract itself is inferred
  rather than first-party-confirmed (conflict #11), and the Responses and
  Messages streaming contracts have never been independently live-verified
  end to end (same evidence tier as chat_completions streaming — see
  "Responses endpoint delivery" and "Anthropic Messages endpoint delivery"
  above). In practice, the only things standing between TokenBazaar and live
  traffic today are (a) no tenant holding a real TokenBazaar provider
  account/credential yet, and (b) the release gates below (live-probe
  confirmation of each wire contract, the pricing-anomaly sign-off) — those
  remain the actual reasons `production_ready` stays `false`, not a flag.
  Implemented-and-tested is not the same claim as production-ready — this is
  truer for embeddings (unconfirmed contract), responses (unconfirmed
  streaming), and messages (unconfirmed streaming) than for chat or images.

## Release gates (unchanged from the task's own bar, extended for images, embeddings, responses, and messages)

Do not mark TokenBazaar, or any individual model on it, production-ready until:
1. The chat/completions wire contract is additionally confirmed by a live, approved TokenBazaar account (this session's evidence is first-party documentation, not a live call).
2. The `/v1/responses` and `/v1/messages` streaming/reasoning contracts are live-verified if GodarAI intends to expose reasoning effort/thinking budgets through TokenBazaar.
3. The `text-embedding-3-large` embeddings contract gap (conflict #11) is closed.
4. The three chat pricing anomalies have an explicit, recorded product/finance decision (not just this document's quantification).
5. Phases 2–5 (credentials, public routes, validation, Silk Road adapter) are
   implemented and pass their own tests as of this delivery — see "Implemented
   and tested in the Phases 2–5 delivery" above. Phases 6–9 (full wallet
   settlement beyond cost computation, Playground UX, Persian docs, readiness
   CLI, canary/kill-switch operational wiring) remain unimplemented.
6. Dashboards, a rollback drill, and (per "Feature-flag removal" above) the
   model-discovery catalog's `enabled` flag actually being wired to
   control-plane's dispatch decision -- today it only affects catalog
   visibility/selection, not dispatch, so it cannot yet serve as the
   operational kill switch this gate originally described. Until that
   connection is built (a separate, general, all-providers piece of work,
   not scoped to this delivery), TokenBazaar's five capabilities have no
   code-level kill switch at all; this gate is not met until either that
   connection exists or another explicit operational safeguard is put in
   its place.
7. **(Images, all 4 models)** The live per-model image-parameter discovery
   cache (conflict #12) is actually populated from a real, user-approved
   TokenBazaar credential at least once — `refreshTokenBazaarImageParameterCache`
   is implemented and unit-tested but not yet wired to any live trigger. Until
   then, any real caller-requested `size`/`quality` fails closed as
   unconfirmed, by design.
8. **(Images, `gpt-image-2` only)** A live, approved TokenBazaar account
   confirms `POST /v1/images/generations`/`edits` actually dispatch correctly
   for `gpt-image-2` end to end (base64 response shape, usage/cost fields) —
   this session's evidence is first-party documentation-table listing, not a
   live call.
9. **(Images, the 3 Gemini models only — independent of gate 8)** A live,
   approved TokenBazaar account confirms `gemini-3-pro-image`,
   `gemini-3.1-flash-image`, and `gemini-3.1-flash-lite-image` actually work
   on these endpoints at all — per conflict #12, TokenBazaar's own
   documentation table never named these 3 models under "Supported Image
   Models," unlike `gpt-image-2`. `gpt-image-2` satisfying gate 8 does not
   satisfy this gate for the 3 Gemini models, and vice versa — each of the 3
   needs its own live confirmation.
10. **(Embeddings, `text-embedding-3-large` only)** A live, approved
    TokenBazaar credential probe against `https://zx1.deepwl.net/v1/embeddings`
    confirms: (a) the path itself is correct (not some other path this
    delivery's inference got wrong); (b) the response envelope matches
    `{object:"list", data:[{object:"embedding", embedding:[...], index}],
    model, usage:{prompt_tokens, total_tokens}}`, not merely that *some*
    response comes back; and (c) `encoding_format`/`dimensions` behave as this
    delivery assumed. This gate is categorically different from gates 8–9
    (images): those confirm a *documented* contract actually works live; this
    one confirms the *contract itself* — see conflict-register item #11.
11. **(Responses, all 6 in-scope models: gpt-5.5, gpt-5.6-luna, gpt-5.6-sol,
    gpt-5.6-terra, deepseek-v4-flash, deepseek-v4-pro)** A live, approved
    TokenBazaar credential probe against `https://zx1.deepwl.net/v1/responses`
    confirms, end to end: (a) the SSE typed-event shape (`response.created`,
    `response.output_text.delta`, `response.completed`, etc.) actually matches
    what this delivery's shared `openAIResponseStreamState` parser expects —
    not merely that *some* stream comes back; (b) `reasoning.effort` actually
    changes model behavior (not just that the field passes through the
    request unmodified, which this delivery already confirmed by test); and
    (c) the three fail-closed rejections (`previous_response_id`, built-in
    tools, `background: true`) behave as documented against a real account —
    i.e. that TokenBazaar's own API would reject or misbehave on these fields
    if this delivery's pre-dispatch validation were ever bypassed, confirming
    the validation is not just defensive but necessary. This gate is
    categorically similar to gate 2 above (streaming/reasoning contracts) but
    scoped specifically to the 6 models this delivery actually wired, not the
    full 21-model chat-capable surface gate 2 was originally written against.
12. **(Messages, all 4 in-scope models: claude-fable-5, claude-haiku-4-5,
    claude-opus-5, claude-sonnet-5)** A live, approved TokenBazaar credential
    probe against `https://zx1.deepwl.net/v1/messages` confirms, end to end:
    (a) the SSE event shape (`message_start`, `content_block_delta`,
    `message_delta`, `message_stop`, etc.) actually matches what
    `streamAnthropicShapedResponse`'s `parseAnthropicSSEUsage`/
    `anthropicSSEIsError` expect — not merely that *some* stream comes back;
    (b) `thinking.budget_tokens` actually changes model behavior (not just
    that the field passes through the request unmodified, which this
    delivery already confirmed by test); (c) `max_tokens` is genuinely
    required by TokenBazaar's own gateway (i.e. that this delivery's
    pre-dispatch rejection of its absence is not just defensive but
    necessary); and (d) the tool-use and prompt-caching rejections in this
    delivery are not needlessly conservative — i.e. that TokenBazaar's
    `/v1/messages` does NOT actually support `tools`/`cache_control` today
    (if a live probe instead shows they work, this document's "unsupported"
    stance should be revisited and upgraded with that new evidence, per this
    delivery's own instruction, rather than left stale). This gate is
    categorically similar to gates 2 and 11 above (streaming/reasoning
    contracts) but scoped specifically to the 4 Claude models this delivery
    actually wired.
13. Since 2026-09-11 (see "Feature-flag removal" at the top of this
    document), none of the five capabilities has an independent code-level
    kill switch to "flip on" after the above gates are met — the prior
    version of this section had one such gate per capability
    (`SHAHRAH_TOKENBAZAAR_ENABLED`/`_IMAGES_ENABLED`/`_EMBEDDINGS_ENABLED`/
    `_RESPONSES_ENABLED`/`_MESSAGES_ENABLED`, each "deliberately flipped on
    by an operator"); those flags no longer exist. Until the
    model-discovery-catalog-to-dispatch connection described in that section
    is built, the actual operational safeguard for each capability going
    live is: no tenant is granted a TokenBazaar provider account/credential
    until every applicable gate above is met for that capability. Treat
    granting the first real TokenBazaar credential to a tenant as the
    equivalent go-live action this document previously described as
    "flipping the flag."

One model's readiness never unlocks another.

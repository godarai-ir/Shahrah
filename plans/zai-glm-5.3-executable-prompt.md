# Executable Prompt: Add GLM-5.3 to Z.ai Provider

## Ready-to-use prompt

```md
You are working inside the GodarAI/Shahrah repository as a senior Go/TypeScript platform engineer. Add exactly one new model, `glm-5.3`, to the existing direct Z.ai provider (`zai`). Reuse the existing zai provider, credentials, adapter, billing pipeline, and route registration built for `glm-5.2`, `glm-4.7-flashx`, `glm-5v-turbo`, `glm-5.3-flash`, and `glm-ocr`. Do not create a new provider, a new slug, or a parallel pipeline. Do not stop at a catalog-only YAML change — the model must be end-to-end usable, billed exactly, tested, documented, and independently gated before it is claimed production-ready.

## Source of truth

The Z.ai official documentation is the only source of truth for every provider-derived fact:

- Model guide (primary): https://docs.z.ai/guides/llm/glm-5.3
- Pricing (primary for rates): https://docs.z.ai/guides/overview/pricing
- Also re-verify against: https://docs.z.ai/openapi.json, https://docs.z.ai/guides/overview/overview, https://docs.z.ai/api-reference/llm/chat-completion, https://docs.z.ai/guides/capabilities/thinking-mode, https://docs.z.ai/guides/capabilities/function-calling, https://docs.z.ai/guides/capabilities/cache, https://docs.z.ai/guides/capabilities/streaming, https://docs.z.ai/guides/capabilities/stream-tool

Do not use model directories, aggregators, third-party gateways, or memory to fill gaps. Treat this prompt's "baseline facts" section as a draft fetched on 2026-09-11 that MUST be reconfirmed against current docs and, where ambiguous, a bounded live probe — not as settled truth.

## Baseline facts observed from the guide, OpenAPI schema, and pricing page — reverify before implementation

- Model ID: `glm-5.3` (the non-Flash, text-only sibling of `glm-5.3-flash`).
- Modalities: text-only input, text-only output. The model guide states GLM-5.3 uses the `ChatCompletionTextRequest` schema, unlike `glm-5.3-flash`/`glm-5v-turbo`, which use `ChatCompletionVisionRequest`. Do not copy the vision modality list from `glm-5.3-flash`.
- Context window: 1,000,000 tokens. Max output: 128K tokens. Confirm the exact numeric limits in the current OpenAPI schema rather than the prose guide.
- Thinking: forced/always-on, matching `glm-5.3-flash` — `thinking.type` only accepts `enabled`; `disabled` must be rejected. The capabilities/thinking-mode page states explicitly: "note that `GLM-5.3` and `GLM-5.3-FLASH` use forced thinking and cannot be disabled."
- **Reasoning effort — a materially narrower contract than every other zai chat model.** The model guide and the OpenAPI schema state that for `glm-5.3`, `reasoning_effort` accepts only `low`, `high`, and `max` (default `max`, recommended for coding). This is NOT the same 7-value set (`none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`) used by `glm-5.2`, `glm-4.7-flashx`, and `glm-5.3-flash`. Do not reuse that wider set for `glm-5.3` without independently re-confirming the schema still restricts it — if the enum has since widened, correct this prompt's baseline rather than assuming it is stale.
- **Conflict to resolve before implementation:** the general thinking-mode capabilities guide does not mention `reasoning_effort` for `glm-5.3` at all, while the model guide and OpenAPI schema for `glm-5.3` specifically both name the `low`/`high`/`max` enum. Treat the model-specific OpenAPI schema as authoritative over the general capabilities prose, but record this conflict in the final report.
- Because `glm-5.3`'s native `reasoning_effort` enum genuinely includes `low` as a distinct value (unlike the other zai chat models, where the platform's existing normalization silently upgrades a client's `low`/`medium` request to `high`), any shared reasoning-effort normalization code touched for this task must NOT collapse a valid `low` request for `glm-5.3` into `high`. Verify this against `services/silk-road/pkg/server/zai.go`'s `normalizeZAIReasoningEffort` and adjust it narrowly, by model ID, without changing behavior for the other four models.
- Function calling, tool streaming (`tool_stream`), structured/JSON output, and context caching are stated as supported — verify parameter names and behavior are identical to the existing `zai_runtime_profile.go` chat contract, or diverge explicitly where the OpenAPI schema says so.
- Streaming via `stream: true` is supported.
- Endpoints: the guide lists three protocol options for GLM-5.3 (OpenAI Chat Completion, OpenAI Response, Anthropic Message). Only the OpenAI chat-completions protocol against the existing zai `/chat/completions` route is in scope; the Response and Anthropic-Message protocols are out of scope for this task.
- Pricing (USD per 1M tokens, verified from https://docs.z.ai/guides/overview/pricing on 2026-09-11): input $1.40, cached input $0.26, cached-input storage "limited-time free" (promotional, no published end date — record as dated promotional provenance, not a permanent zero), output $4.40. This matches the existing `glm-5.2` rate exactly — confirm this is still the case (both models may simply share a rate card) rather than assuming it forever will be. No effective/observation date is shown on the pricing table itself; record the fetch date as the catalog `effective_from`/`verified_at` and re-verify before go-live in case the table changes.
- The architecture description (320B/18B MoE base shared with GLM-5.2, coding/cybersecurity benchmark improvements) is informational only and has no implementation impact.

## Required work

1. Read `AGENTS.md` (root and relevant subdirectories), `plans/zai-four-model-production-readiness.md`, `plans/zai-four-model-full-support-executable-prompt.md`, and `plans/zai-glm-5.3-flash-executable-prompt.md` in full before touching code. Run `git status --short` first and preserve any unrelated in-progress work.
2. Inspect the current zai implementation end to end before writing anything new:
   - `apps/model-discovery/providers/zai/*.yaml` (all five existing models, especially `glm-5.2.yaml` for the text/effort precedent and `glm-5.3-flash.yaml` for the always-enabled-thinking precedent);
   - `apps/model-discovery/src/catalog/runtime-profile.schema.ts` and the zai importer (`apps/model-discovery/scripts/model-import/providers/zai.ts`, `apps/model-discovery/src/catalog/zai-provider-importer.spec.ts`);
   - `services/control-plane/internal/proxy/zai_runtime_profile.go` and its tests (`zai_support_test.go`);
   - `services/control-plane/internal/billing/zai_pricing_test.go` and `catalog_pricing_snapshot.go`;
   - `services/control-plane/internal/server/routes_proxy.go` for how the existing chat models and `glm-ocr` are already registered — no new route should be needed since this is a chat-completions model, but confirm;
   - `services/silk-road/pkg/server/zai.go` and its tests (`zai_test.go`), in particular `normalizeZAIReasoningEffort`;
   - `services/control-plane/cmd/zai-readiness/main.go`;
   - `apps/tenant-admin/features/playground/domain.ts` and `domain.test.ts` for how model-specific parameter gating (e.g. effort selector, thinking toggle) is currently expressed per model, and how it currently gates `glm-5.3-flash`'s always-enabled thinking;
   - `apps/docs/docs/providers/zai.mdx`.
3. Fetch and record the current `docs.z.ai/openapi.json` checksum/observation date and the model guide's content at time of implementation. Note any conflict between the prose guide, the thinking-mode capabilities page, and the OpenAPI schema (e.g. the exact `reasoning_effort` enum) in a conflict register, following the same fail-closed pattern used for `glm-5v-turbo`'s tool-calling gate and `glm-ocr`'s page-limit gate in the existing plans.
4. Add `apps/model-discovery/providers/zai/glm-5.3.yaml` following the exact shape of the existing five model files: model ID, display name, lifecycle/status, `verified_at`, `api_format: zai`, mode, `supportedModes`, `modalities.input`/`output` (text only), `features`, `costs` (input_cost_per_token 0.0000014, cache_read_input_token_cost 0.00000026, output_cost_per_token 0.0000044), `sources` (all URLs consulted), and a full `runtimeProfile` block: `thinking.allowedModes` reflecting always-enabled thinking (mirroring `glm-5.3-flash`'s `alwaysEnabled: true` / `allowedModes: [adaptive]` shape), `thinking.effortValues` set to the model's real, narrower `[low, high, max]` enum (not the wider 7-value set used by other models), `sampling.parameters`, `tools.custom`/`parallel`/`structuredOutput`, `parameters.chat_completions` (exact allowed parameter list), `unsupportedParameters`, `limits` (contextWindow, maxInputTokens, maxOutputTokens).
5. Extend `apps/model-discovery/scripts/model-import/providers/zai.ts`: add `glm-5.3` to `ZAI_SUPPORTED_MODEL_IDS` and its pricing/limits expectation entry, so the importer's allowlist and fact-checking cover six models.
6. In `services/silk-road/pkg/server/zai.go`, add `glm-5.3` to the `zaiModels` route map, and adjust `normalizeZAIReasoningEffort` so a `low` reasoning-effort request is preserved verbatim for `glm-5.3` (a genuinely distinct, natively supported tier per the OpenAPI schema) instead of being collapsed into `high` the way it is for every other zai chat model. Do not change the collapse behavior for `glm-5.2`, `glm-4.7-flashx`, or `glm-5.3-flash`. Add targeted unit tests in `zai_test.go` proving: (a) `glm-5.3` dispatches correctly with `thinking.type: "enabled"`; (b) `reasoning_effort: "low"` for `glm-5.3` reaches the upstream request unchanged, while `reasoning_effort: "low"` for `glm-5.2` is still bumped to `"high"`; (c) `thinking.type: "disabled"` is rejected for `glm-5.3`.
7. Confirm the control-plane's generic `validateZAIChatRuntimeProfile`/`validateZAIThinkingRequest`/`AllowsEffort` machinery in `zai_runtime_profile.go` needs no code change for `glm-5.3` (only correct YAML data), since it already derives allowed thinking modes and effort values from the runtime profile. Add a unit test in `zai_support_test.go` proving a `glm-5.3`-shaped profile (effort values `[low, high, max]`) rejects `reasoning_effort: "medium"` while a `glm-5.2`-shaped profile (the wider 7-value set) still accepts it.
8. Confirm whether `glm-5.3` reuses the existing `/v1/chat/completions` route and model-routing candidates with no route changes (expected), and that FQN resolution `zai:<provider-account>:glm-5.3` works through the existing deployment/credential machinery without new code in `routes_proxy.go`.
9. Add/extend billing: a `TestZAIPricingUsesGLM53ConfirmedRates`-style test case in `zai_pricing_test.go` using the confirmed rates — input $0.0000014/token, cached input $0.00000026/token, output $0.0000044/token (exact per-token decimals, never float64). Assert cache-miss, cache-hit, and mixed-usage billing alongside the existing model cases.
10. Extend `services/control-plane/cmd/zai-readiness/main.go` with a bounded, approval-gated live check for `glm-5.3` (`thinking.type: "enabled"`, `reasoning_effort: "low"` to exercise the narrower enum, `max_tokens` minimal), following the existing spend-ceiling and secret-redaction guarantees. Also add a check (or extend an existing one) proving a `reasoning_effort: "medium"`/`"xhigh"` request is rejected for `glm-5.3` by the gateway before reaching Z.ai, if this can be done without spending — otherwise document why it stayed out of the live matrix.
11. Update `apps/docs/docs/providers/zai.mdx`: add the model to the models/route table, the pricing table (input $1.40 / cached $0.26 / output $4.40 per 1M tokens), the "گفتگو و استدلال" section noting both that thinking cannot be disabled for this model and that it only accepts `low`/`high`/`max` for `reasoning_effort` (unlike `glm-5.2`'s full set), and the intro/front-matter model count and list. Keep the same Persian technical-writing style as the rest of the file.
12. Update tenant-admin Playground (`domain.ts`/`domain.test.ts`) so `glm-5.3` is selectable, its mandatory-thinking constraint is enforced the same way as `glm-5.3-flash`'s (reject disabling thinking), and it is included among the models for which the `reasoningEffort` field is accepted. Add a regression test proving `glm-5.3` and `glm-5.3-flash` are gated independently (their `.endsWith` checks must not falsely match each other).
13. Add source-drift/importer tests proving the new YAML validates, and that the six-model allowlist, pricing facts, and runtime-profile facts are enforced, and that the existing five models remain unaffected.

## Explicitly out of scope

- Any other new Z.ai model.
- The OpenAI Response protocol and the Anthropic Message protocol endpoints for GLM-5.3.
- Coding Plan endpoints, quotas, or pricing.
- Changing the existing five models' contracts, pricing, or gates.
- Widening `glm-5.3`'s `reasoning_effort` enum to match the other models' 7-value set without a doc-verified schema change.
- Enabling this model as production-ready without a green live-readiness check, even though pricing is now confirmed.

## Validation

Use `pnpm`, never `npm`. Run model-discovery and tenant-admin lint/typecheck/tests/build, docs build, and from every touched Go module: `gofmt`, `goimports`, `go vet ./...`, `go test ./...`, `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`. Do not make live/billable Z.ai calls in CI.

## Definition of done

`glm-5.3` is importable, discoverable via `GET /v1/models`, callable through the existing `/v1/chat/completions` pipeline with FQN `zai:<provider-account>:glm-5.3`, billed at $1.40/$0.26/$4.40 per 1M tokens (input/cached/output), validated against its real (not copied) parameter contract — including mandatory thinking and its narrower `low`/`high`/`max` reasoning-effort enum — documented in Persian, selectable in Playground, and covered by unit tests. It stays gated `production-ready: false` (i.e., not claimed live-verified) until the live readiness check passes; nothing about the other five existing zai models regresses, and in particular the `low`→`high` reasoning-effort collapse used by the other chat models must remain unchanged for them.

## Required final report

1. Verified facts vs. this prompt's baseline (confirmed, corrected, or still unresolved), with source URLs and dates.
2. Files changed, by subsystem.
3. Confirmation that the shipped pricing matches $1.40/$0.26/$4.40 per 1M tokens or, if the pricing page has since changed, the corrected rate and its source/date.
4. Test/lint/build commands run and their results.
5. Any documentation conflicts found (the `reasoning_effort` enum discrepancy between the capabilities guide and the model-specific OpenAPI schema, chief among them) and how each was resolved or gated.
```

## Evidence used to create this prompt

Verified from official Z.ai sources on 2026-09-11:

- [GLM-5.3 guide](https://docs.z.ai/guides/llm/glm-5.3) — model ID, context/output limits, forced reasoning with `low`/`high`/`max` effort levels (default `max`), three API protocol options, text-only input.
- [Pricing](https://docs.z.ai/guides/overview/pricing) — GLM-5.3 input $1.40, cached $0.26, output $4.40 per 1M tokens (limited-time free cache storage), vs. GLM-5.3-Flash's $0.15/$0.03/$0.50.
- [Thinking mode capabilities guide](https://docs.z.ai/guides/capabilities/thinking-mode) — explicit statement that GLM-5.3 and GLM-5.3-Flash use forced thinking and cannot be disabled; no mention of `reasoning_effort` for GLM-5.3 (conflict with the model guide/OpenAPI schema, noted above).
- [OpenAPI schema](https://docs.z.ai/openapi.json) — GLM-5.3 uses `ChatCompletionTextRequest` (text-only, no vision schema); `reasoning_effort` enum for GLM-5.3 is restricted to `low`/`high`/`max`; `thinking.type` enum is `enabled`/`disabled` but only `enabled` is valid for this model, with depth controlled by `reasoning_effort`.

Shahrah-specific constraints were derived from the existing zai provider implementation: `apps/model-discovery/providers/zai/glm-5.3-flash.yaml`, `apps/model-discovery/scripts/model-import/providers/zai.ts`, `services/control-plane/internal/proxy/zai_runtime_profile.go`, `services/silk-road/pkg/server/zai.go`, `services/control-plane/internal/billing/zai_pricing_test.go`, `services/control-plane/cmd/zai-readiness/main.go`, `apps/tenant-admin/features/playground/domain.ts`, `apps/docs/docs/providers/zai.mdx`, and `plans/zai-glm-5.3-flash-executable-prompt.md`.

# Executable Prompt: Add GLM-5.3-Flash to Z.ai Provider

## Ready-to-use prompt

```md
You are working inside the GodarAI/Shahrah repository as a senior Go platform engineer. Add exactly one new model, `glm-5.3-flash`, to the existing direct Z.ai provider (`zai`). Reuse the existing zai provider, credentials, adapter, billing pipeline, and route registration built for `glm-5.2`, `glm-4.7-flashx`, `glm-5v-turbo`, and `glm-ocr`. Do not create a new provider, a new slug, or a parallel pipeline. Do not stop at a catalog-only YAML change — the model must be end-to-end usable, billed exactly, tested, documented, and independently gated before it is claimed production-ready.

## Source of truth

The Z.ai official documentation is the only source of truth for every provider-derived fact:

- Model guide (primary): https://docs.z.ai/guides/vlm/glm-5.3-flash
- Pricing (primary for rates): https://docs.z.ai/guides/overview/pricing
- Also re-verify against: https://docs.z.ai/openapi.json, https://docs.z.ai/guides/overview/overview, https://docs.z.ai/api-reference/llm/chat-completion, https://docs.z.ai/guides/capabilities/thinking-mode, https://docs.z.ai/guides/capabilities/function-calling, https://docs.z.ai/guides/capabilities/cache, https://docs.z.ai/guides/capabilities/streaming, https://docs.z.ai/guides/capabilities/stream-tool

Do not use model directories, aggregators, third-party gateways, or memory to fill gaps. Treat this prompt's "baseline facts" section as a draft fetched on 2026-09-11 that MUST be reconfirmed against current docs and, where ambiguous, a bounded live probe — not as settled truth.

## Baseline facts observed from the guide and pricing page — reverify before implementation

- Model ID: `glm-5.3-flash`.
- Input modalities: text, image, video, file — the guide's exact phrase is "Video / Image / Text / File"; resolve this against the current OpenAPI schema and the existing `glm-5v-turbo` contract before assuming identical format/size/count limits. Do not silently copy `glm-5v-turbo` limits without independent verification for this model.
- Output modality: text only.
- Context window: 1,000,000 tokens. Max output: 128K tokens. Confirm the exact numeric limits in the current OpenAPI schema rather than the prose guide.
- Thinking: the guide states `thinking.type` supports only `enabled` for this model — thinking CANNOT be disabled, unlike `glm-5.2`/`glm-4.7-flashx`. This is a materially different contract from the existing two text models; do not reuse their `disabled` code path for this model. Also carries `thinking.clear_thinking`, following the same preserved-reasoning contract as the other GLM models.
- Recommended/documented parameters: `temperature: 1`, `top_p: 0.95`, `reasoning_effort: max` — confirm whether these are defaults, recommendations, or hard requirements, and whether `reasoning_effort` follows the same effort-value set and any compatibility-mapping behavior established for `glm-5.2` (`none`/`minimal`/`low`/`medium`/`high`/`xhigh`/`max`), or a different one.
- Function calling, tool streaming (`tool_stream`), structured/JSON output, and context caching are stated as supported — verify parameter names and behavior are identical to the existing `zai_runtime_profile.go` chat contract, or diverge explicitly where the OpenAPI schema says so.
- Streaming via `stream: true` is supported.
- The guide mentions GLM Coding Plan quota and off-peak discounting — these are subscription/coding-plan concepts, explicitly out of scope. Do not use or reference the Coding Plan host or pricing for this general-API model.
- Pricing (USD per 1M tokens, verified from https://docs.z.ai/guides/overview/pricing on 2026-09-11): input $0.15, cached input $0.03, cached-input storage "limited-time free" (promotional, no published end date — record as dated promotional provenance, not a permanent zero), output $0.50. For comparison, the non-Flash `glm-5.3` is input $1.40 / cached $0.26 / output $4.40 (out of scope for this task — do not add it). No effective/observation date is shown on the pricing table itself; record the fetch date as the catalog `effective_from`/`verified_at` and re-verify before go-live in case the table changes.
- The architecture description (320B/18B MoE, hybrid attention, 45 layers) is informational only and has no implementation impact.

## Required work

1. Read `AGENTS.md` (root and relevant subdirectories), `plans/zai-four-model-production-readiness.md`, and `plans/zai-four-model-full-support-executable-prompt.md` in full before touching code. Run `git status --short` first and preserve any unrelated in-progress work.
2. Inspect the current zai implementation end to end before writing anything new:
   - `apps/model-discovery/providers/zai/*.yaml` (all four existing models, especially `glm-5v-turbo.yaml` for multimodal-input precedent and `glm-5.2.yaml` for the text/effort precedent);
   - `apps/model-discovery/src/catalog/runtime-profile.schema.ts` and the zai importer (`apps/model-discovery/scripts/model-import/providers/zai.ts`, `apps/model-discovery/src/catalog/zai-provider-importer.spec.ts`);
   - `services/control-plane/internal/proxy/zai_runtime_profile.go` and its tests (`zai_support_test.go`);
   - `services/control-plane/internal/billing/zai_pricing_test.go` and `catalog_pricing_snapshot.go`;
   - `services/control-plane/internal/server/routes_proxy.go` for how the three chat models and `glm-ocr` are already registered — no new route should be needed since this is a chat-completions model, but confirm;
   - `services/silk-road/pkg/server/zai.go` and `adapters.go`;
   - `services/control-plane/cmd/zai-readiness/main.go`;
   - `apps/tenant-admin/features/playground/domain.ts` and `PlaygroundColumn.tsx` for how model-specific parameter gating (e.g. effort selector, thinking toggle) is currently expressed per model;
   - `apps/docs/docs/providers/zai.mdx`.
3. Fetch and record the current `docs.z.ai/openapi.json` checksum/observation date and the model guide's content at time of implementation. Note any conflict between the prose guide and the OpenAPI schema (e.g. exact input modality list, parameter set, effort values, thinking flexibility) in a conflict register, following the same fail-closed pattern used for `glm-5v-turbo`'s tool-calling gate and `glm-ocr`'s page-limit gate in the existing plan.
4. Add `apps/model-discovery/providers/zai/glm-5.3-flash.yaml` following the exact shape of the existing four model files: model ID, display name, lifecycle/status, `verified_at`, `api_format: zai`, mode, `supportedModes`, `modalities.input`/`output`, `features`, `costs` (input_cost_per_token 0.00000015, cached_input_cost_per_token 0.00000003 if the schema supports a cached dimension, output_cost_per_token 0.0000005 — confirm the exact YAML field names against `glm-5.2.yaml`'s cached-input representation), `sources` (all URLs consulted), and a full `runtimeProfile` block: `thinking.allowedModes` must reflect that only `enabled` is valid for this model (not `[disabled, enabled]`), `sampling.parameters`, `tools.custom`/`parallel`/`structuredOutput`, `parameters.chat_completions` (exact allowed parameter list), `unsupportedParameters`, `limits` (contextWindow, maxInputTokens, maxOutputTokens, fileBytes if applicable per modality).
5. Extend `zai_runtime_profile.go` validation (`validateZAIChatRuntimeProfile` and helpers) only where `glm-5.3-flash` genuinely diverges from the current text/vision contract — specifically the mandatory-thinking-enabled behavior. Do not weaken existing validation for the other four models. Add a targeted unit test proving `thinking.type: "disabled"` is rejected for this model while still accepted for `glm-5.2`/`glm-4.7-flashx`.
6. Confirm whether `glm-5.3-flash` reuses the existing `/v1/chat/completions` route and model-routing candidates with no route changes (expected), and that FQN resolution `zai:<provider-account>:glm-5.3-flash` works through the existing deployment/credential machinery without new code in `routes_proxy.go`.
7. Add/extend billing: a pricing snapshot entry and `zai_pricing_test.go` case using the confirmed rates — input $0.00000015/token, cached input $0.00000003/token, output $0.0000005/token (convert to exact per-token decimals, never float64). Assert cache-miss, cache-hit, and mixed-usage billing in a dedicated test alongside the existing `glm-5.2`/`glm-4.7-flashx`/`glm-5v-turbo`/`glm-ocr` cases. Since the rate is now confirmed, this model does not need to ship pricing-gated — only the live-readiness and parameter-verification gates in step 8 remain blocking for production-ready status.
8. Extend `services/control-plane/cmd/zai-readiness/main.go` with a bounded, approval-gated live check for `glm-5.3-flash` (thinking always enabled, `max_tokens` minimal, similar to the existing `chat_glm_4_7_flashx`/`tools_glm_5_2` checks), following the existing spend-ceiling and secret-redaction guarantees.
9. Update `apps/docs/docs/providers/zai.mdx`: add the model to the models/route table, the pricing table (input $0.15 / cached $0.03 / output $0.50 per 1M tokens), the "گفتگو و استدلال" section noting thinking cannot be disabled for this model, and any multimodal notes consistent with its verified input modalities. Keep the same Persian technical-writing style as the rest of the file.
10. Update tenant-admin Playground (`domain.ts`/`PlaygroundColumn.tsx`) only as needed so the model is selectable and its mandatory-thinking / effort / multimodal constraints are reflected in the UI (e.g. disable or hide any "disable thinking" control for this model).
11. Add source-drift/importer tests proving the new YAML validates, and that Z.ai/xAI slug isolation and the existing four models remain unaffected.

## Explicitly out of scope

- `glm-5.3` (non-Flash) or any other new Z.ai model.
- Coding Plan endpoints, quotas, or pricing.
- Changing the existing four models' contracts, pricing, or gates.
- Enabling this model as production-ready without a green live-readiness check, even though pricing is now confirmed.

## Validation

Use `pnpm`, never `npm`. Run model-discovery and tenant-admin lint/typecheck/tests/build, docs build, and from every touched Go module: `gofmt`, `goimports`, `go vet ./...`, `go test ./...`, `golangci-lint run --timeout=5m --max-issues-per-linter=0 --max-same-issues=0`. Do not make live/billable Z.ai calls in CI.

## Definition of done

`glm-5.3-flash` is importable, discoverable via `GET /v1/models`, callable through the existing `/v1/chat/completions` pipeline with FQN `zai:<provider-account>:glm-5.3-flash`, billed at $0.15/$0.03/$0.50 per 1M tokens (input/cached/output), validated against its real (not copied) parameter/modality contract — including mandatory thinking — documented in Persian, selectable in Playground, and covered by unit tests. It stays gated `production-ready: false` until the live readiness check passes; nothing about the other four existing zai models regresses.

## Required final report

1. Verified facts vs. this prompt's baseline (confirmed, corrected, or still unresolved), with source URLs and dates.
2. Files changed, by subsystem.
3. Confirmation that the shipped pricing matches $0.15/$0.03/$0.50 per 1M tokens or, if the pricing page has since changed, the corrected rate and its source/date.
4. Test/lint/build commands run and their results.
5. Any documentation conflicts found (input-modality wording, effort values, thinking flexibility) and how each was resolved or gated.
```

## Evidence used to create this prompt

Verified from official Z.ai sources on 2026-09-11:

- [GLM-5.3-Flash guide](https://docs.z.ai/guides/vlm/glm-5.3-flash)
- [Pricing](https://docs.z.ai/guides/overview/pricing)

Shahrah-specific constraints were derived from the existing zai provider implementation: `apps/model-discovery/providers/zai/glm-ocr.yaml`, `services/control-plane/internal/proxy/zai_layout_parsing_handler.go`, `services/control-plane/internal/proxy/zai_runtime_profile.go`, `services/control-plane/cmd/zai-readiness/main.go`, `services/control-plane/internal/config/provider_features.go`, `services/control-plane/internal/server/routes_proxy.go`, `apps/docs/docs/providers/zai.mdx`, and `plans/zai-four-model-full-support-executable-prompt.md`.

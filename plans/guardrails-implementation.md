# Implementation Notes: Guardrails MVP

## Objective

Ship a scoped LLM guardrails MVP for Shahrah's Go control-plane gateway. This implementation follows TrueFoundry's useful concepts of input/output hooks, validate versus mutate operations, and enforce/audit-style outcomes, but keeps Shahrah's first cut local, tenant-safe, and OpenAI-compatible.

## Inputs Reviewed

- TrueFoundry docs: `.truefoundry-docs/markdown/docs__ai-gateway__guardrails-overview.md`
- TrueFoundry docs: `.truefoundry-docs/markdown/docs__ai-gateway__guardrails-configuration.md`
- TrueFoundry docs: `.truefoundry-docs/markdown/docs__ai-gateway__guardrails-getting-started.md`
- GodarAI plans: `plans/readme.md`
- GodarAI plans: `plans/09_shahrah_features.md`
- GodarAI code: `services/control-plane/internal/proxy`
- GodarAI code: `services/control-plane/internal/observability`
- GodarAI schema: `packages/database/schema.prisma`

## What Shipped

- Per-request LLM input and output guardrails for `/v1/chat/completions`.
- Header support for `X-Shahrah-Guardrails` and `X-Shahrah-Guardrails-Scope`.
- Compatibility aliases for `X-TFY-GUARDRAILS`, `X-TFY-GUARDRAILS-SCOPE`, `input_guardrails`, and `output_guardrails`.
- Built-in local guardrails:
  - `keyword_blocklist`
  - `regex`
  - `secrets_detection`
  - `basic_pii` with aliases `pii_redaction` and `pii_detection`
- Input mutation before routing/provider calls.
- Input validation blocking before provider calls.
- Output mutation after provider response and wallet debit.
- Output validation/blocking after provider response and wallet debit.
- OpenAI-style guardrail error envelope with `error.type = guardrail_checks_failed`.
- Trace spans for guardrail checks with selector, hook, status, operation, enforcement, latency, and mutation metadata.
- Tenant-scoped guardrail persistence in `tenant_guardrails`.
- Control-plane CRUD APIs under `/api/v1/guardrails`.
- Tenant-admin management UI under `/guardrails` in the Governance navigation group.
- Stable tenant selectors in the form `guardrails/<slug>` so future Prompt Management can reference guardrails without owning their configuration.
- Runtime resolution of enabled tenant guardrails by selector slug, with built-in selectors still available as fallback.

## Intentionally Deferred

- Gateway-level automatic policy attachment. The current MVP applies guardrails when callers or future prompt records provide selectors.
- Prompt Management integration. The prompt feature is not implemented yet; it should store guardrail selectors, not duplicate guardrail configuration.
- MCP tool pre/post guardrails.
- OPA, Cedar, and third-party provider guardrails.
- Async input validation with provider cancellation.
- Streaming output guardrail buffering. MVP rejects streaming requests when output guardrails are selected.

## Operational Risks

- Built-in detectors are pragmatic pattern checks, not compliance-grade DLP.
- Output guardrails run after provider response. Model cost is already incurred and GodarAI debits the wallet before blocking or redacting output.
- Tenant guardrail selectors resolve within the authenticated tenant context. A matching slug in another tenant is not loaded.

## Rollout Guidance

- Start with request headers in staging and traces enabled.
- Prefer input mutation guardrails for PII/secrets where possible to avoid sending sensitive content to providers.
- Use output guardrails for high-risk tenants only until streaming buffering and configurable policies are implemented.
- Let Prompt Management persist selected guardrail selectors once that feature lands.

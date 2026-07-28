# Implementation Plan: Unified Provider Credential Validation

**Status:** Implemented and verified on 2026-07-28.

## Objective

Replace the two divergent provider-credential validators with one provider-aware,
non-billable validation engine shared by tenant BYOK onboarding and
platform-managed credentials. The change must make canonical provider slugs,
authentication headers, endpoint construction, base-URL safety, result
classification, and UI preflight behavior consistent with the connection-schema
catalog.

## Inputs Reviewed

- LiteLLM docs:
  - `.llmLite_docs/providers/gemini.md`
  - `.llmLite_docs/proxy/ui_credentials.md`
- LiteLLM source: not inspected; Shahrah's duplicated validators and provider
  contracts were sufficient to identify the defect.
- Shahrah plans:
  - `plans/README.md`
  - `plans/platform-managed-provider-credentials-provider-matrix.md`
  - `plans/platform-managed-provider-credentials-product-analysis.md`
  - `plans/platform-managed-provider-credentials-implementation-prompt.md`
  - `plans/provider-account-models-ux-flow.md`
  - `plans/phase2-liteLLM-provider-support.md`
- Shahrah implementation:
  - tenant credential validation under `services/control-plane/internal/storage`
  - platform credential probing under
    `services/control-plane/internal/platformcredentials`
  - provider connection schemas under `apps/model-discovery`
  - tenant onboarding under `apps/tenant-admin`

## Root Cause

- The tenant validator was introduced on 2026-06-20 as a small Wizard feature
  with a hard-coded set of the then-prioritized OpenAI-compatible providers.
- SambaNova was added later the same day as another one-off switch case.
- The database-derived 21-provider schema catalog arrived on 2026-07-17.
- A substantially more complete platform-managed credential prober arrived on
  2026-07-18, but the tenant Wizard and saved-account test path continued using
  the June validator.
- The July 24 Wizard preflight made successful direct validation mandatory for
  API-key onboarding, exposing the stale allow-list as a blocking product error.
- No contract test compared `runtimeReady` schema rows with implemented
  non-billable probes.

## Architecture

Create `services/control-plane/internal/providervalidation` as the single
control-plane validation engine.

It owns:

- canonical provider slug normalization;
- the `(provider, auth_mode)` validation registry;
- provider-specific, non-billable request construction;
- exact authentication headers and optional provider headers;
- canonical connection-field aliases (`base_url`, `api_base`, `endpoint`,
  `serving_endpoint_url`);
- provider-host allow rules and SSRF-safe base URL validation;
- redirect suppression so a validated upstream cannot redirect probes to an
  unapproved network target;
- timeout, response-body limits, secret redaction, and stable result categories;
- a discoverable `Supports(provider, authMode)` capability used to enrich
  connection schemas.

Both tenant storage and `platformcredentials.HTTPProber` adapt their native
records to this package. Neither path keeps its own provider switch.

## Work Breakdown

### Phase 1: Foundation

- [x] Add the shared provider validation package and canonical contract registry.
- [x] Move base-URL validation and provider host rules behind the shared package.
- [x] Preserve the existing exported storage API as a compatibility wrapper.
- [x] Include all current `runtimeReady` schema contracts:
  Anthropic, Azure OpenAI, Cerebras, Cohere, Databricks PAT, DeepInfra,
  Google Gemini, Groq, Mistral AI, OpenAI, OpenRouter, Together AI, and xAI.

### Phase 2: Core Behavior

- [x] Replace the tenant API-key-only validator with the shared credential
  validator, including arbitrary schema credential fields such as
  `access_token`.
- [x] Replace the platform-managed credential probe implementation with an
  adapter over the same shared engine.
- [x] Use the authenticated OpenRouter `/api/v1/key` probe instead of the public
  models catalog.
- [x] Normalize Gemini API-version composition so a `base_url` already ending
  in `v1beta` is not duplicated.
- [x] Normalize Azure legacy and `/openai/v1` model-list endpoints.
- [x] Treat authenticated quota/billing/rate-limit responses as a valid
  credential with a non-OK category.
- [x] Prevent provider-controlled redirects from escaping the approved host
  policy.

### Phase 3: API and UX Consistency

- [x] Add a derived `credential_validation_supported` flag to connection-schema
  responses.
- [x] Require a successful preflight for every supported auth mode, not only the
  literal `api_key` mode.
- [x] Return unsupported validation as a warning and never mislabel it as an
  upstream connection failure.
- [x] Use credential-neutral Persian and English copy so API keys, PATs, and
  future token types share one UX.

### Phase 4: Drift Prevention and Hardening

- [x] Add table-driven request/header/URL/status tests for every supported
  contract.
- [x] Add a contract test comparing the Go registry with every
  `runtimeReady=true` schema row across model-discovery migrations.
- [x] Add tenant-store tests for Gemini, Mistral's canonical slug, Databricks
  PAT, unsupported auth, custom `base_url`, and secret redaction.
- [x] Run Go formatting/tests plus tenant-admin unit tests, typecheck, lint, and
  production build.

## API Contract

Existing endpoints remain compatible:

- `POST /api/v1/provider-connections/test`
- `POST /api/v1/tenant-provider-accounts/{accountID}/test-credentials`

The response preserves the existing fields and adds a stable category:

```json
{
  "success": true,
  "unsupported": false,
  "category": "ok",
  "message": "Provider accepted the credential."
}
```

Stable categories include:

- `ok`
- `quota_exhausted`
- `invalid_credential`
- `forbidden`
- `configuration`
- `missing_credential`
- `network`
- `timeout`
- `provider`
- `unsupported`

Connection-schema responses add:

```json
{
  "runtime_ready": true,
  "credential_validation_supported": true
}
```

## Data Model

No persisted schema or tenant-account migration is required.
`credential_validation_supported` is derived from the shared registry when the
control plane reads model-discovery schemas. Credential values remain write-only
and are not added to logs, traces, or API responses.

## Testing Strategy

- Unit-test every provider contract with an in-memory HTTP transport.
- Assert HTTP method, canonical URL, query, authentication header, and relevant
  provider headers.
- Assert no network I/O for unsupported contracts or invalid base URLs.
- Assert redirects are surfaced without issuing a request to their target.
- Assert secrets are redacted from provider error details.
- Assert all `runtimeReady` schema contracts across current and future migration
  files exactly match the shared registry.
- Run the existing storage and platform-credential suites as regression tests.
- Run tenant-admin credential helper and route tests, followed by typecheck and
  lint.

## Rollout Plan

1. Ship control-plane and tenant-admin in the same release.
2. Confirm schema responses expose `credential_validation_supported`.
3. Smoke-test one API-key provider, Gemini, Azure OpenAI, and Databricks PAT.
4. Monitor validation result categories; never log raw credentials.
5. Roll back the application release if provider probes regress. No database
   rollback is needed.

## Open Questions

- Providers marked `runtimeReady=false` remain configurable but their automatic
  validation returns `unsupported`; enabling their runtime remains a separate
  provider-adapter decision.
- Customer-hosted SambaStack cannot use a fixed hostname allow-list safely. It
  remains unsupported by automatic validation until a dedicated outbound SSRF
  policy is implemented.

## Verification Results

- `go test ./...` under `services/control-plane`: passed.
- Shared registry, URL safety, catalog-alignment, platform credential, and
  tenant storage suites: passed.
- Tenant-admin targeted credential helper: 4/4 passed.
- Tenant-admin i18n validation: passed.
- Tenant-admin ESLint: passed.
- Tenant-admin full Vitest: 362/363 passed. The sole unrelated pre-existing
  failure is the arbitrary-font-size design-system guardrail against
  `app/components/admin-shell-client.tsx`.
- Tenant-admin production compilation: passed; its TypeScript phase stopped on
  the pre-existing duplicate React type incompatibility in
  `components/ui/command.tsx`. A direct typecheck also reported the same
  unrelated UI dependency mismatch plus existing test-only typing issues.

# Execution Prompt: LibreChat × GodarAI Credits Integration

Use this prompt with a coding agent working in the LibreChat fork at
`D:\source\liberchat\LibreChat` (remote `Delpak/LibreChat`, branch `dev`).

**Dependency:** this prompt starts only after the Shahrah side has shipped these items
(see `D:\source\Shahrah\plans\credit-pricing-execution-prompt.md`):

| LibreChat phase | Needs from Shahrah |
| --- | --- |
| A | Shahrah Phase 1: `x-godarai-credits-used`, `credits` in `godarai.telemetry`, `GET /v1/rate-card`, service key + `x-godarai-on-behalf-of` |
| B | Shahrah Phase 3: `GET /v1/budget/me`, 402 `budget_exceeded`, `x-godarai-model-substituted` |

---

You are a senior engineer integrating LibreChat with GodarAI's credit system. The core
rule: **GodarAI is the only authority for pricing and budget enforcement.** LibreChat
records and displays the credits GodarAI returns. It never computes prices, applies
exchange rates, or blocks on its own balance when the feature is on.

## 1. Read First

- `AGENTS.md` in the LibreChat repo (coding, test, and i18n rules).
- The Shahrah PRD section on LibreChat integration:
  `D:\source\Shahrah\docs\PRD چارچوب قیمتگذاری کردیتی گدارAI.md`.
- The Shahrah API contracts as shipped. Read the Shahrah docs pages for `/v1/rate-card`,
  `/v1/budget/me`, and the request headers. Do not re-derive them from the PRD.

## 2. Known Code Locations (verify)

- Transactions and usage: `packages/data-schemas/src/methods/{transaction,spendTokens}.ts`,
  `packages/data-schemas/src/models/{transaction,balance}.ts`, `packages/api/src/agents/{usage,transactions}.ts`
- Balance middleware and UI: `packages/api/src/middleware/{balance,checkBalance}.ts`,
  `client/src/components/Balance/*`, `client/src/hooks/useBalanceSummary.ts`,
  `client/src/components/Messages/Content/Error/BalanceError.tsx`
- Config schema and error types: `packages/data-provider/src/config.ts`
  (`baseEndpointSchema` ~L749, `endpointSchema` ~L2026, `enum ErrorTypes` ~L4779)
- Locales: `client/src/locales/{en,fa}/translation.json`; digit helpers in
  `client/src/locales/digits.ts`

## 3. Invariants

1. **Opt-in per endpoint.** Add `costSource: z.enum(['librechat', 'gateway']).default('librechat')`
   to the endpoint config schema. With the default, behavior is byte-for-byte what it is
   today (tests prove it).
2. With `costSource: gateway`, the deployment uses `balance.enabled: false` and
   `transactions.enabled: true`. Validate this at startup and log a clear warning if it is
   misconfigured.
3. Never surface GodarAI's raw error text to users. Map the stable codes only.
4. Scale mapping happens once, in the recording module. LibreChat's
   `tokenValue = tokens × rate` has the same shape as GodarAI credits; only the scale
   differs. Document the factor in the module and test it.
5. Every recorded transaction carries the GodarAI `trace_id` and `messageId` in its
   metadata so reconciliation can find it.
6. All new strings are in `en` and `fa`. Numbers use Persian digits and thousands
   separators in `fa`, with two decimals.

## 4. Phase A: Identity, Recording, Rate Card

1. **Identity:** document and ship the `librechat.yaml` example for the GodarAI endpoint.
   - Authenticate with a GodarAI **service key**.
   - Send `headers: { x-godarai-on-behalf-of: "{{LIBRECHAT_USER_OPENIDID}}", x-godarai-include-stream-telemetry: "true" }`.
   - Users sign in with GodarAI OIDC.
   - Confirm `OPENIDID` equals the GodarAI OIDC `sub`. If it does not, **stop and ask**.
2. **Recording module** (`packages/api`, used when `costSource: gateway`):
   - Non-stream: read `x-godarai-credits-used`.
   - Stream: read `credits` from the `godarai.telemetry` event.
   - Write one Transaction per response with the GodarAI credits (scaled), `trace_id`,
     `model`, and `messageId`. Skip LibreChat's own token pricing for that response.
   - Idempotent on `trace_id`, so retries or duplicate events do not double-record.
   - If no credits arrive (abort, disconnect), write a placeholder transaction marked
     `pending_reconciliation`.
3. **Rate card in the model selector:**
   - Fetch `GET /v1/rate-card` server-side, cached and respecting its `ETag`, and
     expose it to the client.
   - Show `typical_message_credits` next to each model ("about 4.50 credits per
     message").
   - Hide the hint when the model is missing from the rate card.

## 5. Phase B: Budget Display, Exceed Handling, Reconciliation

1. **Balance display:** replace the balance widget, when `costSource: gateway` is set,
   with data from `GET /v1/budget/me`, called server-side with the service key and
   on-behalf-of.
   - Use React Query, with loading, error, and empty states.
   - Show the most restrictive remaining rule, with the others in a popover or list.
   - Refresh after each completed response.
2. **Exceed handling:**
   - 402 with `code: "budget_exceeded"` → a new `ErrorTypes.BUDGET_EXCEEDED` with the
     translation key `com_error_budget_exceeded`, including the scope (personal, team,
     organization) in the message.
   - `x-godarai-model-substituted` (or the same field in telemetry) → a non-blocking
     notice on the message: "Answered with model X because the budget is exhausted".
3. **Reconciliation job:** periodic, configurable interval.
   - Find `pending_reconciliation` transactions.
   - Query the GodarAI log API filtered by `metadata.messageId` / `trace_id`.
   - Write the corrected transaction. Do not edit the original; follow LibreChat's
     existing transaction conventions.
   - Bounded retries, then mark the transaction `unreconciled` and log it.
   - Idempotent and safe to run on multiple instances (use a lock).

## 6. Tests

- Unit: scale mapping, header and telemetry parsing (including malformed and missing
  values), idempotency on `trace_id`, error mapping, `costSource` default parity.
- Integration: a mock GodarAI server for stream and non-stream paths, abort mid-stream,
  402, substitution, and reconciliation.
- Client: model-selector hint, budget widget states, fa and en rendering.
- Run the repo's lint, typecheck, test, and build commands (see LibreChat `AGENTS.md`)
  before finishing each phase.

## 7. Stop and Ask

- `OPENIDID` ≠ GodarAI OIDC `sub`.
- The GodarAI log API does not support filtering by `metadata.messageId`.
- Any need to compute or estimate a price locally.

## 8. Report

At the end of each phase report what shipped, the config example, test evidence,
decisions and assumptions, and the open risks. Do not push or open a PR without approval.

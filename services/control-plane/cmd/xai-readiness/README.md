# xAI readiness runner

This command performs a bounded live smoke test through GodarAI for the four
catalogued xAI models. It writes only status codes and case names to its JSON
report; credentials and provider response bodies are never retained.

Required environment:

- `GODARAI_BASE_URL`
- `GODARAI_VIRTUAL_KEY`
- `GODARAI_XAI_PROVIDER_ACCOUNT`
- `GODARAI_XAI_MAX_SPEND_USD` (defaults to `0.25`, maximum `10`)

Run from `services/control-plane`:

```sh
go run ./cmd/xai-readiness -report xai-readiness-report.json
```

The command exits non-zero when a case fails or is blocked. `XAI_API_KEY` is
not accepted for these gateway probes and is never included in the report.

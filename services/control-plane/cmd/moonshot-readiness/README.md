# Moonshot readiness runner

This command sends a bounded smoke matrix through GodarAI for the four supported
Moonshot models. It accepts only a GodarAI virtual key; the Moonshot API key
stays in the configured provider account.

Required environment variables:

- `GODARAI_BASE_URL`
- `GODARAI_VIRTUAL_KEY`
- `GODARAI_MOONSHOT_PROVIDER_ACCOUNT`

Set `GODARAI_MOONSHOT_MAX_SPEND_USD` to an approved positive ceiling. The
default is USD 0.10. Run from `services/control-plane`:

```powershell
go run ./cmd/moonshot-readiness -report moonshot-readiness-report.json
```

The report contains request names, model IDs, HTTP status codes, and sanitized
errors. It never contains credentials or response bodies. A passing smoke report
does not set `production_ready`; release owners must also approve load, billing
reconciliation, privacy, observability, and rollback evidence.

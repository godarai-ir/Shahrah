# OpenAI contract readiness audit

This command compares the seven public GodarAI Responses REST operations with
the Responses subset of the official OpenAI OpenAPI baseline pinned to commit
`db14b6e1712aaf5265cf5a6871adff7a9c61d31c` (2026-07-28).
It is intentionally offline and never labels a model as live-validated.

Run from `services/control-plane`:

```powershell
go run ./cmd/openai-readiness `
  -baseline ../../packages/contracts/openapi/openai-scoped-routes-db14b6e.json `
  -routes internal/server/routes_proxy.go `
  -model-snapshot openai-model-discovery-snapshot.json `
  -report openai-readiness-report.json
```

The command exits non-zero while any Responses REST operation is absent.
`route_ready` and `contract_ready` cover this static REST-route comparison
only. The report deliberately leaves `websocket_ready`, `production_ready`,
and `live_validated` false until those separate gates are actually completed.
`-allow-incomplete` is only for generating an intermediate audit artifact; it
does not enable the provider release gate.

# OpenAI contract readiness audit

This command compares the seven public GodarAI Responses REST operations with
the Responses subset of the official OpenAI OpenAPI baseline pinned to commit
`db14b6e1712aaf5265cf5a6871adff7a9c61d31c` (2026-07-28). By default it also
downloads the current official Responses resource, create, input-items, and
streaming-event Markdown references and requires exact agreement with the
checked-in inventory of operations, top-level fields, output and input-item
variants, statuses, response envelopes, and SSE event types.

Run from `services/control-plane`:

```powershell
go run ./cmd/openai-readiness `
  -baseline ../../packages/contracts/openapi/openai-scoped-routes-db14b6e.json `
  -routes internal/server/routes_proxy.go `
  -model-snapshot openai-model-discovery-snapshot.json `
  -report openai-readiness-report.json
```

The command exits non-zero while any Responses REST operation is absent or
the official-documentation reconciliation is incomplete or drifted.
`route_ready` reports the static route comparison. `contract_ready` becomes
true only when the routes are complete, the snapshot-driven test surface is
100%, and the same run verifies the snapshot against current official OpenAI
documentation. Passing `-verify-official-docs=false` is an offline diagnostic
mode, reports snapshot coverage as zero, and deliberately leaves
`contract_ready=false`; combine it with
`-allow-incomplete` when an offline report is required. The report leaves
`websocket_ready`, `production_ready`, and `live_validated` false until those
separate gates are actually completed.

The snapshot-driven gateway tests run with the normal control-plane suite. The
official OpenAI Go SDK consumer suite is isolated from production dependencies
and must also pass:

```powershell
go -C tests/openai-sdk-contract test ./...
```
`-allow-incomplete` is only for generating an intermediate audit artifact; it
does not enable the provider release gate.

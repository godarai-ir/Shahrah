# tunnel-server AGENTS

## Purpose

Europe-side tunnel relay for GodarAI. This service forwards tunnel requests to `silk-road`.

## Ownership

This service owns:

- `/health` proxy health checks
- `/dispatch` forwarding to `silk-road`
- request metadata propagation for trace and request IDs
- streaming relay for SSE or streamed responses

It does not own provider logic, persistence, auth, or database access.

## Working Rules

1. Keep the handler thin and transport-focused.
2. Preserve `X-GodarAI-Trace-ID` and `X-Shahrah-Request-ID` when dispatching.
3. Do not buffer or transform streaming bodies unless the transport requires it.
4. Keep all behavior compatible with `platform.TunnelRequest`.
5. Avoid adding stateful features or side effects.

## Runtime

Main entrypoint: `cmd/server`

Environment variables:

- `SILK_ROAD_URL`: upstream silk-road base URL
- `TUNNEL_SERVER_ADDR`: local listen address

Default local ports:

- tunnel-server: `:8082`
- silk-road upstream: `http://127.0.0.1:8083`

## Commands

Run from the repository root:

```bash
go test ./services/tunnel-server/...
go run ./services/tunnel-server/cmd/server
go build ./services/tunnel-server/cmd/server
```

## Implementation Notes

- Streaming requests should keep the HTTP timeout disabled on the outbound client.
- If the upstream returns `text/event-stream`, relay the body with flush semantics.
- Preserve upstream response headers when relaying the response.

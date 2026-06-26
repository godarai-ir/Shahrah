# tunnel-client AGENTS

## Purpose

Iran-side tunnel relay for Shahrah. This service is a small HTTP hop that forwards requests to `tunnel-server`.

## Ownership

This service owns:

- `/health` proxy health checks
- `/dispatch` forwarding to the upstream tunnel server
- request metadata propagation for trace and request IDs
- streaming relay for SSE or streamed responses

It does not own provider logic, persistence, auth, or database access.

## Working Rules

1. Keep the handler thin and transport-focused.
2. Preserve `X-Shahrah-Trace-ID` and `X-Shahrah-Request-ID` when dispatching.
3. Do not buffer or transform streaming bodies unless the transport requires it.
4. Keep all behavior compatible with `platform.TunnelRequest`.
5. Avoid adding stateful features or side effects.

## Runtime

Main entrypoint: `cmd/server`

Environment variables:

- `TUNNEL_SERVER_URL`: upstream tunnel-server base URL
- `TUNNEL_CLIENT_ADDR`: local listen address

Default local ports:

- tunnel-client: `:8081`
- tunnel-server upstream: `http://127.0.0.1:8082`

## Commands

Run from the repository root:

```bash
go test ./services/tunnel-client/...
go run ./services/tunnel-client/cmd/server
go build ./services/tunnel-client/cmd/server
```

## Implementation Notes

- Streaming requests should keep the HTTP timeout disabled on the outbound client.
- If the upstream returns `text/event-stream`, relay the body with flush semantics.
- Preserve upstream response headers when relaying the response.

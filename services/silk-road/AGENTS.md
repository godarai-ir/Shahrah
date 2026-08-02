# silk-road AGENTS

## Purpose

Europe-side stateless provider proxy for GodarAI. This service receives tunnel dispatches and translates OpenAI-compatible payloads to upstream provider APIs.

## Ownership

This service owns:

- `/health`
- `/dispatch`
- provider registry and adapter selection
- request translation for OpenAI-compatible, Anthropic, and Gemini upstreams
- response adaptation back to OpenAI-compatible chat completions
- streaming conversion where supported

It does not own tunnel transport, persistence, auth, or database access.

## Supported Providers

The registry currently includes:

- OpenAI
- Groq
- Mistral and `mistral-ai`
- Together AI and `together-ai`
- Anthropic
- Gemini
- OpenAI-compatible fallbacks such as OpenRouter, Perplexity, Perplexity AI, SambaNova, and xAI

## Working Rules

1. Keep provider-specific code isolated in `pkg/server`.
2. Preserve OpenAI-compatible request and response shapes at the service boundary.
3. Do not introduce persistence or long-lived state.
4. Keep provider selection deterministic and explicit.
5. Avoid logging secrets, API keys, or raw user payloads.

## Runtime

Main entrypoint: `cmd/server`

Environment variables:

- `SILK_ROAD_ADDR`: local listen address

Default local port:

- silk-road: `:8083`

## Commands

Run from the repository root:

```bash
go test ./services/silk-road/...
go run ./services/silk-road/cmd/server
go build ./services/silk-road/cmd/server
```

## Implementation Notes

- OpenAI-compatible providers default to `POST /chat/completions` unless a request metadata override specifies another endpoint path.
- Anthropic requests use the provider API key and `anthropic-version` header, and streaming is adapted to OpenAI-compatible SSE chunks.
- Gemini streaming through the OpenAI-compatible chat-completions path is
  supported and must preserve upstream SSE semantics.
- If the upstream responds with an error status, relay that status and headers without inventing fallback responses.

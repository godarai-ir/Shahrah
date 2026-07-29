# Silk Road

Reserved for the Europe-side stateless provider proxy.

## Anthropic WIF configuration

- `ANTHROPIC_WIF_ENABLED`: independent rollback switch for Anthropic Workload Identity Federation; default `true`
- `ANTHROPIC_WIF_SOURCES_JSON`: JSON object that maps opaque token-source IDs to absolute, allowlisted OIDC assertion file paths

Tenant and platform connection payloads carry only a token-source ID. Silk Road reads the mapped assertion, exchanges it at Anthropic, and keeps the short-lived access token in a connection-scoped cache. Assertions and access tokens must not be logged or returned.

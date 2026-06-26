# Model Discovery Web AGENTS

Public Next.js catalog for browsing Shahrah models and providers.

## Ownership

This app owns public discovery and read-only catalog flows:

- model and provider browsing
- search and filtering
- model details
- provider details
- model comparison
- category pages
- local display preferences

It does not own human administration, tenant operations, mutations, or service credentials.

## Working Rules

1. The app is deployed under the `/catalog` base path.
2. Browser requests use the same-origin read-only proxy under `/catalog/api/v1/[...path]`.
3. Reject admin paths and keep the proxy allow-listed.
4. Do not add model-discovery service-account credentials here.
5. Run commands from the repository root with `pnpm --filter @shahrah/model-discovery-web ...`.

## API Contract

Before changing request or response handling, verify it against:

- `docs/model-discovery-api-contract.md`
- `apps/model-discovery/openapi/model-discovery.openapi.json`

## Local Run

```bash
pnpm --filter @shahrah/model-discovery-web dev
```

The local app listens on `http://localhost:3003/catalog`.
Server-to-server requests default to `http://localhost:3002/api/v1`.

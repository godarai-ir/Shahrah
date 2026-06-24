# Model Discovery Web

Public Next.js frontend for browsing, searching, and comparing the Shahrah model catalog.

## Ownership

This app owns public model and provider discovery, model details, comparisons, categories, facets, and user-local display preferences. It does not own human administration, tenant operations, service credentials, or catalog mutations.

The backend and source of truth is `apps/model-discovery`.

## Run

From the repository root:

```powershell
pnpm --filter @shahrah/model-discovery-web dev
```

The local app listens on `http://localhost:3003`; `support-admin` uses port `3001` and `model-discovery` uses port `3002`.

The development server uses `http://localhost:3002/api/v1` for server-side model-discovery requests by default. Browser requests use the same-origin read-only proxy under `/api/v1`.

Environment variables:

- `MODEL_DISCOVERY_API_BASE_INTERNAL`: server-to-server API base; default `http://localhost:3002/api/v1`
- `NEXT_PUBLIC_MODEL_DISCOVERY_API_BASE`: optional browser-visible API base; default `/api/v1`

No model-discovery service-account credential belongs in this application.

## Public routes

- `/`
- `/models`
- `/models/[providerSlug]/[...modelSlug]`
- `/providers`
- `/providers/[providerSlug]`
- `/compare`
- `/categories/[categorySlug]`

The same-origin `/api/v1/[...path]` route only proxies allow-listed public GET endpoints. Paths containing `admin` are rejected.

## API contract

Before changing request or response handling, verify it against:

- `docs/model-discovery-api-contract.md`
- `apps/model-discovery/openapi/model-discovery.openapi.json`

Legacy public URLs under `/catalog/*` should be redirected by the deployment ingress to this application while old links are retired.

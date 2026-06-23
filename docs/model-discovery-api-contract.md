# Model Discovery API Contract

This document is the contract entrypoint for every Shahrah component that consumes `model-discovery`.

## Canonical Spec Sources

Use these sources in this order:

1. Live Swagger UI:
   - `http://localhost:3002/docs`
2. Live machine-readable spec:
   - `http://localhost:3002/docs/openapi.json`
   - `http://localhost:3002/docs/openapi.yaml`
3. Repo-exported artifact:
   - [apps/model-discovery/openapi/model-discovery.openapi.json](/D:/source/Shahrah/apps/model-discovery/openapi/model-discovery.openapi.json)
   - [apps/model-discovery/openapi/model-discovery.openapi.yaml](/D:/source/Shahrah/apps/model-discovery/openapi/model-discovery.openapi.yaml)

The repo artifact is the stable review target for code changes and tests. The live docs are the fastest local inspection surface.

## Regenerate The Spec

From repo root:

```bash
pnpm --dir apps/model-discovery openapi:export
```

This writes the OpenAPI artifact under `apps/model-discovery/openapi/`.

## Change Workflow

When changing any `model-discovery` API route, method, query parameter, request body, or response shape:

1. Update the Nest controller/DTO behavior.
2. Regenerate the OpenAPI artifact.
3. Review the diff in `apps/model-discovery/openapi/`.
4. Update consuming code in `support-admin` and/or `control-plane`.
5. Update contract tests where the consumer depends on specific fields or constraints.

Do not merge API changes based on tribal knowledge or memory of the route shape.

## Support Admin Rules

`support-admin` is a direct client of `model-discovery`.

Public catalog requests currently go through:

- [apps/support-admin/lib/api.ts](/D:/source/Shahrah/apps/support-admin/lib/api.ts)

Admin requests go through Next route handlers that validate the local support-admin session and proxy with a scoped `model-discovery` service credential, for example:

- [apps/support-admin/app/api/admin/providers/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/providers/route.ts)
- [apps/support-admin/app/api/admin/models/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/models/route.ts)

Rules for future changes:

1. Before adding or changing a request, read the spec first.
2. Prefer the live `openapi.json` while implementing locally.
3. Use the repo artifact in PR review to verify that route/path/query/body expectations still match.
4. If `support-admin` starts consuming a new endpoint, add that endpoint to its README notes.
5. Do not guess response fields from UI usage; confirm them in the spec.
6. Never forward a human credential or browser session to `model-discovery`; use `MODEL_DISCOVERY_SERVICE_TOKEN`.

## Control Plane Rules

`control-plane` consumes `model-discovery` as an upstream dependency for provider catalog flows.

The current contract test reads:

- [services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go](/D:/source/Shahrah/services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go)

Rules for future changes:

1. `control-plane` must treat the exported OpenAPI artifact as the canonical review input.
2. When a `model-discovery` endpoint shape changes, update both the client code and the contract test assertions.
3. If `control-plane` starts calling a new `model-discovery` endpoint, extend the contract test to assert that route and its critical params/constraints exist in the spec.
4. Do not add hidden fallback behavior that bypasses the documented upstream contract.

## Minimum Review Checklist

Before merging changes that touch the `model-discovery` contract, confirm:

1. `pnpm --dir apps/model-discovery openapi:export`
2. `go test ./services/control-plane/internal/admin`
3. `support-admin` routes and clients still point to documented `model-discovery` endpoints

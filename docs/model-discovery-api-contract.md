# Model Discovery API Contract

This document is the contract entrypoint for every Shahrah component that consumes `model-discovery`.

## Pricing Contract

Pricing is unit-aware and is returned through `pricingStatus`,
`pricingProfiles`, and `pricingSummary`. A profile identifies a region,
currency, source, and effective interval; every profile contains dimensioned
rates:

```json
{
  "pricingStatus": "priced",
  "pricingProfiles": [
    {
      "region": "us-west-2",
      "currency": "USD",
      "rates": [
        {
          "key": "output:second:standard:resolution=540p:tier=base",
          "usage": "output",
          "unit": "second",
          "amount": "0.75",
          "quantity": 1,
          "serviceTier": "standard",
          "resolution": "540p",
          "tier": null,
          "qualifiers": {}
        }
      ]
    }
  ]
}
```

`amount` is always an exact decimal string. Token-like source prices are
normalized to `quantity: 1000000`; per-second, per-request, and per-image rates
use `quantity: 1`. `pricingStatus=unknown` is distinct from an explicit
`pricingStatus=free`; a missing or zero-only source price is never inferred to
be free.

Numeric price filters and sorting require both `price_usage` and `price_unit`.
Optional selectors are `price_region`, `price_resolution`, and
`price_service_tier`. Prices with different units must not be compared.

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
4. Update consuming code in `model-discovery-web`, `support-admin`, and/or `control-plane`.
5. Update contract tests where the consumer depends on specific fields or constraints.

Do not merge API changes based on tribal knowledge or memory of the route shape.

## Model Discovery Web Rules

`model-discovery-web` owns the public read-only catalog client:

- [apps/model-discovery-web/lib/api.ts](/D:/source/Shahrah/apps/model-discovery-web/lib/api.ts)
- [apps/model-discovery-web/app/api/v1/[...path]/route.ts](/D:/source/Shahrah/apps/model-discovery-web/app/api/v1/[...path]/route.ts)

The same-origin proxy only forwards allow-listed public GET endpoints. It must never receive or forward `MODEL_DISCOVERY_SERVICE_TOKEN`.

## Support Admin Rules

`support-admin` is a direct client of `model-discovery`.

Admin requests go through Next route handlers that validate the local support-admin session and proxy with a scoped `model-discovery` service credential, for example:

- [apps/support-admin/app/api/admin/providers/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/providers/route.ts)
- [apps/support-admin/app/api/admin/models/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/models/route.ts)

Rules for future changes:

1. Before adding or changing a request, read the spec first.
2. Prefer the live `openapi.json` while implementing locally.
3. Use the repo artifact in PR review to verify that route/path/query/body expectations still match.
4. If either web application starts consuming a new endpoint, add that endpoint to the owning app's README notes.
5. Do not guess response fields from UI usage; confirm them in the spec.
6. Never forward a human credential or browser session to `model-discovery`; use `MODEL_DISCOVERY_SERVICE_TOKEN`.

## Control Plane Rules

`control-plane` consumes `model-discovery` as an upstream dependency for provider catalog flows.

The current contract test reads:

- [services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go](/D:/source/Shahrah/services/control-plane/internal/admin/provider_catalog_openapi_contract_test.go)

Provider onboarding uses:

- `GET /api/v1/providers`
- `GET /api/v1/providers/{slug}/models`
- `GET /api/v1/providers/{slug}/connection-schemas` for value-free,
  versioned provider-auth field definitions; control-plane may cache this only
  for operator workflows and must not call it during gateway request handling
- `GET /api/v1/models?provider=...` as a paginated compatibility fallback

Rules for future changes:

1. `control-plane` must treat the exported OpenAPI artifact as the canonical review input.
2. When a `model-discovery` endpoint shape changes, update both the client code and the contract test assertions.
3. If `control-plane` starts calling a new `model-discovery` endpoint, extend the contract test to assert that route and its critical params/constraints exist in the spec.
4. Do not add hidden fallback behavior that bypasses the documented upstream contract.

## Minimum Review Checklist

Before merging changes that touch the `model-discovery` contract, confirm:

1. `pnpm --dir apps/model-discovery openapi:export`
2. `go test ./services/control-plane/internal/admin`
3. `model-discovery-web` public clients and `support-admin` admin proxies still point to documented `model-discovery` endpoints

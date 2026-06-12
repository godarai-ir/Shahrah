# Support Admin / Catalog Frontend

Next.js UI for model discovery + admin workflows.

## Run
- `pnpm --filter @shahrah/support-admin dev`
- open `http://localhost:3000`

Configure API base with:
- `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` (default: `http://localhost:3002/api/v1`)

## Model Discovery Contract

`support-admin` is a direct client of `model-discovery`.

- Live Swagger UI: `http://localhost:3002/docs`
- Live OpenAPI JSON: `http://localhost:3002/docs/openapi.json`
- Exported repo spec: [docs/model-discovery-api-contract.md](/D:/source/Shahrah/docs/model-discovery-api-contract.md)

Current client entrypoints:

- Public catalog client: [apps/support-admin/lib/api.ts](/D:/source/Shahrah/apps/support-admin/lib/api.ts)
- Admin login proxy: [apps/support-admin/app/api/auth/login/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/auth/login/route.ts)
- Admin providers proxy: [apps/support-admin/app/api/admin/providers/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/providers/route.ts)
- Admin models proxy: [apps/support-admin/app/api/admin/models/route.ts](/D:/source/Shahrah/apps/support-admin/app/api/admin/models/route.ts)

Implementation rule:

- Before changing any `model-discovery` request or response handling in `support-admin`, read the spec first and verify the route shape against the exported artifact.

## Routes
Public:
- `/`
- `/models`
- `/models/[providerSlug]/[modelSlug]`
- `/providers`
- `/providers/[providerSlug]`
- `/compare`
- `/categories/[categorySlug]`

Admin:
- `/admin`
- `/admin/providers`
- `/admin/models`
- `/admin/models/new`
- `/admin/models/[id]/edit`
- `/admin/pricing`
- `/admin/capabilities`
- `/admin/import`
- `/admin/indexing`
- `/admin/audit-log`

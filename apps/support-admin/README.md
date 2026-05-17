# Support Admin / Catalog Frontend

Next.js UI for model discovery + admin workflows.

## Run
- `pnpm --filter @shahrah/support-admin dev`
- open `http://localhost:3000`

Configure API base with:
- `NEXT_PUBLIC_MODEL_CATALOG_API_BASE` (default: `http://localhost:3002/api/v1`)

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

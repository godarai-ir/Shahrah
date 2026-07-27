# Image model publication

Image models are public and routable only when their model-discovery
`lifecycleStatus` is `published`.

## Workflow

1. Imported image models with incomplete pricing are stored as `draft`.
2. Valid image pricing moves an unpublished model to `review`.
3. A catalog administrator publishes the model explicitly in Support Admin.
4. Control-plane reconciles published models into its local
   `catalog_available` runtime gate.

An image model is pricing-ready when it is explicitly free, or every pricing
profile has valid USD rates and at least one standard output strategy in this
order:

1. `output/image`
2. `output/image_token`
3. `output/token`

The first compatible strategy is used; rates are not added together.

Admin-entered pricing sets `pricingAuthority=admin`, so provider YAML imports
continue updating metadata without replacing the reviewed pricing.

## Operations

- Audit source YAML: `pnpm --dir apps/model-discovery audit:image-pricing`
- Dry-run database classification:
  `pnpm --dir apps/model-discovery reconcile:image-publication -- --expect-unresolved 89`
- Apply demotion after reviewing the report:
  `pnpm --dir apps/model-discovery reconcile:image-publication -- --apply --expect-unresolved 89`
- Run a full model-discovery reindex, then call control-plane
  `POST /api/v1/provider-catalog/reconcile`.

If model-discovery is unavailable, control-plane preserves the last successful
availability snapshot rather than disabling every model.

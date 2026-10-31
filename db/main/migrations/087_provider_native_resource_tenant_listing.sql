-- +goose Up
-- Video jobs are listed per tenant newest-first with cursor pagination, across
-- provider accounts. The existing indexes are keyed by provider, so add one
-- keyed by tenant, resource kind and creation time.
CREATE INDEX IF NOT EXISTS idx_provider_native_resource_tenant_listing
    ON public.provider_native_resource_mappings
    (tenant_id, resource_kind, created_at DESC, public_id DESC);

-- +goose Down
DROP INDEX IF EXISTS public.idx_provider_native_resource_tenant_listing;

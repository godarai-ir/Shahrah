-- +goose Up
-- One-shot contract cutover: legacy token-only JSON cannot be interpreted as
-- canonical pricing profiles without inventing dimensions.
UPDATE public.tenant_provider_account_models
SET public_pricing = NULL,
    private_pricing = NULL;

-- +goose Down
-- Intentionally irreversible: the previous payload shape is unsupported.
SELECT 1;

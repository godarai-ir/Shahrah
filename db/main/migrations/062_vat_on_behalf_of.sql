-- +goose Up
-- +goose StatementBegin
-- On-behalf-of is a privileged VAT capability: a key that has it may send
-- x-godarai-on-behalf-of to act for a tenant member (for example LibreChat).
-- It lives in dedicated columns, not in client-writable token metadata.
ALTER TABLE public.verification_tokens
    ADD COLUMN IF NOT EXISTS on_behalf_of_enabled boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS on_behalf_of_product character varying(32),
    ADD COLUMN IF NOT EXISTS on_behalf_of_updated_by character varying(255),
    ADD COLUMN IF NOT EXISTS on_behalf_of_updated_at timestamp with time zone;

ALTER TABLE public.verification_tokens
    ADD CONSTRAINT chk_verification_tokens_on_behalf_of_vat
        CHECK (NOT on_behalf_of_enabled OR key_type = 'vat'),
    ADD CONSTRAINT chk_verification_tokens_on_behalf_of_product
        CHECK (on_behalf_of_product IS NULL OR on_behalf_of_product ~ '^[a-z][a-z_]{0,31}$');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.verification_tokens
    DROP CONSTRAINT IF EXISTS chk_verification_tokens_on_behalf_of_product,
    DROP CONSTRAINT IF EXISTS chk_verification_tokens_on_behalf_of_vat,
    DROP COLUMN IF EXISTS on_behalf_of_updated_at,
    DROP COLUMN IF EXISTS on_behalf_of_updated_by,
    DROP COLUMN IF EXISTS on_behalf_of_product,
    DROP COLUMN IF EXISTS on_behalf_of_enabled;
-- +goose StatementEnd

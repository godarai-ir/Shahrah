-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.tenant_provider_accounts
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_accounts_auth_mode;

ALTER TABLE public.tenant_provider_accounts
    ADD CONSTRAINT chk_tenant_provider_accounts_auth_mode
    CHECK (auth_mode ~ '^[a-z0-9][a-z0-9_-]{0,63}$');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.tenant_provider_accounts
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_accounts_auth_mode;

ALTER TABLE public.tenant_provider_accounts
    ADD CONSTRAINT chk_tenant_provider_accounts_auth_mode
    CHECK (auth_mode IN ('api_key', 'azure_entra_cert', 'aws_sigv4', 'google_service_account', 'no_credential'))
    NOT VALID;
-- +goose StatementEnd

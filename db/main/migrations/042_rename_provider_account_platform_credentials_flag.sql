-- +goose Up
-- +goose StatementBegin
ALTER TABLE tenant_provider_accounts
    RENAME COLUMN allow_virtual_keys_only TO use_platform_managed_credentials;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tenant_provider_accounts
    RENAME COLUMN use_platform_managed_credentials TO allow_virtual_keys_only;
-- +goose StatementEnd

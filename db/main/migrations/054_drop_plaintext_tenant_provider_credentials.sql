-- +goose Up
-- +goose StatementBegin
-- This is an intentionally destructive cutover. Credentials copied from the
-- readable legacy column remain rotation-required, so revoke and purge those
-- versions instead of carrying potentially exposed secrets forward.
UPDATE public.tenant_provider_accounts
SET active_credential_version_id = NULL,
    credential_hint = NULL,
    has_credential = false
WHERE credential_rotation_required = true;

DELETE FROM public.tenant_provider_account_credential_versions AS versions
USING public.tenant_provider_accounts AS accounts
WHERE versions.account_id = accounts.account_id
  AND accounts.credential_rotation_required = true;

UPDATE public.tenant_provider_accounts
SET credential_rotation_required = false;

-- Fail closed for any inconsistent user-owned API-key account that claims to
-- have a credential without an active encrypted version.
UPDATE public.tenant_provider_accounts
SET credential_hint = NULL,
    has_credential = false
WHERE auth_mode = 'api_key'
  AND use_platform_managed_credentials = false
  AND active_credential_version_id IS NULL;

ALTER TABLE public.tenant_provider_accounts
    DROP COLUMN credential_data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Removed plaintext values cannot and must not be reconstructed on rollback.
ALTER TABLE public.tenant_provider_accounts
    ADD COLUMN credential_data jsonb;
-- +goose StatementEnd

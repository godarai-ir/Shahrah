-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.tenant_provider_accounts
    ADD COLUMN active_credential_version_id uuid,
    ADD COLUMN credential_hint varchar(64),
    ADD COLUMN credential_rotation_required boolean NOT NULL DEFAULT false;

CREATE TABLE public.tenant_provider_account_credential_versions (
    credential_version_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    account_id uuid NOT NULL REFERENCES public.tenant_provider_accounts(account_id) ON DELETE CASCADE,
    provider varchar(100) NOT NULL,
    auth_mode varchar(64) NOT NULL,
    schema_version integer NOT NULL DEFAULT 1,
    version bigint NOT NULL,
    ciphertext bytea NOT NULL,
    data_nonce bytea NOT NULL,
    wrapped_dek bytea NOT NULL,
    wrap_nonce bytea NOT NULL,
    kek_version varchar(64) NOT NULL,
    fingerprint varchar(64) NOT NULL,
    display_hint varchar(64) NOT NULL DEFAULT 'configured',
    status varchar(32) NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CONSTRAINT tenant_provider_credential_version_positive CHECK (version > 0),
    CONSTRAINT tenant_provider_credential_status_valid CHECK (status IN ('active', 'retiring', 'revoked')),
    CONSTRAINT tenant_provider_credential_account_version_unique UNIQUE (account_id, version),
    CONSTRAINT tenant_provider_credential_tenant_account_id_unique UNIQUE (tenant_id, account_id, credential_version_id)
);

CREATE UNIQUE INDEX uq_tenant_provider_credential_active
    ON public.tenant_provider_account_credential_versions(account_id)
    WHERE status = 'active';
CREATE INDEX idx_tenant_provider_credential_runtime
    ON public.tenant_provider_account_credential_versions(tenant_id, account_id, status);

ALTER TABLE public.tenant_provider_accounts
    ADD CONSTRAINT tenant_provider_accounts_active_credential_fkey
    FOREIGN KEY (tenant_id, account_id, active_credential_version_id)
    REFERENCES public.tenant_provider_account_credential_versions(tenant_id, account_id, credential_version_id)
    ON DELETE RESTRICT;

-- Existing readable credentials must be migrated and then rotated. New writes
-- never populate credential_data.
UPDATE public.tenant_provider_accounts
SET credential_rotation_required = true
WHERE has_credential = true
  AND COALESCE(credential_data, '{}'::jsonb) <> '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.tenant_provider_accounts
    DROP CONSTRAINT IF EXISTS tenant_provider_accounts_active_credential_fkey;
DROP TABLE IF EXISTS public.tenant_provider_account_credential_versions;
ALTER TABLE public.tenant_provider_accounts
    DROP COLUMN IF EXISTS active_credential_version_id,
    DROP COLUMN IF EXISTS credential_hint,
    DROP COLUMN IF EXISTS credential_rotation_required;
-- +goose StatementEnd

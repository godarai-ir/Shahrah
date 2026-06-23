-- +goose Up
CREATE TABLE IF NOT EXISTS public.platform_service_account_credentials (
    credential_id text PRIMARY KEY,
    account_name text NOT NULL,
    token_hash character varying(64) NOT NULL UNIQUE,
    scopes text[] NOT NULL DEFAULT '{}',
    active boolean NOT NULL DEFAULT true,
    expires_at timestamptz,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_platform_service_account_name
    ON public.platform_service_account_credentials(account_name);
CREATE INDEX IF NOT EXISTS idx_platform_service_account_expiry
    ON public.platform_service_account_credentials(expires_at);

-- +goose Down
DROP TABLE IF EXISTS public.platform_service_account_credentials;

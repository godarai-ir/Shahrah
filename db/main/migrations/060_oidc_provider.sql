-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.oidc_clients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    client_id text NOT NULL,
    name text NOT NULL,
    secret_hash text NOT NULL,
    redirect_uris text[] NOT NULL DEFAULT '{}',
    post_logout_redirect_uris text[] NOT NULL DEFAULT '{}',
    active boolean NOT NULL DEFAULT true,
    created_by text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    secret_rotated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT oidc_clients_client_id_key UNIQUE (client_id)
);

CREATE INDEX IF NOT EXISTS oidc_clients_tenant_idx
    ON public.oidc_clients (tenant_id, created_at DESC);

-- One row per authorization request. The row carries the login challenge and,
-- once a user accepts it, the single-use authorization code.
CREATE TABLE IF NOT EXISTS public.oidc_authorization_requests (
    id text PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    tenant_slug text NOT NULL,
    issuer text NOT NULL,
    client_id text NOT NULL REFERENCES public.oidc_clients(client_id) ON DELETE CASCADE,
    redirect_uri text NOT NULL,
    scope text NOT NULL,
    state text NOT NULL DEFAULT '',
    nonce text NOT NULL DEFAULT '',
    code_challenge text NOT NULL,
    code_challenge_method text NOT NULL,
    prompt text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'pending',
    user_id uuid,
    grant_id uuid,
    auth_time timestamptz,
    code_hash text,
    code_expires_at timestamptz,
    code_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    CONSTRAINT oidc_authorization_requests_status_check
        CHECK (status IN ('pending', 'accepted', 'rejected')),
    CONSTRAINT oidc_authorization_requests_method_check
        CHECK (code_challenge_method = 'S256')
);

CREATE UNIQUE INDEX IF NOT EXISTS oidc_authorization_requests_code_uidx
    ON public.oidc_authorization_requests (code_hash)
    WHERE code_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS oidc_authorization_requests_expiry_idx
    ON public.oidc_authorization_requests (expires_at);

-- Refresh tokens rotate on every use. grant_id groups one login's rotation
-- chain so that replaying a used token revokes the whole chain.
CREATE TABLE IF NOT EXISTS public.oidc_refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    grant_id uuid NOT NULL,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    client_id text NOT NULL REFERENCES public.oidc_clients(client_id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    scope text NOT NULL,
    nonce text NOT NULL DEFAULT '',
    auth_time timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    grant_expires_at timestamptz NOT NULL,
    used_at timestamptz,
    revoked_at timestamptz,
    CONSTRAINT oidc_refresh_tokens_token_hash_key UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS oidc_refresh_tokens_grant_idx
    ON public.oidc_refresh_tokens (grant_id);
CREATE INDEX IF NOT EXISTS oidc_refresh_tokens_user_idx
    ON public.oidc_refresh_tokens (tenant_id, user_id)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS oidc_refresh_tokens_expiry_idx
    ON public.oidc_refresh_tokens (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.oidc_refresh_tokens;
DROP TABLE IF EXISTS public.oidc_authorization_requests;
DROP TABLE IF EXISTS public.oidc_clients;
-- +goose StatementEnd

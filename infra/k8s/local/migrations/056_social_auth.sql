-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.social_auth_users (
    id text PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL,
    "emailVerified" boolean NOT NULL DEFAULT false,
    image text,
    "createdAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS social_auth_users_email_uidx
    ON public.social_auth_users (LOWER(email));

CREATE TABLE IF NOT EXISTS public.social_auth_sessions (
    id text PRIMARY KEY,
    "expiresAt" timestamptz NOT NULL,
    token text NOT NULL UNIQUE,
    "createdAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "ipAddress" text,
    "userAgent" text,
    "userId" text NOT NULL REFERENCES public.social_auth_users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS social_auth_sessions_user_id_idx
    ON public.social_auth_sessions ("userId");
CREATE INDEX IF NOT EXISTS social_auth_sessions_expires_at_idx
    ON public.social_auth_sessions ("expiresAt");

CREATE TABLE IF NOT EXISTS public.social_auth_accounts (
    id text PRIMARY KEY,
    "accountId" text NOT NULL,
    "providerId" text NOT NULL,
    "userId" text NOT NULL REFERENCES public.social_auth_users(id) ON DELETE CASCADE,
    "accessToken" text,
    "refreshToken" text,
    "idToken" text,
    "accessTokenExpiresAt" timestamptz,
    "refreshTokenExpiresAt" timestamptz,
    scope text,
    password text,
    "createdAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT social_auth_accounts_provider_subject_key UNIQUE ("providerId", "accountId")
);

CREATE INDEX IF NOT EXISTS social_auth_accounts_user_id_idx
    ON public.social_auth_accounts ("userId");

CREATE TABLE IF NOT EXISTS public.social_auth_verifications (
    id text PRIMARY KEY,
    identifier text NOT NULL,
    value text NOT NULL,
    "expiresAt" timestamptz NOT NULL,
    "createdAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS social_auth_verifications_identifier_idx
    ON public.social_auth_verifications (identifier);
CREATE INDEX IF NOT EXISTS social_auth_verifications_expires_at_idx
    ON public.social_auth_verifications ("expiresAt");

CREATE TABLE IF NOT EXISTS public.social_auth_rate_limits (
    id text PRIMARY KEY,
    key text NOT NULL UNIQUE,
    count integer NOT NULL,
    "lastRequest" bigint NOT NULL
);

ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS account_id text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_account_id_fkey'
    ) THEN
        ALTER TABLE public.users
            ADD CONSTRAINT users_account_id_fkey
            FOREIGN KEY (account_id)
            REFERENCES public.social_auth_users(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_tenant_account
    ON public.users (tenant_id, account_id)
    WHERE account_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_account_id
    ON public.users (account_id)
    WHERE account_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.social_auth_handoffs (
    handoff_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash character varying(64) NOT NULL UNIQUE,
    purpose character varying(16) NOT NULL,
    account_id text REFERENCES public.social_auth_users(id) ON DELETE CASCADE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    user_id uuid REFERENCES public.users(user_id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT social_auth_handoffs_purpose_check
        CHECK (purpose IN ('login', 'link')),
    CONSTRAINT social_auth_handoffs_shape_check CHECK (
        (purpose = 'login' AND account_id IS NOT NULL AND user_id IS NOT NULL)
        OR
        (purpose = 'link' AND account_id IS NULL AND user_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS social_auth_handoffs_expiry_idx
    ON public.social_auth_handoffs (expires_at)
    WHERE consumed_at IS NULL;
CREATE INDEX IF NOT EXISTS social_auth_handoffs_account_idx
    ON public.social_auth_handoffs (account_id, tenant_id)
    WHERE account_id IS NOT NULL AND consumed_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.social_auth_handoffs;
DROP INDEX IF EXISTS public.idx_users_account_id;
DROP INDEX IF EXISTS public.uq_users_tenant_account;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_account_id_fkey;
ALTER TABLE public.users DROP COLUMN IF EXISTS account_id;
DROP TABLE IF EXISTS public.social_auth_rate_limits;
DROP TABLE IF EXISTS public.social_auth_verifications;
DROP TABLE IF EXISTS public.social_auth_accounts;
DROP TABLE IF EXISTS public.social_auth_sessions;
DROP TABLE IF EXISTS public.social_auth_users;
-- +goose StatementEnd

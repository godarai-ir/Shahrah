-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.platform_provider_connections (
    connection_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_slug varchar(100) NOT NULL,
    name varchar(255) NOT NULL,
    environment varchar(64) NOT NULL DEFAULT 'production',
    auth_mode varchar(64) NOT NULL,
    schema_version integer NOT NULL DEFAULT 1,
    config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    status varchar(32) NOT NULL DEFAULT 'draft',
    region varchar(100),
    priority integer NOT NULL DEFAULT 0,
    weight double precision NOT NULL DEFAULT 1,
    is_primary boolean NOT NULL DEFAULT false,
    active_credential_version_id uuid,
    last_test_status varchar(32),
    last_test_category varchar(64),
    last_test_summary varchar(255),
    last_tested_at timestamptz,
    last_success_at timestamptz,
    last_rotation_at timestamptz,
    revision bigint NOT NULL DEFAULT 1,
    created_by varchar(255),
    updated_by varchar(255),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT platform_provider_connections_provider_chk CHECK (provider_slug = lower(trim(provider_slug)) AND provider_slug <> ''),
    CONSTRAINT platform_provider_connections_environment_chk CHECK (trim(environment) <> ''),
    CONSTRAINT platform_provider_connections_schema_version_chk CHECK (schema_version > 0),
    CONSTRAINT platform_provider_connections_status_chk CHECK (status IN ('draft', 'testing', 'active', 'degraded', 'disabled')),
    CONSTRAINT platform_provider_connections_weight_chk CHECK (weight >= 0),
    CONSTRAINT platform_provider_connections_revision_chk CHECK (revision > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_provider_connections_name
    ON public.platform_provider_connections (provider_slug, environment, lower(name));

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_provider_connections_active_primary
    ON public.platform_provider_connections (provider_slug, environment)
    WHERE status = 'active' AND is_primary = true;

CREATE INDEX IF NOT EXISTS idx_platform_provider_connections_lookup
    ON public.platform_provider_connections (provider_slug, environment, status, is_primary);

CREATE TABLE IF NOT EXISTS public.platform_provider_credential_versions (
    credential_version_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL REFERENCES public.platform_provider_connections(connection_id) ON DELETE CASCADE,
    version bigint NOT NULL,
    secret_backend varchar(32) NOT NULL DEFAULT 'encrypted_db',
    ciphertext bytea NOT NULL,
    data_nonce bytea NOT NULL,
    wrapped_dek bytea NOT NULL,
    wrap_nonce bytea NOT NULL,
    kek_version varchar(64) NOT NULL,
    fingerprint varchar(64) NOT NULL,
    display_hint varchar(64) NOT NULL DEFAULT 'configured',
    status varchar(32) NOT NULL DEFAULT 'draft',
    valid_from timestamptz,
    expires_at timestamptz,
    retire_at timestamptz,
    created_by varchar(255),
    activated_by varchar(255),
    revoked_by varchar(255),
    revoked_at timestamptz,
    idempotency_key varchar(128),
    request_hash varchar(64),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT platform_provider_credential_versions_version_chk CHECK (version > 0),
    CONSTRAINT platform_provider_credential_versions_backend_chk CHECK (secret_backend IN ('encrypted_db')),
    CONSTRAINT platform_provider_credential_versions_status_chk CHECK (status IN ('draft', 'validated', 'active', 'retiring', 'revoked')),
    CONSTRAINT uq_platform_provider_credential_versions_version UNIQUE (connection_id, version),
    CONSTRAINT uq_platform_provider_credential_versions_connection_id UNIQUE (connection_id, credential_version_id),
    CONSTRAINT uq_platform_provider_credential_versions_idempotency UNIQUE (connection_id, idempotency_key)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_provider_credential_versions_active
    ON public.platform_provider_credential_versions (connection_id)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_platform_provider_credential_versions_connection
    ON public.platform_provider_credential_versions (connection_id, version DESC);

ALTER TABLE public.platform_provider_connections
    ADD CONSTRAINT platform_provider_connections_active_version_fkey
    FOREIGN KEY (connection_id, active_credential_version_id)
    REFERENCES public.platform_provider_credential_versions(connection_id, credential_version_id)
    ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS public.platform_provider_connection_scopes (
    scope_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL REFERENCES public.platform_provider_connections(connection_id) ON DELETE CASCADE,
    model_pattern varchar(255),
    deployment_pattern varchar(255),
    capabilities text[] NOT NULL DEFAULT '{}',
    regions text[] NOT NULL DEFAULT '{}',
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT platform_provider_connection_scopes_shape_chk CHECK (
        model_pattern IS NOT NULL OR deployment_pattern IS NOT NULL OR cardinality(capabilities) > 0 OR cardinality(regions) > 0
    )
);

CREATE INDEX IF NOT EXISTS idx_platform_provider_connection_scopes_connection
    ON public.platform_provider_connection_scopes (connection_id, enabled);

CREATE TABLE IF NOT EXISTS public.platform_provider_credential_events (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id uuid NOT NULL REFERENCES public.platform_provider_connections(connection_id) ON DELETE CASCADE,
    credential_version_id uuid REFERENCES public.platform_provider_credential_versions(credential_version_id) ON DELETE SET NULL,
    actor_id varchar(255),
    actor_role varchar(100),
    service_account_id varchar(255),
    request_id varchar(255),
    action varchar(64) NOT NULL,
    result varchar(32) NOT NULL,
    reason varchar(255),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_ip varchar(64),
    idempotency_key varchar(128),
    request_hash varchar(64),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT platform_provider_credential_events_result_chk CHECK (result IN ('success', 'failure', 'rejected'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_provider_credential_events_idempotency
    ON public.platform_provider_credential_events (action, idempotency_key);

CREATE INDEX IF NOT EXISTS idx_platform_provider_credential_events_connection
    ON public.platform_provider_credential_events (connection_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.platform_provider_credential_events;
DROP TABLE IF EXISTS public.platform_provider_connection_scopes;
ALTER TABLE IF EXISTS public.platform_provider_connections
    DROP CONSTRAINT IF EXISTS platform_provider_connections_active_version_fkey;
DROP TABLE IF EXISTS public.platform_provider_credential_versions;
DROP TABLE IF EXISTS public.platform_provider_connections;
-- +goose StatementEnd

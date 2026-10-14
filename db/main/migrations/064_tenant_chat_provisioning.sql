-- +goose Up
-- +goose StatementBegin
-- Tracks what has been registered with GodarChat (LibreChat) for each tenant. A worker
-- reconciles every tenant against this table, so a tenant created by any path gets a chat.
CREATE TABLE IF NOT EXISTS public.tenant_chat_provisioning (
    tenant_id uuid PRIMARY KEY REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    state text NOT NULL DEFAULT 'pending',
    synced_slug text NOT NULL DEFAULT '',
    synced_status text NOT NULL DEFAULT '',
    oidc_client_id text NOT NULL DEFAULT '',
    credentials_delivered boolean NOT NULL DEFAULT false,
    attempts integer NOT NULL DEFAULT 0,
    last_error text NOT NULL DEFAULT '',
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_until timestamptz,
    ready_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tenant_chat_provisioning_state_check CHECK (state IN ('pending', 'ready', 'skipped'))
);

CREATE INDEX IF NOT EXISTS tenant_chat_provisioning_due_idx
    ON public.tenant_chat_provisioning (next_attempt_at)
    WHERE state <> 'skipped';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.tenant_chat_provisioning;
-- +goose StatementEnd

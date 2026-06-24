-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_events (
    event_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    event_type character varying(96) NOT NULL,
    severity character varying(16) NOT NULL DEFAULT 'info',
    actor_id character varying(255), actor_role character varying(64), reason text,
    request_id character varying(255), payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_tenant_events_severity CHECK (severity IN ('info', 'warning', 'critical'))
);
CREATE INDEX IF NOT EXISTS idx_tenant_events_tenant_created ON public.tenant_events (tenant_id, created_at DESC, event_id DESC);

CREATE TABLE IF NOT EXISTS public.wallet_manual_adjustments (
    adjustment_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    wallet_transaction_id uuid NOT NULL REFERENCES public.wallet_transactions(transaction_id),
    idempotency_key character varying(128) NOT NULL, direction character varying(16) NOT NULL,
    amount_minor numeric(18,0) NOT NULL, reason text NOT NULL,
    external_reference character varying(255), actor_id character varying(255) NOT NULL,
    actor_role character varying(64), request_id character varying(255),
    balance_before numeric(18,0) NOT NULL, balance_after numeric(18,0) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_wallet_manual_adjustment_direction CHECK (direction IN ('credit', 'debit')),
    CONSTRAINT chk_wallet_manual_adjustment_amount CHECK (amount_minor > 0),
    CONSTRAINT uq_wallet_manual_adjustment_idempotency UNIQUE (tenant_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_wallet_manual_adjustments_tenant_created ON public.wallet_manual_adjustments (tenant_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.wallet_manual_adjustments;
DROP TABLE IF EXISTS public.tenant_events;

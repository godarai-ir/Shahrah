-- +goose Up
CREATE TABLE IF NOT EXISTS public.support_tickets (
    ticket_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id character varying(24) NOT NULL,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE RESTRICT,
    requester_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    subject character varying(160) NOT NULL,
    category character varying(32) NOT NULL,
    priority character varying(16) NOT NULL DEFAULT 'normal',
    status character varying(32) NOT NULL DEFAULT 'open',
    assignee_id character varying(255),
    last_message_author_type character varying(32) NOT NULL DEFAULT 'tenant_user',
    last_message_at timestamp with time zone NOT NULL DEFAULT now(),
    first_support_response_at timestamp with time zone,
    resolved_at timestamp with time zone,
    closed_at timestamp with time zone,
    idempotency_key character varying(128) NOT NULL,
    request_hash character varying(64) NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT uq_support_tickets_public_id UNIQUE (public_id),
    CONSTRAINT uq_support_tickets_create_idempotency UNIQUE (tenant_id, requester_user_id, idempotency_key),
    CONSTRAINT chk_support_tickets_subject CHECK (char_length(btrim(subject)) BETWEEN 5 AND 160),
    CONSTRAINT chk_support_tickets_category CHECK (category IN ('technical', 'billing', 'account', 'feature_request', 'other')),
    CONSTRAINT chk_support_tickets_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT chk_support_tickets_status CHECK (status IN ('open', 'in_progress', 'waiting_for_customer', 'resolved', 'closed')),
    CONSTRAINT chk_support_tickets_last_author CHECK (last_message_author_type IN ('tenant_user', 'support_agent', 'system')),
    CONSTRAINT chk_support_tickets_version CHECK (version > 0)
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_tenant_activity
    ON public.support_tickets (tenant_id, last_message_at DESC, ticket_id DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_tenant_status_activity
    ON public.support_tickets (tenant_id, status, last_message_at DESC, ticket_id DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_support_queue
    ON public.support_tickets (status, last_message_author_type, last_message_at, ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assignee
    ON public.support_tickets (assignee_id, status, last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_triage
    ON public.support_tickets (category, priority, status, last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_resolved_close
    ON public.support_tickets (resolved_at, ticket_id)
    WHERE status = 'resolved';

CREATE TABLE IF NOT EXISTS public.support_ticket_messages (
    message_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    ticket_id uuid NOT NULL REFERENCES public.support_tickets(ticket_id) ON DELETE RESTRICT,
    author_type character varying(32) NOT NULL,
    author_id character varying(255) NOT NULL,
    body text NOT NULL DEFAULT '',
    idempotency_key character varying(128) NOT NULL,
    request_hash character varying(64) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT uq_support_ticket_message_idempotency UNIQUE (ticket_id, author_type, author_id, idempotency_key),
    CONSTRAINT chk_support_ticket_message_author CHECK (author_type IN ('tenant_user', 'support_agent', 'system')),
    CONSTRAINT chk_support_ticket_message_body CHECK (char_length(body) <= 10000)
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_ticket_created
    ON public.support_ticket_messages (ticket_id, created_at, message_id);

CREATE TABLE IF NOT EXISTS public.support_ticket_attachments (
    attachment_id uuid PRIMARY KEY,
    message_id uuid NOT NULL REFERENCES public.support_ticket_messages(message_id) ON DELETE RESTRICT,
    ticket_id uuid NOT NULL REFERENCES public.support_tickets(ticket_id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE RESTRICT,
    original_filename character varying(255) NOT NULL,
    detected_content_type character varying(127) NOT NULL,
    size_bytes bigint NOT NULL,
    sha256 character varying(64) NOT NULL,
    object_key text NOT NULL,
    scan_status character varying(16) NOT NULL,
    scan_result character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    scanned_at timestamp with time zone,
    CONSTRAINT uq_support_ticket_attachment_object_key UNIQUE (object_key),
    CONSTRAINT chk_support_ticket_attachment_size CHECK (size_bytes > 0 AND size_bytes <= 10485760),
    CONSTRAINT chk_support_ticket_attachment_scan CHECK (scan_status IN ('quarantined', 'clean', 'rejected'))
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_attachments_message
    ON public.support_ticket_attachments (message_id, created_at, attachment_id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_attachments_ticket_tenant
    ON public.support_ticket_attachments (tenant_id, ticket_id, attachment_id);

CREATE TABLE IF NOT EXISTS public.support_ticket_ratings (
    rating_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    ticket_id uuid NOT NULL REFERENCES public.support_tickets(ticket_id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE RESTRICT,
    requester_user_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE RESTRICT,
    score smallint NOT NULL,
    comment character varying(500),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT uq_support_ticket_rating_ticket UNIQUE (ticket_id),
    CONSTRAINT chk_support_ticket_rating_score CHECK (score BETWEEN 1 AND 5)
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_ratings_tenant_updated
    ON public.support_ticket_ratings (tenant_id, updated_at DESC, rating_id DESC);

CREATE TABLE IF NOT EXISTS public.support_ticket_mutations (
    mutation_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    ticket_id uuid NOT NULL REFERENCES public.support_tickets(ticket_id) ON DELETE RESTRICT,
    actor_type character varying(32) NOT NULL,
    actor_id character varying(255) NOT NULL,
    operation character varying(64) NOT NULL,
    idempotency_key character varying(128) NOT NULL,
    request_hash character varying(64) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT uq_support_ticket_mutation_idempotency UNIQUE (ticket_id, actor_type, actor_id, operation, idempotency_key),
    CONSTRAINT chk_support_ticket_mutation_actor CHECK (actor_type IN ('tenant_user', 'support_agent', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_mutations_ticket_created
    ON public.support_ticket_mutations (ticket_id, created_at DESC, mutation_id DESC);

CREATE TABLE IF NOT EXISTS public.support_ticket_events (
    event_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    ticket_id uuid NOT NULL REFERENCES public.support_tickets(ticket_id) ON DELETE RESTRICT,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE RESTRICT,
    event_type character varying(96) NOT NULL,
    actor_type character varying(32) NOT NULL,
    actor_id character varying(255),
    actor_role character varying(64),
    request_id character varying(255),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_support_ticket_event_actor CHECK (actor_type IN ('tenant_user', 'support_agent', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_events_ticket_created
    ON public.support_ticket_events (ticket_id, created_at DESC, event_id DESC);

UPDATE public.platform_service_account_credentials AS credential
SET scopes = (
    SELECT ARRAY_AGG(DISTINCT scope ORDER BY scope)
    FROM UNNEST(
        credential.scopes || ARRAY[
            'platform:support-tickets:read',
            'platform:support-tickets:write'
        ]::text[]
    ) AS scope
)
WHERE LOWER(BTRIM(credential.account_name)) IN ('support-admin', 'platform-support')
  AND credential.active = true
  AND credential.revoked_at IS NULL
  AND (credential.expires_at IS NULL OR credential.expires_at > NOW());

-- +goose Down
DROP TABLE IF EXISTS public.support_ticket_events;
DROP TABLE IF EXISTS public.support_ticket_mutations;
DROP TABLE IF EXISTS public.support_ticket_ratings;
DROP TABLE IF EXISTS public.support_ticket_attachments;
DROP TABLE IF EXISTS public.support_ticket_messages;
DROP TABLE IF EXISTS public.support_tickets;

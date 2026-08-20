-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS rag_profiles (
    revision text PRIMARY KEY,
    embedding_model text NOT NULL,
    embedding_dimensions integer NOT NULL,
    rerank_model text NOT NULL,
    chunk_tokens integer NOT NULL,
    chunk_overlap_tokens integer NOT NULL,
    max_chunks integer NOT NULL,
    max_context_tokens integer NOT NULL,
    configuration jsonb NOT NULL DEFAULT '{}'::jsonb,
    active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT rag_profile_dimensions_check CHECK (embedding_dimensions > 0),
    CONSTRAINT rag_profile_chunk_check CHECK (chunk_tokens > 0 AND chunk_overlap_tokens >= 0 AND chunk_overlap_tokens < chunk_tokens)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rag_profiles_active
    ON rag_profiles (active) WHERE active = true;

INSERT INTO rag_profiles (
    revision, embedding_model, embedding_dimensions, rerank_model,
    chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens,
    configuration, active
)
VALUES (
    'rag-v1', 'platform/embedding-1536', 1536, 'platform/cohere-multilingual-rerank',
    700, 100, 8, 8000,
    '{"retrieval_per_knowledge_base":10,"candidate_limit":50,"minimum_rerank_score":0.05,"conversation_turns":6,"conversation_tokens":6000}'::jsonb,
    true
)
ON CONFLICT (revision) DO NOTHING;

CREATE TABLE IF NOT EXISTS knowledge_bases (
    knowledge_base_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    rag_profile_revision text NOT NULL REFERENCES rag_profiles(revision),
    created_by text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT knowledge_base_name_check CHECK (length(btrim(name)) BETWEEN 1 AND 160),
    UNIQUE (tenant_id, knowledge_base_id)
);

CREATE INDEX IF NOT EXISTS idx_knowledge_bases_tenant
    ON knowledge_bases (tenant_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS knowledge_base_documents (
    document_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL,
    title text NOT NULL,
    source_type text NOT NULL,
    media_type text NOT NULL,
    active_version_id uuid,
    processing_version_id uuid,
    created_by text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT knowledge_base_document_source_check CHECK (source_type IN ('file', 'text')),
    CONSTRAINT knowledge_base_documents_kb_tenant_fk
        FOREIGN KEY (tenant_id, knowledge_base_id)
        REFERENCES knowledge_bases(tenant_id, knowledge_base_id) ON DELETE CASCADE,
    UNIQUE (tenant_id, document_id)
);

CREATE INDEX IF NOT EXISTS idx_kb_documents_tenant_kb
    ON knowledge_base_documents (tenant_id, knowledge_base_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS document_versions (
    version_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL,
    document_id uuid NOT NULL,
    version integer NOT NULL,
    status text NOT NULL DEFAULT 'queued',
    object_key text,
    source_text text,
    content_sha256 text NOT NULL,
    size_bytes bigint NOT NULL DEFAULT 0,
    page_count integer,
    chunk_count integer NOT NULL DEFAULT 0,
    failure_code text,
    failure_message text,
    activated_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT document_version_status_check CHECK (status IN ('queued','scanning','parsing','chunking','embedding','indexing','ready','failed','deleting')),
    CONSTRAINT document_version_content_check CHECK (object_key IS NOT NULL OR source_text IS NOT NULL),
    CONSTRAINT document_versions_kb_tenant_fk
        FOREIGN KEY (tenant_id, knowledge_base_id)
        REFERENCES knowledge_bases(tenant_id, knowledge_base_id) ON DELETE CASCADE,
    CONSTRAINT document_versions_document_tenant_fk
        FOREIGN KEY (tenant_id, document_id)
        REFERENCES knowledge_base_documents(tenant_id, document_id) ON DELETE CASCADE,
    UNIQUE (document_id, version),
    UNIQUE (tenant_id, version_id)
);

ALTER TABLE knowledge_base_documents
    ADD CONSTRAINT knowledge_base_documents_active_version_fk
    FOREIGN KEY (tenant_id, active_version_id)
    REFERENCES document_versions(tenant_id, version_id) ON DELETE SET NULL (active_version_id);
ALTER TABLE knowledge_base_documents
    ADD CONSTRAINT knowledge_base_documents_processing_version_fk
    FOREIGN KEY (tenant_id, processing_version_id)
    REFERENCES document_versions(tenant_id, version_id) ON DELETE SET NULL (processing_version_id);

CREATE INDEX IF NOT EXISTS idx_document_versions_processing
    ON document_versions (status, updated_at) WHERE status NOT IN ('ready', 'failed');

CREATE TABLE IF NOT EXISTS rag_ingestion_jobs (
    job_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL,
    document_id uuid NOT NULL,
    version_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    stage text NOT NULL DEFAULT 'queued',
    status text NOT NULL DEFAULT 'queued',
    attempt integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    lease_owner text,
    lease_expires_at timestamptz,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    error_code text,
    error_details jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT rag_ingestion_job_status_check CHECK (status IN ('queued','running','retry','completed','failed','canceled')),
    CONSTRAINT rag_ingestion_jobs_kb_tenant_fk
        FOREIGN KEY (tenant_id, knowledge_base_id)
        REFERENCES knowledge_bases(tenant_id, knowledge_base_id) ON DELETE CASCADE,
    CONSTRAINT rag_ingestion_jobs_document_tenant_fk
        FOREIGN KEY (tenant_id, document_id)
        REFERENCES knowledge_base_documents(tenant_id, document_id) ON DELETE CASCADE,
    CONSTRAINT rag_ingestion_jobs_version_tenant_fk
        FOREIGN KEY (tenant_id, version_id)
        REFERENCES document_versions(tenant_id, version_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_rag_ingestion_jobs_claim
    ON rag_ingestion_jobs (next_attempt_at, created_at)
    WHERE status IN ('queued', 'retry', 'running');

CREATE TABLE IF NOT EXISTS rag_purge_jobs (
    purge_job_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    object_key text NOT NULL,
    status text NOT NULL DEFAULT 'queued',
    attempt integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 10,
    lease_owner text,
    lease_expires_at timestamptz,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    error_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, object_key),
    CONSTRAINT rag_purge_job_status_check CHECK (status IN ('queued','running','retry','completed','failed'))
);

CREATE INDEX IF NOT EXISTS idx_rag_purge_jobs_claim
    ON rag_purge_jobs (next_attempt_at, created_at) WHERE status IN ('queued','retry','running');

CREATE OR REPLACE FUNCTION queue_rag_tenant_purge()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO rag_purge_jobs (tenant_id, object_key)
    SELECT DISTINCT OLD.tenant_id, object_key
    FROM document_versions
    WHERE tenant_id = OLD.tenant_id AND object_key IS NOT NULL
    ON CONFLICT (tenant_id, object_key) DO UPDATE
    SET status = 'queued', attempt = 0, lease_owner = NULL, lease_expires_at = NULL,
        next_attempt_at = now(), error_code = NULL, updated_at = now();

    INSERT INTO rag_purge_jobs (tenant_id, object_key)
    VALUES (OLD.tenant_id, '__tenant_vectors__')
    ON CONFLICT (tenant_id, object_key) DO UPDATE
    SET status = 'queued', attempt = 0, lease_owner = NULL, lease_expires_at = NULL,
        next_attempt_at = now(), error_code = NULL, updated_at = now();

    RETURN OLD;
END;
$$;

DROP TRIGGER IF EXISTS tenants_queue_rag_purge ON tenants;
CREATE TRIGGER tenants_queue_rag_purge
BEFORE DELETE ON tenants
FOR EACH ROW EXECUTE FUNCTION queue_rag_tenant_purge();

CREATE TABLE IF NOT EXISTS chatbots (
    chatbot_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id text NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    system_instruction text NOT NULL DEFAULT '',
    answer_model text NOT NULL,
    tone text NOT NULL DEFAULT 'professional',
    access_mode text NOT NULL DEFAULT 'private',
    retention_days integer NOT NULL DEFAULT 30,
    knowledge_base_ids uuid[] NOT NULL,
    allowed_origins text[] NOT NULL DEFAULT '{}',
    branding jsonb NOT NULL DEFAULT '{}'::jsonb,
    published_version_id uuid,
    created_by text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT chatbot_access_mode_check CHECK (access_mode IN ('private','public')),
    CONSTRAINT chatbot_retention_check CHECK (retention_days IN (0,7,30)),
    CONSTRAINT chatbot_knowledge_base_count_check CHECK (cardinality(knowledge_base_ids) BETWEEN 1 AND 5),
    UNIQUE (tenant_id, chatbot_id)
);

CREATE INDEX IF NOT EXISTS idx_chatbots_tenant
    ON chatbots (tenant_id, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS chatbot_versions (
    version_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    chatbot_id uuid NOT NULL,
    revision integer NOT NULL,
    configuration jsonb NOT NULL,
    rag_profile_revision text NOT NULL REFERENCES rag_profiles(revision),
    published_at timestamptz NOT NULL DEFAULT now(),
    unpublished_at timestamptz,
    CONSTRAINT chatbot_versions_chatbot_tenant_fk
        FOREIGN KEY (tenant_id, chatbot_id)
        REFERENCES chatbots(tenant_id, chatbot_id) ON DELETE CASCADE,
    UNIQUE (chatbot_id, revision),
    UNIQUE (tenant_id, version_id)
);

ALTER TABLE chatbots
    ADD CONSTRAINT chatbots_published_version_fk
    FOREIGN KEY (tenant_id, published_version_id)
    REFERENCES chatbot_versions(tenant_id, version_id) ON DELETE SET NULL (published_version_id);

CREATE TABLE IF NOT EXISTS chatbot_sessions (
    session_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    chatbot_id uuid NOT NULL,
    chatbot_version_id uuid NOT NULL,
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chatbot_sessions_chatbot_tenant_fk
        FOREIGN KEY (tenant_id, chatbot_id)
        REFERENCES chatbots(tenant_id, chatbot_id) ON DELETE CASCADE,
    CONSTRAINT chatbot_sessions_version_tenant_fk
        FOREIGN KEY (tenant_id, chatbot_version_id)
        REFERENCES chatbot_versions(tenant_id, version_id) ON DELETE CASCADE,
    UNIQUE (tenant_id, session_id)
);

CREATE INDEX IF NOT EXISTS idx_chatbot_sessions_expiry ON chatbot_sessions (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS chatbot_messages (
    message_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    session_id uuid NOT NULL,
    role text NOT NULL,
    content_encrypted bytea,
    refusal boolean NOT NULL DEFAULT false,
    citations jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz,
    CONSTRAINT chatbot_message_role_check CHECK (role IN ('user','assistant')),
    CONSTRAINT chatbot_messages_session_tenant_fk
        FOREIGN KEY (tenant_id, session_id)
        REFERENCES chatbot_sessions(tenant_id, session_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chatbot_messages_session ON chatbot_messages (tenant_id, session_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chatbot_messages_expiry ON chatbot_messages (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS rag_usage_events (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    request_id text NOT NULL,
    knowledge_base_ids uuid[] NOT NULL DEFAULT '{}',
    chatbot_id uuid,
    operation text NOT NULL,
    status text NOT NULL,
    embedding_tokens bigint NOT NULL DEFAULT 0,
    rerank_units bigint NOT NULL DEFAULT 0,
    generation_input_tokens bigint NOT NULL DEFAULT 0,
    generation_output_tokens bigint NOT NULL DEFAULT 0,
    provider_cost_usd numeric(20,10) NOT NULL DEFAULT 0,
    platform_query_fee_irr numeric(20,2) NOT NULL DEFAULT 0,
    latency_ms integer NOT NULL DEFAULT 0,
    refusal boolean NOT NULL DEFAULT false,
    citation_valid boolean,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT rag_usage_events_chatbot_tenant_fk
        FOREIGN KEY (tenant_id, chatbot_id)
        REFERENCES chatbots(tenant_id, chatbot_id) ON DELETE SET NULL (chatbot_id)
);

CREATE INDEX IF NOT EXISTS idx_rag_usage_events_tenant_created
    ON rag_usage_events (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS rag_rate_catalog (
    rate_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dimension text NOT NULL,
    unit text NOT NULL,
    currency text NOT NULL DEFAULT 'IRR',
    unit_price_irr numeric(20,2) NOT NULL,
    active boolean NOT NULL DEFAULT true,
    effective_from timestamptz NOT NULL DEFAULT now(),
    effective_until timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT rag_rate_catalog_currency_check CHECK (currency = 'IRR'),
    CONSTRAINT rag_rate_catalog_price_check CHECK (unit_price_irr >= 0),
    CONSTRAINT rag_rate_catalog_window_check CHECK (effective_until IS NULL OR effective_until > effective_from)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rag_rate_catalog_active_dimension
    ON rag_rate_catalog (dimension)
    WHERE active = true AND effective_until IS NULL;

INSERT INTO rag_rate_catalog (dimension, unit, currency, unit_price_irr)
VALUES ('platform_query_completed', 'query', 'IRR', 0)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS rag_audit_events (
    audit_event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    actor_id text,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rag_audit_events_tenant_created
    ON rag_audit_events (tenant_id, created_at DESC);

INSERT INTO feature_catalog (feature_key, value_type, default_value, allow_override, category, description)
VALUES
    ('rag.enabled', 'boolean', 'false'::jsonb, true, 'rag', 'Managed RAG add-on access.'),
    ('rag.knowledge_bases.max', 'integer', '0'::jsonb, true, 'rag', 'Maximum managed knowledge bases.'),
    ('rag.documents.max', 'integer', '0'::jsonb, true, 'rag', 'Maximum managed RAG documents.'),
    ('rag.storage_gib', 'decimal', '0'::jsonb, true, 'rag', 'Managed RAG storage quota in GiB.'),
    ('rag.queries.monthly', 'integer', '0'::jsonb, true, 'rag', 'Monthly managed RAG query quota.'),
    ('rag.chatbots.max', 'integer', '0'::jsonb, true, 'rag', 'Maximum managed chatbots.'),
    ('rag.max_kbs_per_bot', 'integer', '5'::jsonb, true, 'rag', 'Maximum knowledge bases attached to one chatbot.'),
    ('rag.public_bots.enabled', 'boolean', 'false'::jsonb, true, 'rag', 'Public chatbot publication access.'),
    ('rag.widget.enabled', 'boolean', 'false'::jsonb, true, 'rag', 'Hosted chatbot widget access.')
ON CONFLICT (feature_key) DO UPDATE SET
    value_type = EXCLUDED.value_type,
    default_value = EXCLUDED.default_value,
    allow_override = EXCLUDED.allow_override,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    updated_at = now();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM feature_catalog WHERE feature_key LIKE 'rag.%';
DROP TABLE IF EXISTS rag_audit_events;
DROP TABLE IF EXISTS rag_rate_catalog;
DROP TABLE IF EXISTS rag_usage_events;
DROP TABLE IF EXISTS chatbot_messages;
DROP TABLE IF EXISTS chatbot_sessions;
ALTER TABLE IF EXISTS chatbots DROP CONSTRAINT IF EXISTS chatbots_published_version_fk;
DROP TABLE IF EXISTS chatbot_versions;
DROP TABLE IF EXISTS chatbots;
DROP TRIGGER IF EXISTS tenants_queue_rag_purge ON tenants;
DROP FUNCTION IF EXISTS queue_rag_tenant_purge();
DROP TABLE IF EXISTS rag_purge_jobs;
DROP TABLE IF EXISTS rag_ingestion_jobs;
ALTER TABLE IF EXISTS knowledge_base_documents DROP CONSTRAINT IF EXISTS knowledge_base_documents_active_version_fk;
ALTER TABLE IF EXISTS knowledge_base_documents DROP CONSTRAINT IF EXISTS knowledge_base_documents_processing_version_fk;
DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS knowledge_base_documents;
DROP TABLE IF EXISTS knowledge_bases;
DROP TABLE IF EXISTS rag_profiles;
-- +goose StatementEnd

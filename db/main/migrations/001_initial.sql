-- Migration: 001_initial
-- Source: generated from current live database schema (pg_dump --schema-only)

-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

--
-- PostgreSQL database dump
--


-- Dumped from database version 18.3 (Debian 18.3-1.pgdg13+1)
-- Dumped by pg_dump version 18.3 (Debian 18.3-1.pgdg13+1)




--
-- Name: budgets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.budgets (
    budget_id uuid DEFAULT gen_random_uuid() NOT NULL,
    max_budget numeric(10,4),
    soft_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    max_parallel_requests integer,
    budget_duration character varying(50),
    budget_reset_at timestamp with time zone,
    model_max_budget jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.credentials (
    credential_id uuid DEFAULT gen_random_uuid() NOT NULL,
    credential_name character varying(255) NOT NULL,
    credential_values jsonb NOT NULL,
    credential_info jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: end_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.end_users (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_alias character varying(255),
    team_id uuid,
    budget_id uuid,
    spend numeric(10,4) DEFAULT 0.0,
    max_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    blocked boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: exchange_rates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchange_rates (
    rate_id uuid DEFAULT gen_random_uuid() NOT NULL,
    usd_to_irr double precision NOT NULL,
    source character varying(255) NOT NULL,
    fetched_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: guardrails; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guardrails (
    guardrail_name character varying(255) NOT NULL,
    guardrail_type character varying(50),
    event_hook character varying(50),
    guardrail_init_kwargs jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: model_aliases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.model_aliases (
    id integer NOT NULL,
    aliases jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: model_aliases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.model_aliases_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: model_aliases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.model_aliases_id_seq OWNED BY public.model_aliases.id;


--
-- Name: model_deployments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.model_deployments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    model_name character varying(255) NOT NULL,
    provider character varying(50) NOT NULL,
    model character varying(255) NOT NULL,
    provider_config jsonb NOT NULL,
    model_info jsonb,
    rpm bigint,
    tpm bigint,
    timeout double precision,
    stream_timeout double precision,
    health_check_timeout integer DEFAULT 60,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id uuid
);


--
-- Name: object_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.object_permissions (
    object_permission_id uuid DEFAULT gen_random_uuid() NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: organization_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_memberships (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying(50) DEFAULT 'member'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    organization_id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_alias character varying(255) NOT NULL,
    budget_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb,
    models text[] DEFAULT '{}'::text[],
    spend numeric(10,4) DEFAULT 0.0,
    model_spend jsonb DEFAULT '{}'::jsonb,
    object_permission_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: policies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.policies (
    policy_id uuid DEFAULT gen_random_uuid() NOT NULL,
    policy_name character varying(255) NOT NULL,
    policy_type character varying(50),
    policy_config jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    project_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_alias character varying(255),
    description character varying(255),
    team_id uuid,
    budget_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb,
    models text[] DEFAULT '{}'::text[],
    spend numeric(10,4) DEFAULT 0.0,
    model_spend jsonb DEFAULT '{}'::jsonb,
    model_rpm_limit jsonb DEFAULT '{}'::jsonb,
    model_tpm_limit jsonb DEFAULT '{}'::jsonb,
    blocked boolean DEFAULT false,
    object_permission_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);


--
-- Name: request_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.request_logs (
    id bigint NOT NULL,
    request_id character varying(255),
    token_hash character varying(64),
    team_id uuid,
    user_id uuid,
    model character varying(255),
    request_body jsonb,
    response_body jsonb,
    status character varying(20),
    cost numeric(10,6),
    response_time_ms integer,
    created_at timestamp with time zone DEFAULT now(),
    tenant_id uuid
);


--
-- Name: request_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.request_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: request_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.request_logs_id_seq OWNED BY public.request_logs.id;


--
-- Name: spend_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.spend_logs (
    id bigint NOT NULL,
    token_hash character varying(64),
    team_id uuid,
    user_id uuid,
    organization_id uuid,
    end_user_id uuid,
    model character varying(255),
    input_tokens integer,
    output_tokens integer,
    total_tokens integer,
    cost numeric(10,6),
    response_time_ms integer,
    status character varying(20),
    completion_start_time timestamp with time zone,
    completion_end_time timestamp with time zone,
    request_tags text[],
    request_metadata jsonb,
    created_at timestamp with time zone DEFAULT now(),
    tenant_id uuid,
    request_id character varying(255),
    cost_usd numeric(10,6),
    cost_irr numeric(18,2)
);


--
-- Name: spend_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.spend_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: spend_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.spend_logs_id_seq OWNED BY public.spend_logs.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    tag_name character varying(255) NOT NULL,
    budget_id uuid,
    spend numeric(10,4) DEFAULT 0.0,
    max_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    models text[] DEFAULT '{}'::text[],
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: team_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_memberships (
    team_membership_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    team_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying(50) DEFAULT 'team_member'::character varying NOT NULL,
    team_member_permissions text[] DEFAULT '{}'::text[] NOT NULL,
    budget_id uuid,
    spend numeric(10,4) DEFAULT 0.0,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    team_id uuid DEFAULT gen_random_uuid() NOT NULL,
    team_alias character varying(255),
    organization_id uuid,
    object_permission_id uuid,
    admins text[] DEFAULT '{}'::text[],
    members text[] DEFAULT '{}'::text[],
    members_with_roles jsonb DEFAULT '{}'::jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    max_budget numeric(10,4),
    soft_budget numeric(10,4),
    spend numeric(10,4) DEFAULT 0.0,
    models text[] DEFAULT '{}'::text[],
    max_parallel_requests integer,
    tpm_limit bigint,
    rpm_limit bigint,
    budget_duration character varying(50),
    budget_reset_at timestamp with time zone,
    blocked boolean DEFAULT false,
    model_spend jsonb DEFAULT '{}'::jsonb,
    model_max_budget jsonb DEFAULT '{}'::jsonb,
    router_settings jsonb DEFAULT '{}'::jsonb,
    team_member_permissions text[] DEFAULT '{}'::text[],
    access_group_ids text[] DEFAULT '{}'::text[],
    policies text[] DEFAULT '{}'::text[],
    default_team_member_models text[] DEFAULT '{}'::text[],
    budget_limits jsonb,
    model_id integer,
    allow_team_guardrail_config boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: tenant_admin_captchas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_admin_captchas (
    id text NOT NULL,
    answer text NOT NULL,
    "createdAt" timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "expiresAt" timestamp(3) without time zone NOT NULL
);


--
-- Name: tenant_admin_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_admin_invitations (
    invitation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    "tenantSlug" text NOT NULL,
    email text,
    mobile text,
    role text DEFAULT 'tenant_member'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    token_hash character varying(64) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tenant_admin_invitations_check CHECK (((email IS NOT NULL) OR (mobile IS NOT NULL))),
    CONSTRAINT tenant_admin_invitations_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'accepted'::text, 'expired'::text, 'revoked'::text])))
);


--
-- Name: tenant_admin_otp_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_admin_otp_requests (
    id text NOT NULL,
    "tenantSlug" text NOT NULL,
    purpose text NOT NULL,
    channel text NOT NULL,
    identifier text NOT NULL,
    "userID" text,
    "pendingValue" text,
    "otpCode" text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    "createdAt" timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "expiresAt" timestamp(3) without time zone NOT NULL
);


--
-- Name: tenant_admin_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_admin_sessions (
    id text NOT NULL,
    "tenantSlug" text NOT NULL,
    "userID" text NOT NULL,
    "createdAt" timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "expiresAt" timestamp(3) without time zone NOT NULL
);


--
-- Name: tenant_onboarding_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_onboarding_requests (
    request_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_slug character varying(63) NOT NULL,
    tenant_domain character varying(255) NOT NULL,
    plan_slug character varying(64) DEFAULT 'free'::character varying NOT NULL,
    email text,
    mobile text,
    customer_name text,
    otp_code text NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    verified_at timestamp with time zone,
    tenant_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_tenant_onboarding_contact CHECK (((email IS NOT NULL) OR (mobile IS NOT NULL))),
    CONSTRAINT chk_tenant_onboarding_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'verified'::character varying, 'expired'::character varying, 'canceled'::character varying])::text[])))
);


--
-- Name: tenant_provider_account_access; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_provider_account_access (
    account_access_id uuid DEFAULT gen_random_uuid() NOT NULL,
    account_id uuid NOT NULL,
    role character varying(20) NOT NULL,
    principal_type character varying(20) NOT NULL,
    principal_id character varying(255),
    display_name character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    access_scope character varying(20) DEFAULT 'account'::character varying NOT NULL,
    model character varying(255),
    CONSTRAINT chk_tenant_provider_account_access_everyone CHECK (((((principal_type)::text = 'everyone'::text) AND (principal_id IS NULL)) OR (((principal_type)::text <> 'everyone'::text) AND (principal_id IS NOT NULL)))),
    CONSTRAINT chk_tenant_provider_account_access_manager_principal CHECK ((((role)::text <> 'manager'::text) OR ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying])::text[])))),
    CONSTRAINT chk_tenant_provider_account_access_principal_type CHECK (((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying, 'everyone'::character varying])::text[]))),
    CONSTRAINT chk_tenant_provider_account_access_role CHECK (((role)::text = ANY ((ARRAY['manager'::character varying, 'user'::character varying])::text[]))),
    CONSTRAINT ck_tenant_provider_account_access_model_scope CHECK (((((access_scope)::text = 'account'::text) AND (model IS NULL)) OR (((access_scope)::text = 'model'::text) AND (model IS NOT NULL) AND (btrim((model)::text) <> ''::text)))),
    CONSTRAINT ck_tenant_provider_account_access_scope CHECK (((access_scope)::text = ANY ((ARRAY['account'::character varying, 'model'::character varying])::text[])))
);


--
-- Name: tenant_provider_account_models; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_provider_account_models (
    account_model_id uuid DEFAULT gen_random_uuid() NOT NULL,
    account_id uuid NOT NULL,
    model character varying(255) NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    types text[] DEFAULT '{}'::text[] NOT NULL,
    pricing_mode character varying(20) DEFAULT 'public'::character varying NOT NULL,
    private_pricing jsonb,
    settings jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_tenant_provider_account_models_pricing_mode CHECK (((pricing_mode)::text = ANY ((ARRAY['public'::character varying, 'private'::character varying])::text[])))
);


--
-- Name: tenant_provider_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_provider_accounts (
    account_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    provider character varying(100) NOT NULL,
    name character varying(255) NOT NULL,
    auth_mode character varying(50) NOT NULL,
    provider_config jsonb DEFAULT '{}'::jsonb NOT NULL,
    credential_data jsonb,
    has_credential boolean DEFAULT false NOT NULL,
    allow_virtual_keys_only boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_tenant_provider_accounts_auth_mode CHECK (((auth_mode)::text = ANY ((ARRAY['api_key'::character varying, 'azure_entra_cert'::character varying, 'aws_sigv4'::character varying, 'google_service_account'::character varying, 'no_credential'::character varying])::text[])))
);


--
-- Name: tenants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenants (
    tenant_id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_alias character varying(255),
    email character varying(255),
    role character varying(50) DEFAULT 'internal_user'::character varying,
    team_id uuid,
    organization_id uuid,
    max_budget numeric(10,4),
    budget_duration character varying(50),
    spend numeric(10,4) DEFAULT 0.0,
    tpm_limit bigint,
    rpm_limit bigint,
    metadata jsonb DEFAULT '{}'::jsonb,
    object_permission_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: verification_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    token_hash character varying(64) NOT NULL,
    key_alias character varying(255),
    key_name character varying(255),
    team_id uuid,
    user_id uuid,
    organization_id uuid,
    project_id uuid,
    models text[] DEFAULT '{}'::text[],
    max_budget numeric(10,4),
    soft_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    max_parallel_requests integer,
    budget_duration character varying(50),
    budget_reset_at timestamp with time zone,
    spend numeric(10,4) DEFAULT 0.0,
    expires timestamp with time zone,
    blocked boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    tenant_id uuid,
    created_by_user_id text,
    last_used_at timestamp with time zone,
    revoked_at timestamp with time zone,
    revoked_by_user_id text,
    key_type character varying(16)
);


--
-- Name: virtual_model_targets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.virtual_model_targets (
    target_id uuid DEFAULT gen_random_uuid() NOT NULL,
    virtual_model_id uuid NOT NULL,
    target_order integer NOT NULL,
    provider character varying(64) NOT NULL,
    model character varying(255) NOT NULL,
    timeout_sec double precision,
    retry_attempts integer DEFAULT 2 NOT NULL,
    retry_delay_ms integer DEFAULT 100 NOT NULL,
    retry_status_codes integer[] DEFAULT ARRAY[429, 500, 502, 503] NOT NULL,
    fallback_status_codes integer[] DEFAULT ARRAY[429, 500, 502, 503, 504] NOT NULL,
    fallback_candidate boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: virtual_models; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.virtual_models (
    virtual_model_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    alias character varying(255) NOT NULL,
    routing_strategy character varying(64) DEFAULT 'priority'::character varying NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    exposed_modes text[] DEFAULT '{}'::text[] NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_by character varying(255),
    updated_by character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: wallet_topup_intents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallet_topup_intents (
    topup_intent_id uuid DEFAULT gen_random_uuid() NOT NULL,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    provider character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    amount_usd double precision NOT NULL,
    amount_irr double precision NOT NULL,
    exchange_rate double precision NOT NULL,
    tax_percent double precision NOT NULL,
    fee_percent double precision NOT NULL,
    tax_amount_irr double precision NOT NULL,
    fee_amount_irr double precision NOT NULL,
    payable_irr double precision NOT NULL,
    authority character varying(255),
    payment_url text,
    ref_id character varying(255),
    gateway_txn_id character varying(255),
    card_mask character varying(64),
    payer_phone character varying(64),
    callback_url text,
    return_url text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    raw_payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    failure_reason text,
    wallet_tx_id uuid,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallet_topup_intents_status_check CHECK (((status)::text = ANY ((ARRAY['PENDING'::character varying, 'COMPLETED'::character varying, 'FAILED'::character varying, 'CANCELED'::character varying])::text[])))
);


--
-- Name: wallet_topup_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallet_topup_settings (
    settings_id boolean DEFAULT true NOT NULL,
    min_amount_usd double precision DEFAULT 1 NOT NULL,
    max_amount_usd double precision DEFAULT 1000 NOT NULL,
    tax_percent double precision DEFAULT 0 NOT NULL,
    fee_percent double precision DEFAULT 2 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT wallet_topup_settings_check CHECK ((max_amount_usd >= min_amount_usd)),
    CONSTRAINT wallet_topup_settings_fee_percent_check CHECK ((fee_percent >= (0)::double precision)),
    CONSTRAINT wallet_topup_settings_min_amount_usd_check CHECK ((min_amount_usd > (0)::double precision)),
    CONSTRAINT wallet_topup_settings_tax_percent_check CHECK ((tax_percent >= (0)::double precision))
);


--
-- Name: wallet_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallet_transactions (
    transaction_id uuid DEFAULT gen_random_uuid() NOT NULL,
    wallet_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tx_type character varying(50) NOT NULL,
    amount_irr double precision NOT NULL,
    amount_usd double precision,
    exchange_rate double precision,
    status character varying(50) NOT NULL,
    description text,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: wallets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wallets (
    wallet_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    balance_irr double precision DEFAULT 0 NOT NULL,
    pending_irr double precision DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: model_aliases id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_aliases ALTER COLUMN id SET DEFAULT nextval('public.model_aliases_id_seq'::regclass);


--
-- Name: request_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.request_logs ALTER COLUMN id SET DEFAULT nextval('public.request_logs_id_seq'::regclass);


--
-- Name: spend_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spend_logs ALTER COLUMN id SET DEFAULT nextval('public.spend_logs_id_seq'::regclass);


--
-- Name: budgets budgets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.budgets
    ADD CONSTRAINT budgets_pkey PRIMARY KEY (budget_id);


--
-- Name: credentials credentials_credential_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credentials
    ADD CONSTRAINT credentials_credential_name_key UNIQUE (credential_name);


--
-- Name: credentials credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credentials
    ADD CONSTRAINT credentials_pkey PRIMARY KEY (credential_id);


--
-- Name: end_users end_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_pkey PRIMARY KEY (user_id);


--
-- Name: exchange_rates exchange_rates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exchange_rates
    ADD CONSTRAINT exchange_rates_pkey PRIMARY KEY (rate_id);


--
-- Name: guardrails guardrails_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guardrails
    ADD CONSTRAINT guardrails_pkey PRIMARY KEY (guardrail_name);


--
-- Name: model_aliases model_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_aliases
    ADD CONSTRAINT model_aliases_pkey PRIMARY KEY (id);


--
-- Name: model_deployments model_deployments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_deployments
    ADD CONSTRAINT model_deployments_pkey PRIMARY KEY (id);


--
-- Name: object_permissions object_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.object_permissions
    ADD CONSTRAINT object_permissions_pkey PRIMARY KEY (object_permission_id);


--
-- Name: organization_memberships organization_memberships_organization_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_organization_id_user_id_key UNIQUE (organization_id, user_id);


--
-- Name: organization_memberships organization_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (organization_id);


--
-- Name: policies policies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_pkey PRIMARY KEY (policy_id);


--
-- Name: policies policies_policy_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.policies
    ADD CONSTRAINT policies_policy_name_key UNIQUE (policy_name);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (project_id);


--
-- Name: request_logs request_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.request_logs
    ADD CONSTRAINT request_logs_pkey PRIMARY KEY (id);


--
-- Name: spend_logs spend_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spend_logs
    ADD CONSTRAINT spend_logs_pkey PRIMARY KEY (id);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (tag_name);


--
-- Name: team_memberships team_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_pkey PRIMARY KEY (team_membership_id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (team_id);


--
-- Name: tenant_admin_captchas tenant_admin_captchas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_captchas
    ADD CONSTRAINT tenant_admin_captchas_pkey PRIMARY KEY (id);


--
-- Name: tenant_admin_invitations tenant_admin_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_invitations
    ADD CONSTRAINT tenant_admin_invitations_pkey PRIMARY KEY (invitation_id);


--
-- Name: tenant_admin_invitations tenant_admin_invitations_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_invitations
    ADD CONSTRAINT tenant_admin_invitations_token_hash_key UNIQUE (token_hash);


--
-- Name: tenant_admin_otp_requests tenant_admin_otp_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_otp_requests
    ADD CONSTRAINT tenant_admin_otp_requests_pkey PRIMARY KEY (id);


--
-- Name: tenant_admin_sessions tenant_admin_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_sessions
    ADD CONSTRAINT tenant_admin_sessions_pkey PRIMARY KEY (id);


--
-- Name: tenant_onboarding_requests tenant_onboarding_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_onboarding_requests
    ADD CONSTRAINT tenant_onboarding_requests_pkey PRIMARY KEY (request_id);


--
-- Name: tenant_provider_account_access tenant_provider_account_access_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_account_access
    ADD CONSTRAINT tenant_provider_account_access_pkey PRIMARY KEY (account_access_id);


--
-- Name: tenant_provider_account_models tenant_provider_account_models_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_account_models
    ADD CONSTRAINT tenant_provider_account_models_pkey PRIMARY KEY (account_model_id);


--
-- Name: tenant_provider_accounts tenant_provider_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_accounts
    ADD CONSTRAINT tenant_provider_accounts_pkey PRIMARY KEY (account_id);


--
-- Name: tenants tenants_domain_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_domain_key UNIQUE (domain);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (tenant_id);


--
-- Name: tenants tenants_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_slug_key UNIQUE (slug);


--
-- Name: team_memberships uq_team_memberships_team_user; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT uq_team_memberships_team_user UNIQUE (team_id, user_id);


--
-- Name: tenant_provider_account_models uq_tenant_provider_account_models_account_model; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_account_models
    ADD CONSTRAINT uq_tenant_provider_account_models_account_model UNIQUE (account_id, model);


--
-- Name: tenant_provider_accounts uq_tenant_provider_accounts_tenant_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_accounts
    ADD CONSTRAINT uq_tenant_provider_accounts_tenant_name UNIQUE (tenant_id, name);


--
-- Name: virtual_model_targets uq_virtual_model_targets_order; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_model_targets
    ADD CONSTRAINT uq_virtual_model_targets_order UNIQUE (virtual_model_id, target_order);


--
-- Name: virtual_models uq_virtual_models_tenant_alias; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_models
    ADD CONSTRAINT uq_virtual_models_tenant_alias UNIQUE (tenant_id, alias);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: verification_tokens verification_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_pkey PRIMARY KEY (id);


--
-- Name: verification_tokens verification_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: virtual_model_targets virtual_model_targets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_model_targets
    ADD CONSTRAINT virtual_model_targets_pkey PRIMARY KEY (target_id);


--
-- Name: virtual_models virtual_models_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_models
    ADD CONSTRAINT virtual_models_pkey PRIMARY KEY (virtual_model_id);


--
-- Name: wallet_topup_intents wallet_topup_intents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_topup_intents
    ADD CONSTRAINT wallet_topup_intents_pkey PRIMARY KEY (topup_intent_id);


--
-- Name: wallet_topup_intents wallet_topup_intents_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_topup_intents
    ADD CONSTRAINT wallet_topup_intents_public_id_key UNIQUE (public_id);


--
-- Name: wallet_topup_settings wallet_topup_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_topup_settings
    ADD CONSTRAINT wallet_topup_settings_pkey PRIMARY KEY (settings_id);


--
-- Name: wallet_transactions wallet_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_pkey PRIMARY KEY (transaction_id);


--
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (wallet_id);


--
-- Name: wallets wallets_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_tenant_id_key UNIQUE (tenant_id);


--
-- Name: idx_end_user_team; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_end_user_team ON public.end_users USING btree (team_id);


--
-- Name: idx_exchange_rates_fetched_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exchange_rates_fetched_at ON public.exchange_rates USING btree (fetched_at DESC);


--
-- Name: idx_model_deployments_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_model_deployments_tenant_id ON public.model_deployments USING btree (tenant_id);


--
-- Name: idx_model_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_model_name ON public.model_deployments USING btree (model_name);


--
-- Name: idx_org_memberships_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_memberships_org ON public.organization_memberships USING btree (organization_id);


--
-- Name: idx_org_memberships_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_memberships_user ON public.organization_memberships USING btree (user_id);


--
-- Name: idx_spend_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spend_created ON public.spend_logs USING btree (created_at);


--
-- Name: idx_spend_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spend_org ON public.spend_logs USING btree (organization_id);


--
-- Name: idx_spend_team; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spend_team ON public.spend_logs USING btree (team_id);


--
-- Name: idx_spend_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spend_token ON public.spend_logs USING btree (token_hash);


--
-- Name: idx_spend_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_spend_user ON public.spend_logs USING btree (user_id);


--
-- Name: idx_team_alias; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_alias ON public.teams USING btree (team_alias);


--
-- Name: idx_team_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_created_at ON public.teams USING btree (created_at);


--
-- Name: idx_team_memberships_team; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_memberships_team ON public.team_memberships USING btree (team_id);


--
-- Name: idx_team_memberships_tenant; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_memberships_tenant ON public.team_memberships USING btree (tenant_id);


--
-- Name: idx_team_memberships_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_memberships_user ON public.team_memberships USING btree (user_id);


--
-- Name: idx_team_org_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_org_id ON public.teams USING btree (organization_id);


--
-- Name: idx_tenant_onboarding_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_onboarding_expires_at ON public.tenant_onboarding_requests USING btree (expires_at);


--
-- Name: idx_tenant_onboarding_slug_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_onboarding_slug_status ON public.tenant_onboarding_requests USING btree (tenant_slug, status);


--
-- Name: idx_tenant_provider_account_access_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_provider_account_access_account_id ON public.tenant_provider_account_access USING btree (account_id);


--
-- Name: idx_tenant_provider_account_access_scope; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_provider_account_access_scope ON public.tenant_provider_account_access USING btree (account_id, access_scope);


--
-- Name: idx_tenant_provider_account_models_account_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_provider_account_models_account_id ON public.tenant_provider_account_models USING btree (account_id);


--
-- Name: idx_tenant_provider_accounts_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_provider_accounts_tenant_id ON public.tenant_provider_accounts USING btree (tenant_id);


--
-- Name: idx_tenant_provider_accounts_tenant_provider; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_provider_accounts_tenant_provider ON public.tenant_provider_accounts USING btree (tenant_id, provider);


--
-- Name: idx_verification_tokens_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_verification_tokens_tenant_id ON public.verification_tokens USING btree (tenant_id);


--
-- Name: idx_virtual_model_targets_model; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_virtual_model_targets_model ON public.virtual_model_targets USING btree (model);


--
-- Name: idx_virtual_models_alias; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_virtual_models_alias ON public.virtual_models USING btree (alias);


--
-- Name: idx_virtual_models_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_virtual_models_tenant_id ON public.virtual_models USING btree (tenant_id);


--
-- Name: idx_vt_key_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vt_key_type ON public.verification_tokens USING btree (key_type);


--
-- Name: idx_vt_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vt_org ON public.verification_tokens USING btree (organization_id);


--
-- Name: idx_vt_team; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vt_team ON public.verification_tokens USING btree (team_id);


--
-- Name: idx_vt_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vt_token_hash ON public.verification_tokens USING btree (token_hash);


--
-- Name: idx_vt_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_vt_user ON public.verification_tokens USING btree (user_id);


--
-- Name: idx_wallet_topup_intents_authority_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_wallet_topup_intents_authority_unique ON public.wallet_topup_intents USING btree (authority) WHERE (authority IS NOT NULL);


--
-- Name: idx_wallet_topup_intents_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_topup_intents_status ON public.wallet_topup_intents USING btree (status);


--
-- Name: idx_wallet_topup_intents_tenant_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_topup_intents_tenant_created ON public.wallet_topup_intents USING btree (tenant_id, created_at DESC);


--
-- Name: idx_wallet_transactions_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wallet_transactions_tenant_id ON public.wallet_transactions USING btree (tenant_id);


--
-- Name: tenant_admin_captcha_expires_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_captcha_expires_at_idx ON public.tenant_admin_captchas USING btree ("expiresAt");


--
-- Name: tenant_admin_invite_expiry_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_invite_expiry_idx ON public.tenant_admin_invitations USING btree (expires_at);


--
-- Name: tenant_admin_invite_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_invite_status_idx ON public.tenant_admin_invitations USING btree (status);


--
-- Name: tenant_admin_invite_tenant_slug_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_invite_tenant_slug_idx ON public.tenant_admin_invitations USING btree ("tenantSlug");


--
-- Name: tenant_admin_otp_expires_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_otp_expires_at_idx ON public.tenant_admin_otp_requests USING btree ("expiresAt");


--
-- Name: tenant_admin_otp_tenant_purpose_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_otp_tenant_purpose_idx ON public.tenant_admin_otp_requests USING btree ("tenantSlug", purpose);


--
-- Name: tenant_admin_session_expires_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_session_expires_at_idx ON public.tenant_admin_sessions USING btree ("expiresAt");


--
-- Name: tenant_admin_session_tenant_slug_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tenant_admin_session_tenant_slug_idx ON public.tenant_admin_sessions USING btree ("tenantSlug");


--
-- Name: uq_tenant_provider_account_access_principal; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_tenant_provider_account_access_principal ON public.tenant_provider_account_access USING btree (account_id, role, principal_type, COALESCE(principal_id, ''::character varying));


--
-- Name: end_users end_users_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);


--
-- Name: end_users end_users_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);


--
-- Name: model_deployments model_deployments_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_deployments
    ADD CONSTRAINT model_deployments_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id);


--
-- Name: organization_memberships organization_memberships_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(organization_id) ON DELETE CASCADE;


--
-- Name: organization_memberships organization_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_memberships
    ADD CONSTRAINT organization_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: organizations organizations_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);


--
-- Name: organizations organizations_object_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_object_permission_id_fkey FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);


--
-- Name: projects projects_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);


--
-- Name: projects projects_object_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_object_permission_id_fkey FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);


--
-- Name: projects projects_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);


--
-- Name: request_logs request_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.request_logs
    ADD CONSTRAINT request_logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id);


--
-- Name: spend_logs spend_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.spend_logs
    ADD CONSTRAINT spend_logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id);


--
-- Name: tags tags_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);


--
-- Name: team_memberships team_memberships_budget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);


--
-- Name: team_memberships team_memberships_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;


--
-- Name: team_memberships team_memberships_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: team_memberships team_memberships_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_memberships
    ADD CONSTRAINT team_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;


--
-- Name: teams teams_model_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_model_id_fkey FOREIGN KEY (model_id) REFERENCES public.model_aliases(id);


--
-- Name: teams teams_object_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_object_permission_id_fkey FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);


--
-- Name: teams teams_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(organization_id);


--
-- Name: tenant_admin_invitations tenant_admin_invitations_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admin_invitations
    ADD CONSTRAINT tenant_admin_invitations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: tenant_onboarding_requests tenant_onboarding_requests_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_onboarding_requests
    ADD CONSTRAINT tenant_onboarding_requests_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE SET NULL;


--
-- Name: tenant_provider_account_access tenant_provider_account_access_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_account_access
    ADD CONSTRAINT tenant_provider_account_access_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.tenant_provider_accounts(account_id) ON DELETE CASCADE;


--
-- Name: tenant_provider_account_models tenant_provider_account_models_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_account_models
    ADD CONSTRAINT tenant_provider_account_models_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.tenant_provider_accounts(account_id) ON DELETE CASCADE;


--
-- Name: tenant_provider_accounts tenant_provider_accounts_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_provider_accounts
    ADD CONSTRAINT tenant_provider_accounts_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: users users_object_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_object_permission_id_fkey FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);


--
-- Name: users users_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(organization_id);


--
-- Name: users users_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);


--
-- Name: verification_tokens verification_tokens_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(organization_id);


--
-- Name: verification_tokens verification_tokens_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id);


--
-- Name: verification_tokens verification_tokens_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);


--
-- Name: verification_tokens verification_tokens_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id);


--
-- Name: verification_tokens verification_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: virtual_model_targets virtual_model_targets_virtual_model_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_model_targets
    ADD CONSTRAINT virtual_model_targets_virtual_model_id_fkey FOREIGN KEY (virtual_model_id) REFERENCES public.virtual_models(virtual_model_id) ON DELETE CASCADE;


--
-- Name: virtual_models virtual_models_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.virtual_models
    ADD CONSTRAINT virtual_models_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: wallet_topup_intents wallet_topup_intents_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_topup_intents
    ADD CONSTRAINT wallet_topup_intents_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: wallet_topup_intents wallet_topup_intents_wallet_tx_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_topup_intents
    ADD CONSTRAINT wallet_topup_intents_wallet_tx_id_fkey FOREIGN KEY (wallet_tx_id) REFERENCES public.wallet_transactions(transaction_id) ON DELETE SET NULL;


--
-- Name: wallet_transactions wallet_transactions_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- Name: wallet_transactions wallet_transactions_wallet_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallet_transactions
    ADD CONSTRAINT wallet_transactions_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(wallet_id) ON DELETE CASCADE;


--
-- Name: wallets wallets_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--




-- +goose Down
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

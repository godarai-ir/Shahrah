-- +goose Up
-- Remove unused tables inherited from the legacy access-control and policy model.
-- Refuse to discard unexpected data or still-linked entity fields.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.credentials) THEN
        RAISE EXCEPTION 'refusing to drop public.credentials because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.guardrails) THEN
        RAISE EXCEPTION 'refusing to drop public.guardrails because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.policies) THEN
        RAISE EXCEPTION 'refusing to drop public.policies because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.model_aliases) THEN
        RAISE EXCEPTION 'refusing to drop public.model_aliases because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.object_permissions) THEN
        RAISE EXCEPTION 'refusing to drop public.object_permissions because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.teams WHERE model_id IS NOT NULL) THEN
        RAISE EXCEPTION 'refusing to drop teams.model_id because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.organizations WHERE object_permission_id IS NOT NULL) THEN
        RAISE EXCEPTION 'refusing to drop organizations.object_permission_id because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.teams WHERE object_permission_id IS NOT NULL) THEN
        RAISE EXCEPTION 'refusing to drop teams.object_permission_id because it contains data';
    END IF;
    IF EXISTS (SELECT 1 FROM public.users WHERE object_permission_id IS NOT NULL) THEN
        RAISE EXCEPTION 'refusing to drop users.object_permission_id because it contains data';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE ONLY public.organizations
    DROP CONSTRAINT IF EXISTS organizations_object_permission_id_fkey,
    DROP COLUMN IF EXISTS object_permission_id;

ALTER TABLE ONLY public.teams
    DROP CONSTRAINT IF EXISTS teams_model_id_fkey,
    DROP CONSTRAINT IF EXISTS teams_object_permission_id_fkey,
    DROP COLUMN IF EXISTS model_id,
    DROP COLUMN IF EXISTS object_permission_id;

ALTER TABLE ONLY public.users
    DROP CONSTRAINT IF EXISTS users_object_permission_id_fkey,
    DROP COLUMN IF EXISTS object_permission_id;

DROP TABLE public.credentials;
DROP TABLE public.guardrails;
DROP TABLE public.policies;
DROP TABLE public.model_aliases;
DROP TABLE public.object_permissions;

-- +goose Down
CREATE TABLE public.credentials (
    credential_id uuid DEFAULT gen_random_uuid() NOT NULL,
    credential_name character varying(255) NOT NULL,
    credential_values jsonb NOT NULL,
    credential_info jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255),
    CONSTRAINT credentials_pkey PRIMARY KEY (credential_id),
    CONSTRAINT credentials_credential_name_key UNIQUE (credential_name)
);

CREATE TABLE public.guardrails (
    guardrail_name character varying(255) NOT NULL,
    guardrail_type character varying(50),
    event_hook character varying(50),
    guardrail_init_kwargs jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT guardrails_pkey PRIMARY KEY (guardrail_name)
);

CREATE TABLE public.policies (
    policy_id uuid DEFAULT gen_random_uuid() NOT NULL,
    policy_name character varying(255) NOT NULL,
    policy_type character varying(50),
    policy_config jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT policies_pkey PRIMARY KEY (policy_id),
    CONSTRAINT policies_policy_name_key UNIQUE (policy_name)
);

CREATE TABLE public.model_aliases (
    id integer NOT NULL,
    aliases jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255),
    CONSTRAINT model_aliases_pkey PRIMARY KEY (id)
);

CREATE SEQUENCE public.model_aliases_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.model_aliases_id_seq OWNED BY public.model_aliases.id;
ALTER TABLE ONLY public.model_aliases
    ALTER COLUMN id SET DEFAULT nextval('public.model_aliases_id_seq'::regclass);

CREATE TABLE public.object_permissions (
    object_permission_id uuid DEFAULT gen_random_uuid() NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255),
    CONSTRAINT object_permissions_pkey PRIMARY KEY (object_permission_id)
);

ALTER TABLE ONLY public.organizations
    ADD COLUMN object_permission_id uuid,
    ADD CONSTRAINT organizations_object_permission_id_fkey
        FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);

ALTER TABLE ONLY public.teams
    ADD COLUMN model_id integer,
    ADD COLUMN object_permission_id uuid,
    ADD CONSTRAINT teams_model_id_fkey
        FOREIGN KEY (model_id) REFERENCES public.model_aliases(id),
    ADD CONSTRAINT teams_object_permission_id_fkey
        FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);

ALTER TABLE ONLY public.users
    ADD COLUMN object_permission_id uuid,
    ADD CONSTRAINT users_object_permission_id_fkey
        FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);

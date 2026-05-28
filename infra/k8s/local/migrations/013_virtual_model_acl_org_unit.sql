-- +goose Up
ALTER TABLE public.tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_principal_type;

ALTER TABLE public.tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_principal_type
    CHECK (
        (principal_type)::text = ANY (
            (ARRAY[
                'user'::character varying,
                'team'::character varying,
                'org_unit'::character varying,
                'everyone'::character varying
            ])::text[]
        )
    );

ALTER TABLE public.tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

ALTER TABLE public.tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_manager_principal
    CHECK (
        ((role)::text <> 'manager'::text)
        OR (
            (principal_type)::text = ANY (
                (ARRAY[
                    'user'::character varying,
                    'team'::character varying,
                    'org_unit'::character varying
                ])::text[]
            )
        )
    );

CREATE TABLE IF NOT EXISTS public.virtual_model_access (
    vm_access_id uuid DEFAULT gen_random_uuid() NOT NULL,
    virtual_model_id uuid NOT NULL,
    role character varying(20) NOT NULL,
    principal_type character varying(20) NOT NULL,
    principal_id character varying(255),
    display_name character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT virtual_model_access_role_chk
        CHECK (((role)::text = ANY ((ARRAY['manager'::character varying, 'user'::character varying])::text[]))),
    CONSTRAINT virtual_model_access_principal_type_chk
        CHECK (
            ((principal_type)::text = ANY (
                (ARRAY[
                    'user'::character varying,
                    'team'::character varying,
                    'org_unit'::character varying,
                    'everyone'::character varying
                ])::text[]
            ))
        ),
    CONSTRAINT virtual_model_access_principal_required_chk
        CHECK (
            (
                ((principal_type)::text = 'everyone'::text)
                AND principal_id IS NULL
            )
            OR (
                ((principal_type)::text <> 'everyone'::text)
                AND principal_id IS NOT NULL
            )
        ),
    CONSTRAINT virtual_model_access_manager_principal_chk
        CHECK (
            ((role)::text <> 'manager'::text)
            OR (
                (principal_type)::text = ANY (
                    (ARRAY[
                        'user'::character varying,
                        'team'::character varying,
                        'org_unit'::character varying
                    ])::text[]
                )
            )
        )
);

ALTER TABLE ONLY public.virtual_model_access
    ADD CONSTRAINT virtual_model_access_pkey PRIMARY KEY (vm_access_id);

ALTER TABLE ONLY public.virtual_model_access
    ADD CONSTRAINT virtual_model_access_virtual_model_id_fkey
    FOREIGN KEY (virtual_model_id)
    REFERENCES public.virtual_models(virtual_model_id)
    ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_virtual_model_access_principal
    ON public.virtual_model_access USING btree
    (virtual_model_id, role, principal_type, COALESCE(principal_id, ''::character varying));

CREATE INDEX IF NOT EXISTS idx_virtual_model_access_virtual_model
    ON public.virtual_model_access USING btree (virtual_model_id);

-- +goose Down
DROP INDEX IF EXISTS idx_virtual_model_access_virtual_model;
DROP INDEX IF EXISTS uq_virtual_model_access_principal;
DROP TABLE IF EXISTS public.virtual_model_access;

ALTER TABLE public.tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

ALTER TABLE public.tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_manager_principal
    CHECK (
        ((role)::text <> 'manager'::text)
        OR ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying])::text[]))
    );

ALTER TABLE public.tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_principal_type;

ALTER TABLE public.tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_principal_type
    CHECK (
        ((principal_type)::text = ANY (
            (ARRAY[
                'user'::character varying,
                'team'::character varying,
                'everyone'::character varying
            ])::text[]
        ))
    );

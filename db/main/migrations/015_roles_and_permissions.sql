-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.roles (
    role_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(64) NOT NULL,
    display_name character varying(255) NOT NULL,
    description text,
    is_system boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT roles_pkey PRIMARY KEY (role_id),
    CONSTRAINT roles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_tenant_name
    ON public.roles (tenant_id, name);

CREATE TABLE IF NOT EXISTS public.role_permissions (
    role_permission_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    role_id uuid NOT NULL,
    permission character varying(128) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT role_permissions_pkey PRIMARY KEY (role_permission_id),
    CONSTRAINT role_permissions_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(role_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_role_permissions_role_permission_ci
    ON public.role_permissions (role_id, LOWER(permission));

CREATE INDEX IF NOT EXISTS idx_role_permissions_tenant_id
    ON public.role_permissions (tenant_id);

CREATE TABLE IF NOT EXISTS public.user_role_bindings (
    binding_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    resource_type character varying(32) DEFAULT 'tenant' NOT NULL,
    resource_id uuid,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_role_bindings_pkey PRIMARY KEY (binding_id),
    CONSTRAINT user_role_bindings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT user_role_bindings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT user_role_bindings_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(role_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_role_bindings_active
    ON public.user_role_bindings (tenant_id, user_id, role_id, resource_type, COALESCE(resource_id::text, ''))
    WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_user_role_bindings_tenant_user_active
    ON public.user_role_bindings (tenant_id, user_id, is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_role_bindings_tenant_user_active;
DROP INDEX IF EXISTS uq_user_role_bindings_active;
DROP TABLE IF EXISTS public.user_role_bindings;

DROP INDEX IF EXISTS idx_role_permissions_tenant_id;
DROP INDEX IF EXISTS uq_role_permissions_role_permission_ci;
DROP TABLE IF EXISTS public.role_permissions;

DROP INDEX IF EXISTS uq_roles_tenant_name;
DROP TABLE IF EXISTS public.roles;
-- +goose StatementEnd

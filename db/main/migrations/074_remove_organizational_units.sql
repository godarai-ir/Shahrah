-- +goose Up
-- Removes the organizational unit concept (product decision, 2026-10-19).
-- An org unit was an alias over public.organizations, so this migration deletes
-- everything that referenced it and then drops the legacy organizations and
-- organization_memberships tables together with every organization_id and
-- org_unit_id column. The control-plane code that read these columns was
-- removed first; ClickHouse keeps its historical columns untouched.
--
-- This migration deletes data and is not reversible.

-- +goose StatementBegin
DO $$
DECLARE
    removed bigint;
BEGIN
    -- Access-control rows that grant to an org unit.
    DELETE FROM public.tenant_provider_account_access WHERE principal_type = 'org_unit';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org_unit rows from tenant_provider_account_access', removed;

    DELETE FROM public.virtual_model_access WHERE principal_type = 'org_unit';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org_unit rows from virtual_model_access', removed;

    -- Budget and rate-limit rules scoped to an org unit can no longer match.
    DELETE FROM public.budget_limit_rules
    WHERE budget_applies_per = 'org_unit'
       OR conditions_json::text ~* '(organization_unit|org_unit)';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org-unit scoped budget_limit_rules', removed;

    DELETE FROM public.rate_limit_rules
    WHERE conditions_json::text ~* '(organization_unit|org_unit)';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org-unit scoped rate_limit_rules', removed;

    -- Org-unit permissions and role bindings scoped to an org unit.
    DELETE FROM public.role_permissions WHERE LOWER(permission) LIKE 'org\_unit:%';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org_unit role permissions', removed;

    DELETE FROM public.user_role_bindings WHERE resource_type = 'org_unit';
    GET DIAGNOSTICS removed = ROW_COUNT;
    RAISE NOTICE 'removed % org_unit role bindings', removed;

    -- The plan feature that gated the concept.
    DELETE FROM public.tenant_entitlement_overrides WHERE feature_key = 'org_units.enabled';
    DELETE FROM public.plan_entitlements WHERE feature_key = 'org_units.enabled';
    DELETE FROM public.feature_catalog WHERE feature_key = 'org_units.enabled';
END
$$;
-- +goose StatementEnd

ALTER TABLE public.tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_principal_type,
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

ALTER TABLE public.tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_principal_type
    CHECK ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying, 'everyone'::character varying])::text[])),
    ADD CONSTRAINT chk_tenant_provider_account_access_manager_principal
    CHECK (((role)::text <> 'manager'::text) OR ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying])::text[])));

ALTER TABLE public.virtual_model_access
    DROP CONSTRAINT IF EXISTS virtual_model_access_principal_type_chk,
    DROP CONSTRAINT IF EXISTS virtual_model_access_manager_principal_chk;

ALTER TABLE public.virtual_model_access
    ADD CONSTRAINT virtual_model_access_principal_type_chk
    CHECK ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying, 'everyone'::character varying])::text[])),
    ADD CONSTRAINT virtual_model_access_manager_principal_chk
    CHECK (((role)::text <> 'manager'::text) OR ((principal_type)::text = ANY ((ARRAY['user'::character varying, 'team'::character varying])::text[])));

-- Dropping a column also drops the foreign keys and indexes that use it.
ALTER TABLE public.teams
    DROP COLUMN IF EXISTS org_unit_id,
    DROP COLUMN IF EXISTS organization_id;

ALTER TABLE public.virtual_users DROP COLUMN IF EXISTS org_unit_id;
ALTER TABLE public.users DROP COLUMN IF EXISTS organization_id;
ALTER TABLE public.verification_tokens DROP COLUMN IF EXISTS organization_id;
ALTER TABLE public.spend_logs
    DROP COLUMN IF EXISTS organization_id,
    DROP COLUMN IF EXISTS org_unit_id;
ALTER TABLE public.credit_ledger DROP COLUMN IF EXISTS org_unit_id;

DROP TABLE IF EXISTS public.organization_memberships;
DROP TABLE IF EXISTS public.organizations;

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'migration 074_remove_organizational_units deletes data and cannot be rolled back';
END
$$;
-- +goose StatementEnd

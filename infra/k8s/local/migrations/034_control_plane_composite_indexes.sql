-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vt_pat_owner_created_active
    ON public.verification_tokens (tenant_id, (COALESCE(created_by_user_id, ''::text)), created_at DESC)
    WHERE key_type = 'pat' AND revoked_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vt_vat_virtual_user_created
    ON public.verification_tokens (tenant_id, virtual_user_id, created_at DESC)
    WHERE key_type = 'vat';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_team_memberships_tenant_user_team
    ON public.team_memberships (tenant_id, (user_id::text), (team_id::text));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_memberships_tenant_user_org
    ON public.organization_memberships (tenant_id, (user_id::text), (organization_id::text));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_teams_tenant_org_unit_team
    ON public.teams (tenant_id, (org_unit_id::text), (team_id::text));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tpa_tenant_created_desc
    ON public.tenant_provider_accounts (tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tpam_account_enabled_created_model
    ON public.tenant_provider_account_models (account_id, created_at ASC, model ASC)
    INCLUDE (types, pricing_mode, private_pricing)
    WHERE enabled = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tpaa_account_role_principal
    ON public.tenant_provider_account_access (account_id, role, principal_type, principal_id, access_scope, model);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wallet_transactions_tenant_created_desc
    ON public.wallet_transactions (tenant_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_spend_logs_tenant_request_latest
    ON public.spend_logs (tenant_id, request_id, completion_end_time DESC NULLS LAST, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_teams_tenant_updated_created
    ON public.teams (tenant_id, updated_at DESC, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orgs_tenant_updated_created
    ON public.organizations (tenant_id, updated_at DESC, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_virtual_users_tenant_updated_created
    ON public.virtual_users (tenant_id, updated_at DESC, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rate_limit_rules_tenant_priority_updated
    ON public.rate_limit_rules (tenant_id, priority ASC, updated_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_budget_limit_rules_tenant_priority_updated
    ON public.budget_limit_rules (tenant_id, priority ASC, updated_at DESC);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS public.idx_budget_limit_rules_tenant_priority_updated;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_rate_limit_rules_tenant_priority_updated;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_virtual_users_tenant_updated_created;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_orgs_tenant_updated_created;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_teams_tenant_updated_created;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_spend_logs_tenant_request_latest;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_wallet_transactions_tenant_created_desc;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_tpaa_account_role_principal;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_tpam_account_enabled_created_model;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_tpa_tenant_created_desc;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_teams_tenant_org_unit_team;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_org_memberships_tenant_user_org;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_team_memberships_tenant_user_team;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_vt_vat_virtual_user_created;
DROP INDEX CONCURRENTLY IF EXISTS public.idx_vt_pat_owner_created_active;

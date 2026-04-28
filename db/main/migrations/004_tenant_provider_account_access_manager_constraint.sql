-- Migration: 004_tenant_provider_account_access_manager_constraint.sql
-- Description: Restrict manager access role principals to user/team only.

ALTER TABLE tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

ALTER TABLE tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_manager_principal
    CHECK (
        role <> 'manager'
        OR principal_type IN ('user', 'team')
    );

-- 004_tenant_provider_account_access_manager_constraint

-- +goose Up

-- Migration: 004_tenant_provider_account_access_manager_constraint
-- Description: Restrict manager access role principals to user/team only.

ALTER TABLE tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

ALTER TABLE tenant_provider_account_access
    ADD CONSTRAINT chk_tenant_provider_account_access_manager_principal
    CHECK (
        role <> 'manager'
        OR principal_type IN ('user', 'team')
    );

-- +goose Down

-- Migration: 004_tenant_provider_account_access_manager_constraint (DOWN)
-- Description: Remove manager access role principal constraint.

ALTER TABLE tenant_provider_account_access
    DROP CONSTRAINT IF EXISTS chk_tenant_provider_account_access_manager_principal;

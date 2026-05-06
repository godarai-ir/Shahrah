-- 014_tenant_provider_account_access_hybrid_scope

-- +goose Up
-- +goose StatementBegin

-- Migration: 014_tenant_provider_account_access_hybrid_scope
-- Description: Prepare provider account access rules for optional per-model scope.

ALTER TABLE tenant_provider_account_access
  ADD COLUMN IF NOT EXISTS access_scope VARCHAR(20) NOT NULL DEFAULT 'account',
  ADD COLUMN IF NOT EXISTS model VARCHAR(255);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_tenant_provider_account_access_scope'
  ) THEN
    ALTER TABLE tenant_provider_account_access
      ADD CONSTRAINT ck_tenant_provider_account_access_scope
      CHECK (access_scope IN ('account', 'model'));
  END IF;
END
$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_tenant_provider_account_access_model_scope'
  ) THEN
    ALTER TABLE tenant_provider_account_access
      ADD CONSTRAINT ck_tenant_provider_account_access_model_scope
      CHECK (
        (access_scope = 'account' AND model IS NULL)
        OR
        (access_scope = 'model' AND model IS NOT NULL AND btrim(model) <> '')
      );
  END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_tenant_provider_account_access_scope
  ON tenant_provider_account_access(account_id, access_scope);

-- +goose StatementEnd

-- +goose Down

-- Migration: 014_tenant_provider_account_access_hybrid_scope (DOWN)

DROP INDEX IF EXISTS idx_tenant_provider_account_access_scope;

ALTER TABLE tenant_provider_account_access
  DROP CONSTRAINT IF EXISTS ck_tenant_provider_account_access_model_scope;

ALTER TABLE tenant_provider_account_access
  DROP CONSTRAINT IF EXISTS ck_tenant_provider_account_access_scope;

ALTER TABLE tenant_provider_account_access
  DROP COLUMN IF EXISTS model;

ALTER TABLE tenant_provider_account_access
  DROP COLUMN IF EXISTS access_scope;

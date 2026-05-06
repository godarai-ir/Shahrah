-- 015_remove_tenant_admin_users

-- +goose Up
-- +goose StatementBegin

-- Migration: 015_remove_tenant_admin_users
-- Description: Remove tenant_admin_users and detach tenant-admin session/otp tables from that local auth user store.

ALTER TABLE IF EXISTS tenant_admin_sessions
    DROP CONSTRAINT IF EXISTS tenant_admin_sessions_userID_fkey;

ALTER TABLE IF EXISTS tenant_admin_otp_requests
    DROP CONSTRAINT IF EXISTS tenant_admin_otp_requests_userID_fkey;

DROP INDEX IF EXISTS tenant_admin_user_tenant_email_uq;
DROP INDEX IF EXISTS tenant_admin_user_tenant_mobile_uq;
DROP INDEX IF EXISTS tenant_admin_user_tenant_slug_idx;

DROP TABLE IF EXISTS tenant_admin_users;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Migration: 015_remove_tenant_admin_users (DOWN)
-- Description: Recreate tenant_admin_users and restore FK/index dependencies.

CREATE TABLE IF NOT EXISTS tenant_admin_users (
    id TEXT PRIMARY KEY,
    "tenantSlug" TEXT NOT NULL,
    email TEXT,
    mobile TEXT,
    role TEXT NOT NULL DEFAULT 'tenant_member',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "lastLoginAt" TIMESTAMP(3)
);

CREATE UNIQUE INDEX IF NOT EXISTS tenant_admin_user_tenant_email_uq
    ON tenant_admin_users ("tenantSlug", email);

CREATE UNIQUE INDEX IF NOT EXISTS tenant_admin_user_tenant_mobile_uq
    ON tenant_admin_users ("tenantSlug", mobile);

CREATE INDEX IF NOT EXISTS tenant_admin_user_tenant_slug_idx
    ON tenant_admin_users ("tenantSlug");

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tenant_admin_sessions_userID_fkey'
    ) THEN
        ALTER TABLE tenant_admin_sessions
            ADD CONSTRAINT tenant_admin_sessions_userID_fkey
            FOREIGN KEY ("userID") REFERENCES tenant_admin_users(id)
            ON DELETE CASCADE ON UPDATE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tenant_admin_otp_requests_userID_fkey'
    ) THEN
        ALTER TABLE tenant_admin_otp_requests
            ADD CONSTRAINT tenant_admin_otp_requests_userID_fkey
            FOREIGN KEY ("userID") REFERENCES tenant_admin_users(id)
            ON DELETE SET NULL ON UPDATE CASCADE;
    END IF;
END $$;

-- +goose StatementEnd


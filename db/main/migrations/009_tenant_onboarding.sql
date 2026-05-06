-- 009_tenant_onboarding

-- +goose Up

-- Migration: 009_tenant_onboarding
-- Description: Explicit tenant onboarding reservations and OTP verification.

-- Remove 'default' tenant that was auto-created for development
DELETE FROM wallets
WHERE tenant_id IN (SELECT tenant_id FROM tenants WHERE slug = 'default');

DELETE FROM tenants
WHERE slug = 'default';

CREATE TABLE IF NOT EXISTS tenant_onboarding_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_slug VARCHAR(63) NOT NULL,
    tenant_domain VARCHAR(255) NOT NULL,
    plan_slug VARCHAR(64) NOT NULL DEFAULT 'free',
    email TEXT,
    mobile TEXT,
    customer_name TEXT,
    otp_code TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    tenant_id UUID REFERENCES tenants(tenant_id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tenant_onboarding_contact CHECK (email IS NOT NULL OR mobile IS NOT NULL),
    CONSTRAINT chk_tenant_onboarding_status CHECK (status IN ('pending', 'verified', 'expired', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_onboarding_slug_status
    ON tenant_onboarding_requests(tenant_slug, status);

CREATE INDEX IF NOT EXISTS idx_tenant_onboarding_expires_at
    ON tenant_onboarding_requests(expires_at);

CREATE TABLE IF NOT EXISTS "tenant_admin_users" (
    "id" TEXT PRIMARY KEY,
    "tenantSlug" TEXT NOT NULL,
    "email" TEXT,
    "mobile" TEXT,
    "role" TEXT NOT NULL DEFAULT 'tenant_member',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "lastLoginAt" TIMESTAMP(3)
);

CREATE UNIQUE INDEX IF NOT EXISTS "tenant_admin_user_tenant_email_uq"
    ON "tenant_admin_users" ("tenantSlug", "email");

CREATE UNIQUE INDEX IF NOT EXISTS "tenant_admin_user_tenant_mobile_uq"
    ON "tenant_admin_users" ("tenantSlug", "mobile");

CREATE INDEX IF NOT EXISTS "tenant_admin_user_tenant_slug_idx"
    ON "tenant_admin_users" ("tenantSlug");

-- +goose Down

-- Migration: 009_tenant_onboarding (DOWN)

DROP INDEX IF EXISTS "tenant_admin_user_tenant_slug_idx";
DROP INDEX IF EXISTS "tenant_admin_user_tenant_mobile_uq";
DROP INDEX IF EXISTS "tenant_admin_user_tenant_email_uq";
DROP TABLE IF EXISTS "tenant_admin_users" CASCADE;

DROP INDEX IF EXISTS idx_tenant_onboarding_expires_at;
DROP INDEX IF EXISTS idx_tenant_onboarding_slug_status;
DROP TABLE IF EXISTS tenant_onboarding_requests CASCADE;

-- 007_tenant_admin_invitations

-- +goose Up

-- Migration: 007_tenant_admin_invitations
-- Description: Tenant-admin invitation records for user onboarding and access control.

CREATE TABLE IF NOT EXISTS "tenant_admin_invitations" (
    "invitation_id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "tenant_id" UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    "tenantSlug" TEXT NOT NULL,
    "email" TEXT,
    "mobile" TEXT,
    "role" TEXT NOT NULL DEFAULT 'tenant_member',
    "status" TEXT NOT NULL DEFAULT 'pending',
    "token_hash" VARCHAR(64) NOT NULL UNIQUE,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "accepted_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ("status" IN ('pending', 'accepted', 'expired', 'revoked')),
    CHECK ("email" IS NOT NULL OR "mobile" IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS "tenant_admin_invite_tenant_slug_idx"
ON "tenant_admin_invitations" ("tenantSlug");

CREATE INDEX IF NOT EXISTS "tenant_admin_invite_status_idx"
ON "tenant_admin_invitations" ("status");

CREATE INDEX IF NOT EXISTS "tenant_admin_invite_expiry_idx"
ON "tenant_admin_invitations" ("expires_at");

-- +goose Down

-- Migration: 007_tenant_admin_invitations (DOWN)

DROP INDEX IF EXISTS "tenant_admin_invite_expiry_idx";
DROP INDEX IF EXISTS "tenant_admin_invite_status_idx";
DROP INDEX IF EXISTS "tenant_admin_invite_tenant_slug_idx";

DROP TABLE IF EXISTS "tenant_admin_invitations" CASCADE;

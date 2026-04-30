-- Migration: 008_iam_phase2_foundations.sql
-- Description: Canonical IAM phase-2 foundations for tenant-scoped org units, team memberships, virtual users, and typed keys.

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

UPDATE organizations
SET tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_organizations_tenant_id
    ON organizations(tenant_id);

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS org_unit_id UUID REFERENCES organizations(organization_id);

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS workspace_type VARCHAR(50) NOT NULL DEFAULT 'general';

UPDATE teams
SET org_unit_id = organization_id
WHERE org_unit_id IS NULL
  AND organization_id IS NOT NULL;

UPDATE teams t
SET tenant_id = o.tenant_id
FROM organizations o
WHERE t.tenant_id IS NULL
  AND t.organization_id = o.organization_id
  AND o.tenant_id IS NOT NULL;

UPDATE teams
SET tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_teams_tenant_id
    ON teams(tenant_id);

CREATE INDEX IF NOT EXISTS idx_teams_org_unit_id
    ON teams(org_unit_id);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS mobile VARCHAR(50);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE users u
SET tenant_id = o.tenant_id
FROM organizations o
WHERE u.tenant_id IS NULL
  AND u.organization_id = o.organization_id
  AND o.tenant_id IS NOT NULL;

UPDATE users u
SET tenant_id = t.tenant_id
FROM teams t
WHERE u.tenant_id IS NULL
  AND u.team_id = t.team_id
  AND t.tenant_id IS NOT NULL;

UPDATE users
SET tenant_id = (SELECT tenant_id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_tenant_id
    ON users(tenant_id);

CREATE INDEX IF NOT EXISTS idx_users_mobile
    ON users(mobile);

ALTER TABLE organization_memberships
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

UPDATE organization_memberships om
SET tenant_id = o.tenant_id
FROM organizations o
WHERE om.tenant_id IS NULL
  AND om.organization_id = o.organization_id
  AND o.tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_org_memberships_tenant
    ON organization_memberships(tenant_id);

CREATE TABLE IF NOT EXISTS team_memberships (
    team_membership_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'team_member',
    team_member_permissions TEXT[] NOT NULL DEFAULT '{}',
    budget_id UUID REFERENCES budgets(budget_id),
    spend DECIMAL(10,4) DEFAULT 0.0,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_team_memberships_team_user UNIQUE (team_id, user_id)
);

INSERT INTO team_memberships (tenant_id, team_id, user_id, role, created_at, updated_at)
SELECT
    COALESCE(u.tenant_id, t.tenant_id, (SELECT tenant_id FROM tenants WHERE slug = 'default' LIMIT 1)),
    u.team_id,
    u.user_id,
    CASE WHEN LOWER(COALESCE(u.role, '')) = 'team_admin' THEN 'team_admin' ELSE 'team_member' END,
    NOW(),
    NOW()
FROM users u
JOIN teams t ON t.team_id = u.team_id
WHERE u.team_id IS NOT NULL
ON CONFLICT ON CONSTRAINT uq_team_memberships_team_user DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_team_memberships_tenant
    ON team_memberships(tenant_id);

CREATE INDEX IF NOT EXISTS idx_team_memberships_team
    ON team_memberships(team_id);

CREATE INDEX IF NOT EXISTS idx_team_memberships_user
    ON team_memberships(user_id);

CREATE TABLE IF NOT EXISTS virtual_users (
    virtual_user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(team_id) ON DELETE SET NULL,
    org_unit_id UUID REFERENCES organizations(organization_id) ON DELETE SET NULL,
    alias VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_virtual_users_tenant_alias UNIQUE (tenant_id, alias)
);

CREATE INDEX IF NOT EXISTS idx_virtual_users_tenant
    ON virtual_users(tenant_id);

CREATE INDEX IF NOT EXISTS idx_virtual_users_team
    ON virtual_users(team_id);

CREATE INDEX IF NOT EXISTS idx_virtual_users_org_unit
    ON virtual_users(org_unit_id);

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS key_type VARCHAR(20);

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS virtual_user_id UUID REFERENCES virtual_users(virtual_user_id);

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS org_unit_id UUID REFERENCES organizations(organization_id);

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS allowed_models TEXT[] NOT NULL DEFAULT '{}';

UPDATE verification_tokens
SET org_unit_id = organization_id
WHERE org_unit_id IS NULL
  AND organization_id IS NOT NULL;

UPDATE verification_tokens
SET allowed_models = models
WHERE cardinality(allowed_models) = 0
  AND cardinality(models) > 0;

UPDATE verification_tokens
SET key_type = CASE
    WHEN virtual_user_id IS NOT NULL THEN 'vat'
    WHEN team_id IS NOT NULL AND user_id IS NULL THEN 'vat'
    ELSE 'pat'
END
WHERE key_type IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_verification_tokens_key_type'
    ) THEN
        ALTER TABLE verification_tokens
            ADD CONSTRAINT chk_verification_tokens_key_type
            CHECK (key_type IN ('pat', 'vat'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_vt_virtual_user
    ON verification_tokens(virtual_user_id);

CREATE INDEX IF NOT EXISTS idx_vt_key_type
    ON verification_tokens(key_type);

CREATE INDEX IF NOT EXISTS idx_vt_org_unit
    ON verification_tokens(org_unit_id);

ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS org_unit_id UUID REFERENCES organizations(organization_id);

UPDATE spend_logs
SET org_unit_id = organization_id
WHERE org_unit_id IS NULL
  AND organization_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_spend_org_unit
    ON spend_logs(org_unit_id);

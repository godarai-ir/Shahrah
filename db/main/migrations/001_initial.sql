-- Migration: 001_initial.sql
-- Description: Initial database schema for Shahrah
-- (organizations -> teams -> users -> keys hierarchy)
--
-- Run with: docker-compose exec postgres psql -U postgres -d llmproxy -f /docker-entrypoint-initdb.d/001_initial.sql

-- Drop existing tables if re-running (for development only)
DROP TABLE IF EXISTS request_logs CASCADE;
DROP TABLE IF EXISTS spend_logs CASCADE;
DROP TABLE IF EXISTS credentials CASCADE;
DROP TABLE IF EXISTS model_deployments CASCADE;
DROP TABLE IF EXISTS policies CASCADE;
DROP TABLE IF EXISTS guardrails CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS end_users CASCADE;
DROP TABLE IF EXISTS verification_tokens CASCADE;
DROP TABLE IF EXISTS organization_memberships CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS model_aliases CASCADE;
DROP TABLE IF EXISTS object_permissions CASCADE;
DROP TABLE IF EXISTS budgets CASCADE;

-- Create tables in dependency order

-- Budget / Rate Limits (top-level, referenced by orgs, teams, keys)
CREATE TABLE budgets (
    budget_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    max_budget DECIMAL(10,4),
    soft_budget DECIMAL(10,4),
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    max_parallel_requests INT,
    budget_duration VARCHAR(50),
    budget_reset_at TIMESTAMPTZ,
    model_max_budget JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Object Permission Table (for RBAC)
CREATE TABLE object_permissions (
    object_permission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Model Aliases Table
CREATE TABLE model_aliases (
    id SERIAL PRIMARY KEY,
    aliases JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Organizations (top-level hierarchy)
CREATE TABLE organizations (
    organization_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_alias VARCHAR(255) NOT NULL,
    budget_id UUID REFERENCES budgets(budget_id),
    metadata JSONB DEFAULT '{}',
    models TEXT[] DEFAULT '{}',
    spend DECIMAL(10,4) DEFAULT 0.0,
    model_spend JSONB DEFAULT '{}',
    object_permission_id UUID REFERENCES object_permissions(object_permission_id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Teams (belong to organizations)
CREATE TABLE teams (
    team_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_alias VARCHAR(255),
    organization_id UUID REFERENCES organizations(organization_id),
    object_permission_id UUID REFERENCES object_permissions(object_permission_id),
    admins TEXT[] DEFAULT '{}',
    members TEXT[] DEFAULT '{}',
    members_with_roles JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    max_budget DECIMAL(10,4),
    soft_budget DECIMAL(10,4),
    spend DECIMAL(10,4) DEFAULT 0.0,
    models TEXT[] DEFAULT '{}',
    max_parallel_requests INT,
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    budget_duration VARCHAR(50),
    budget_reset_at TIMESTAMPTZ,
    blocked BOOLEAN DEFAULT FALSE,
    model_spend JSONB DEFAULT '{}',
    model_max_budget JSONB DEFAULT '{}',
    router_settings JSONB DEFAULT '{}',
    team_member_permissions TEXT[] DEFAULT '{}',
    access_group_ids TEXT[] DEFAULT '{}',
    policies TEXT[] DEFAULT '{}',
    default_team_member_models TEXT[] DEFAULT '{}',
    budget_limits JSONB,
    model_id INT REFERENCES model_aliases(id),
    allow_team_guardrail_config BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Projects (optional, between teams and keys)
CREATE TABLE projects (
    project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_alias VARCHAR(255),
    description VARCHAR(255),
    team_id UUID REFERENCES teams(team_id),
    budget_id UUID REFERENCES budgets(budget_id),
    metadata JSONB DEFAULT '{}',
    models TEXT[] DEFAULT '{}',
    spend DECIMAL(10,4) DEFAULT 0.0,
    model_spend JSONB DEFAULT '{}',
    model_rpm_limit JSONB DEFAULT '{}',
    model_tpm_limit JSONB DEFAULT '{}',
    blocked BOOLEAN DEFAULT FALSE,
    object_permission_id UUID REFERENCES object_permissions(object_permission_id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Users (Internal)
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_alias VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    role VARCHAR(50) DEFAULT 'internal_user',
    team_id UUID REFERENCES teams(team_id),
    organization_id UUID REFERENCES organizations(organization_id),
    max_budget DECIMAL(10,4),
    budget_duration VARCHAR(50),
    spend DECIMAL(10,4) DEFAULT 0.0,
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    metadata JSONB DEFAULT '{}',
    object_permission_id UUID REFERENCES object_permissions(object_permission_id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Organization Memberships (users within organizations)
CREATE TABLE organization_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- API Keys (Verification Tokens)
CREATE TABLE verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(64) UNIQUE NOT NULL,
    key_alias VARCHAR(255),
    key_name VARCHAR(255),
    team_id UUID REFERENCES teams(team_id),
    user_id UUID REFERENCES users(user_id),
    organization_id UUID REFERENCES organizations(organization_id),
    project_id UUID REFERENCES projects(project_id),
    models TEXT[] DEFAULT '{}',
    max_budget DECIMAL(10,4),
    soft_budget DECIMAL(10,4),
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    max_parallel_requests INT,
    budget_duration VARCHAR(50),
    budget_reset_at TIMESTAMPTZ,
    spend DECIMAL(10,4) DEFAULT 0.0,
    expires TIMESTAMPTZ,
    blocked BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- End Users (for tracking individual consumers)
CREATE TABLE end_users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_alias VARCHAR(255),
    team_id UUID REFERENCES teams(team_id),
    budget_id UUID REFERENCES budgets(budget_id),
    spend DECIMAL(10,4) DEFAULT 0.0,
    max_budget DECIMAL(10,4),
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    blocked BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tags (for grouping keys/requests)
CREATE TABLE tags (
    tag_name VARCHAR(255) PRIMARY KEY,
    budget_id UUID REFERENCES budgets(budget_id),
    spend DECIMAL(10,4) DEFAULT 0.0,
    max_budget DECIMAL(10,4),
    tpm_limit BIGINT,
    rpm_limit BIGINT,
    models TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Model Deployments (provider configurations)
CREATE TABLE model_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(255) NOT NULL,
    provider_config JSONB NOT NULL,
    model_info JSONB,
    rpm BIGINT,
    tpm BIGINT,
    timeout FLOAT,
    stream_timeout FLOAT,
    health_check_timeout INT DEFAULT 60,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Credentials (secure storage for API keys)
CREATE TABLE credentials (
    credential_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_name VARCHAR(255) UNIQUE NOT NULL,
    credential_values JSONB NOT NULL,
    credential_info JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by VARCHAR(255)
);

-- Spend Logs (batch writes for tracking)
CREATE TABLE spend_logs (
    id BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(64),
    team_id UUID,
    user_id UUID,
    organization_id UUID,
    end_user_id UUID,
    model VARCHAR(255),
    input_tokens INT,
    output_tokens INT,
    total_tokens INT,
    cost DECIMAL(10,6),
    response_time_ms INT,
    status VARCHAR(20),
    completion_start_time TIMESTAMPTZ,
    completion_end_time TIMESTAMPTZ,
    request_tags TEXT[],
    request_metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Request Logs (for audit/observability)
CREATE TABLE request_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(255),
    token_hash VARCHAR(64),
    team_id UUID,
    user_id UUID,
    model VARCHAR(255),
    request_body JSONB,
    response_body JSONB,
    status VARCHAR(20),
    cost DECIMAL(10,6),
    response_time_ms INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Guardrails Configuration
CREATE TABLE guardrails (
    guardrail_name VARCHAR(255) PRIMARY KEY,
    guardrail_type VARCHAR(50), -- 'input' | 'output'
    event_hook VARCHAR(50), -- 'pre_call' | 'post_call'
    guardrail_init_kwargs JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Policies
CREATE TABLE policies (
    policy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name VARCHAR(255) UNIQUE NOT NULL,
    policy_type VARCHAR(50), -- 'budget' | 'rate_limit' | 'model_access'
    policy_config JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Separate CREATE INDEX statements (PostgreSQL doesn't support inline INDEX in CREATE TABLE)
CREATE INDEX idx_team_org_id ON teams(organization_id);
CREATE INDEX idx_team_alias ON teams(team_alias);
CREATE INDEX idx_team_created_at ON teams(created_at);
CREATE INDEX idx_org_memberships_org ON organization_memberships(organization_id);
CREATE INDEX idx_org_memberships_user ON organization_memberships(user_id);
CREATE INDEX idx_vt_token_hash ON verification_tokens(token_hash);
CREATE INDEX idx_vt_team ON verification_tokens(team_id);
CREATE INDEX idx_vt_user ON verification_tokens(user_id);
CREATE INDEX idx_vt_org ON verification_tokens(organization_id);
CREATE INDEX idx_end_user_team ON end_users(team_id);
CREATE INDEX idx_model_name ON model_deployments(model_name);
CREATE INDEX idx_spend_token ON spend_logs(token_hash);
CREATE INDEX idx_spend_created ON spend_logs(created_at);
CREATE INDEX idx_spend_team ON spend_logs(team_id);
CREATE INDEX idx_spend_user ON spend_logs(user_id);
CREATE INDEX idx_spend_org ON spend_logs(organization_id);


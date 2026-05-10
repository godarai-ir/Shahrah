-- +goose Up
-- Remove legacy budget/rate guardrail fields so upcoming rule-engine owns enforcement.

ALTER TABLE IF EXISTS verification_tokens
	DROP COLUMN IF EXISTS max_budget,
	DROP COLUMN IF EXISTS rpm_limit,
	DROP COLUMN IF EXISTS tpm_limit;

ALTER TABLE IF EXISTS organizations
	DROP CONSTRAINT IF EXISTS organizations_budget_id_fkey,
	DROP COLUMN IF EXISTS budget_id;

ALTER TABLE IF EXISTS team_memberships
	DROP CONSTRAINT IF EXISTS team_memberships_budget_id_fkey,
	DROP COLUMN IF EXISTS budget_id;

ALTER TABLE IF EXISTS teams
	DROP COLUMN IF EXISTS max_budget,
	DROP COLUMN IF EXISTS soft_budget,
	DROP COLUMN IF EXISTS tpm_limit,
	DROP COLUMN IF EXISTS rpm_limit,
	DROP COLUMN IF EXISTS budget_duration;


-- +goose Down
-- Re-add legacy budget/rate guardrail fields.

ALTER TABLE IF EXISTS verification_tokens
	ADD COLUMN IF NOT EXISTS max_budget NUMERIC,
	ADD COLUMN IF NOT EXISTS rpm_limit INTEGER,
	ADD COLUMN IF NOT EXISTS tpm_limit INTEGER;

ALTER TABLE IF EXISTS organizations
	ADD COLUMN IF NOT EXISTS budget_id UUID;

ALTER TABLE IF EXISTS team_memberships
	ADD COLUMN IF NOT EXISTS budget_id UUID;

ALTER TABLE IF EXISTS teams
	ADD COLUMN IF NOT EXISTS max_budget NUMERIC,
	ADD COLUMN IF NOT EXISTS soft_budget NUMERIC,
	ADD COLUMN IF NOT EXISTS tpm_limit INTEGER,
	ADD COLUMN IF NOT EXISTS rpm_limit INTEGER,
	ADD COLUMN IF NOT EXISTS budget_duration TEXT;
-- 010_personal_access_tokens

-- +goose Up
-- +goose StatementBegin

-- Migration: 010_personal_access_tokens
-- Description: Extend verification_tokens for PAT lifecycle and ownership.

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS key_type VARCHAR(16);

UPDATE verification_tokens
SET key_type = CASE
    WHEN NULLIF(team_id::text, '') IS NOT NULL THEN 'vat'
    ELSE 'pat'
END
WHERE key_type IS NULL;

ALTER TABLE verification_tokens
    ALTER COLUMN key_type SET DEFAULT 'pat';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'verification_tokens_key_type_check'
    ) THEN
        ALTER TABLE verification_tokens
            ADD CONSTRAINT verification_tokens_key_type_check
            CHECK (key_type IN ('pat', 'vat'));
    END IF;
END $$;

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS created_by_user_id TEXT;

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS revoked_by_user_id TEXT;

CREATE INDEX IF NOT EXISTS idx_verification_tokens_pat_owner_state
ON verification_tokens (tenant_id, key_type, created_by_user_id, revoked_at);

CREATE INDEX IF NOT EXISTS idx_verification_tokens_tenant_keytype
ON verification_tokens (tenant_id, key_type);

-- +goose StatementEnd

-- +goose Down

-- Migration: 010_personal_access_tokens (DOWN)

DROP INDEX IF EXISTS idx_verification_tokens_tenant_keytype;
DROP INDEX IF EXISTS idx_verification_tokens_pat_owner_state;

ALTER TABLE verification_tokens DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE verification_tokens DROP COLUMN IF EXISTS last_used_at;
ALTER TABLE verification_tokens DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE verification_tokens DROP COLUMN IF EXISTS revoked_by_user_id;
ALTER TABLE verification_tokens DROP CONSTRAINT IF EXISTS verification_tokens_key_type_check;
ALTER TABLE verification_tokens DROP COLUMN IF EXISTS key_type;

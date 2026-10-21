-- +goose Up
-- The IRR signup credit is gone; new tenants start with an empty wallet.
DROP TABLE IF EXISTS public.commercial_signup_settings;

-- +goose Down
-- The dropped table is not restored.
SELECT 1;

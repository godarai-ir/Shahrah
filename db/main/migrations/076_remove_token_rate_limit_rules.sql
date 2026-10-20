-- +goose Up
-- Token-based (TPM) rate-limit rules are no longer supported. Delete the stored
-- rules that use a token dimension or the legacy tpm keys. Keeping them would
-- make the runtime fail to parse each rule and reject every matching request.

-- +goose StatementBegin
DELETE FROM public.rate_limit_rules
WHERE lower(COALESCE(limit_json->>'dimension', '')) IN ('tokens', 'token', 'tpm')
   OR limit_json ? 'tpm'
   OR limit_json ? 'tokens_per_minute';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'migration 076_remove_token_rate_limit_rules deletes data and cannot be rolled back';
END
$$;
-- +goose StatementEnd

-- +goose Up
-- Existing support-admin credentials predate the provider-credential API scopes.
-- Keep the token stable while granting only the permissions already exposed by
-- support-admin's independently authenticated and RBAC-protected workflows.
UPDATE public.platform_service_account_credentials AS credential
SET scopes = (
    SELECT ARRAY_AGG(DISTINCT scope ORDER BY scope)
    FROM UNNEST(
        credential.scopes || ARRAY[
            'platform:read',
            'platform:catalog:write',
            'platform:tenant:write',
            'platform:provider-credentials:read-metadata',
            'platform:provider-credentials:create-version',
            'platform:provider-credentials:test',
            'platform:provider-credentials:activate',
            'platform:provider-credentials:disable'
        ]::text[]
    ) AS scope
)
WHERE LOWER(BTRIM(credential.account_name)) IN ('support-admin', 'platform-support')
  AND credential.active = true
  AND credential.revoked_at IS NULL
  AND (credential.expires_at IS NULL OR credential.expires_at > NOW());

-- +goose Down
-- This is an intentional no-op. Removing scopes on rollback could revoke a scope
-- that was explicitly assigned before this migration and cause an outage.
SELECT 1;

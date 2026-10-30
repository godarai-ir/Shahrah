-- +goose Up
-- The chat provisioner now registers the admin panel callback
-- (/api/admin/oauth/openid/callback) on each tenant's GodarChat OIDC client, but it only
-- claims tenants that are pending or whose slug or status changed. Re-queue the tenants
-- that are already ready so their clients gain the callback on the next pass. The pass
-- only updates the client's URIs: credentials_delivered is kept, so no secret is rotated
-- or resent.
UPDATE public.tenant_chat_provisioning
SET state = 'pending',
    next_attempt_at = now(),
    updated_at = now()
WHERE state = 'ready';

-- +goose Down
-- Nothing to undo: the provisioner returns every re-queued tenant to ready.
SELECT 1;

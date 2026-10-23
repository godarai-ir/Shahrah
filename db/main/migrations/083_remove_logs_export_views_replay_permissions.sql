-- +goose Up
-- Removes the dedicated log export, saved view and replay permissions. Export and
-- saved views follow logs:read_metadata, and replay no longer exists.
DELETE FROM public.role_permissions
WHERE LOWER(permission) IN (
    'logs:export',
    'logs:manage_views',
    'logs:replay'
);

-- +goose Down
-- The removed permissions are not restored.
SELECT 1;

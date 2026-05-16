-- +goose Up
-- +goose StatementBegin
-- Ensure new system roles exist for every tenant.
INSERT INTO roles (role_id, tenant_id, name, display_name, description, is_system, created_at, updated_at)
SELECT gen_random_uuid(), t.tenant_id, 'member', 'Member', 'System role', true, NOW(), NOW()
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1
    FROM roles r
    WHERE r.tenant_id = t.tenant_id
      AND LOWER(r.name) = 'member'
);

INSERT INTO roles (role_id, tenant_id, name, display_name, description, is_system, created_at, updated_at)
SELECT gen_random_uuid(), t.tenant_id, 'readonly_member', 'Readonly Member', 'System role', true, NOW(), NOW()
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1
    FROM roles r
    WHERE r.tenant_id = t.tenant_id
      AND LOWER(r.name) = 'readonly_member'
);

-- Keep system-role metadata deterministic.
UPDATE roles
SET is_system = true,
    display_name = 'Member',
    updated_at = NOW()
WHERE LOWER(name) = 'member';

UPDATE roles
SET is_system = true,
    display_name = 'Readonly Member',
    updated_at = NOW()
WHERE LOWER(name) = 'readonly_member';

-- Replace permission sets for the two member roles.
DELETE FROM role_permissions rp
USING roles r
WHERE rp.role_id = r.role_id
  AND LOWER(r.name) IN ('member', 'readonly_member');

INSERT INTO role_permissions (role_permission_id, tenant_id, role_id, permission, created_at, updated_at)
SELECT gen_random_uuid(), r.tenant_id, r.role_id, p.permission, NOW(), NOW()
FROM roles r
JOIN (
    SELECT 'member'::text AS role_name, 'user:list'::text AS permission
    UNION ALL
    SELECT 'member', 'settings:list'
    UNION ALL
    SELECT 'readonly_member', 'settings:list'
) p ON LOWER(r.name) = p.role_name;

-- Remove duplicate bindings where target member binding already exists.
DELETE FROM user_role_bindings old_b
USING roles old_r, roles member_r, user_role_bindings new_b
WHERE old_b.role_id = old_r.role_id
  AND old_b.tenant_id = old_r.tenant_id
  AND old_b.tenant_id = member_r.tenant_id
  AND LOWER(old_r.name) IN ('team_member', 'tenant_user', 'tenant_member')
  AND LOWER(member_r.name) = 'member'
  AND new_b.tenant_id = old_b.tenant_id
  AND new_b.user_id = old_b.user_id
  AND new_b.role_id = member_r.role_id
  AND COALESCE(new_b.resource_type, '') = COALESCE(old_b.resource_type, '')
  AND COALESCE(new_b.resource_id::text, '') = COALESCE(old_b.resource_id::text, '')
  AND new_b.is_active = old_b.is_active;

-- Move legacy user-role bindings to member.
UPDATE user_role_bindings old_b
SET role_id = member_r.role_id,
    updated_at = NOW()
FROM roles old_r, roles member_r
WHERE old_b.role_id = old_r.role_id
  AND old_b.tenant_id = old_r.tenant_id
  AND old_b.tenant_id = member_r.tenant_id
  AND LOWER(old_r.name) IN ('team_member', 'tenant_user', 'tenant_member')
  AND LOWER(member_r.name) = 'member';

-- Normalize legacy role values on users table.
UPDATE users
SET role = 'member',
    updated_at = NOW()
WHERE LOWER(COALESCE(role, '')) IN ('team_member', 'tenant_user', 'tenant_member');

-- Remove deprecated tenant-level system roles.
DELETE FROM roles
WHERE LOWER(name) IN ('team_member', 'tenant_user', 'tenant_member');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- This migration is intentionally irreversible.
SELECT 1;
-- +goose StatementEnd

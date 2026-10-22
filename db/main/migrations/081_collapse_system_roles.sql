-- +goose Up
-- Collapses the tenant system roles to tenant_admin and member. The
-- billing_manager, team_admin, and readonly_member roles overlapped with those
-- two; their holders become member. Narrower access is granted through custom
-- roles. team_memberships.role is a per-team setting and is left untouched.

-- +goose StatementBegin
DO $$
BEGIN
    UPDATE public.users
    SET role = 'member', updated_at = NOW()
    WHERE LOWER(COALESCE(role, '')) IN ('billing_manager', 'team_admin', 'readonly_member');

    UPDATE public.tenant_admin_invitations
    SET role = 'member'
    WHERE LOWER(role) IN ('billing_manager', 'team_admin', 'readonly_member');

    -- Make sure every tenant that is losing a role has a member role to land on.
    INSERT INTO public.roles (role_id, tenant_id, name, display_name, description, is_system, created_at, updated_at)
    SELECT gen_random_uuid(), t.tenant_id, 'member', 'Member', 'System role', true, NOW(), NOW()
    FROM (
        SELECT DISTINCT tenant_id FROM public.roles
        WHERE LOWER(name) IN ('billing_manager', 'team_admin', 'readonly_member')
    ) t
    WHERE NOT EXISTS (
        SELECT 1 FROM public.roles r
        WHERE r.tenant_id = t.tenant_id AND LOWER(r.name) = 'member'
    );

    -- Point bindings at the tenant's member role, skipping users that already hold it.
    UPDATE public.user_role_bindings b
    SET role_id = target.role_id, updated_at = NOW()
    FROM public.roles old_role
    JOIN public.roles target
      ON target.tenant_id = old_role.tenant_id AND LOWER(target.name) = 'member'
    WHERE b.role_id = old_role.role_id
      AND LOWER(old_role.name) IN ('billing_manager', 'team_admin', 'readonly_member')
      AND NOT EXISTS (
          SELECT 1 FROM public.user_role_bindings x
          WHERE x.tenant_id = b.tenant_id AND x.user_id = b.user_id
            AND x.role_id = target.role_id
            AND x.resource_type = b.resource_type
            AND COALESCE(x.resource_id::text, '') = COALESCE(b.resource_id::text, '')
      );

    DELETE FROM public.user_role_bindings b
    USING public.roles old_role
    WHERE b.role_id = old_role.role_id
      AND LOWER(old_role.name) IN ('billing_manager', 'team_admin', 'readonly_member');

    -- role_permissions rows go with the role (ON DELETE CASCADE).
    DELETE FROM public.roles
    WHERE LOWER(name) IN ('billing_manager', 'team_admin', 'readonly_member', 'org_admin');
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'migration 081_collapse_system_roles cannot be rolled back';
END
$$;
-- +goose StatementEnd

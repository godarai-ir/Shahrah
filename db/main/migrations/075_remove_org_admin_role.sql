-- +goose Up
-- Removes the org_admin system role. It managed teams inside an organizational
-- unit, a concept removed in 074. Existing org_admin holders become team_admin.

-- +goose StatementBegin
DO $$
DECLARE
    col record;
    moved bigint;
BEGIN
    -- Any plain role column (users, invitations, memberships) that holds the old role.
    FOR col IN
        SELECT table_name, column_name
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND column_name IN ('role', 'role_name', 'invited_role')
          AND data_type IN ('character varying', 'text')
    LOOP
        EXECUTE format('UPDATE public.%I SET %I = %L WHERE %I = %L',
            col.table_name, col.column_name, 'team_admin', col.column_name, 'org_admin');
        GET DIAGNOSTICS moved = ROW_COUNT;
        IF moved > 0 THEN
            RAISE NOTICE 'remapped % org_admin values in %.%', moved, col.table_name, col.column_name;
        END IF;
    END LOOP;

    -- Role bindings: point them at the same tenant's team_admin role, skipping
    -- users that already hold it, then drop what is left.
    UPDATE public.user_role_bindings b
    SET role_id = target.role_id, updated_at = NOW()
    FROM public.roles old_role
    JOIN public.roles target
      ON target.tenant_id = old_role.tenant_id AND target.name = 'team_admin'
    WHERE b.role_id = old_role.role_id
      AND old_role.name = 'org_admin'
      AND NOT EXISTS (
          SELECT 1 FROM public.user_role_bindings x
          WHERE x.tenant_id = b.tenant_id AND x.user_id = b.user_id
            AND x.role_id = target.role_id
            AND x.resource_type = b.resource_type
            AND COALESCE(x.resource_id::text, '') = COALESCE(b.resource_id::text, '')
      );

    DELETE FROM public.user_role_bindings b
    USING public.roles old_role
    WHERE b.role_id = old_role.role_id AND old_role.name = 'org_admin';

    -- role_permissions rows go with the role (ON DELETE CASCADE).
    DELETE FROM public.roles WHERE name = 'org_admin';
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'migration 075_remove_org_admin_role cannot be rolled back';
END
$$;
-- +goose StatementEnd

DELETE FROM public.role_permissions
WHERE LOWER(permission) IN (
    'support_ticket:create',
    'support_ticket:read_own',
    'support_ticket:read_all',
    'support_ticket:reply',
    'support_ticket:rate'
);

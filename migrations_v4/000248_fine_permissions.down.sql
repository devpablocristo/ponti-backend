BEGIN;

DELETE FROM public.auth_role_permissions
WHERE permission_id IN (SELECT id FROM public.auth_permissions WHERE name IN ('users:manage', 'invites:write'));

DELETE FROM public.auth_permissions WHERE name IN ('users:manage', 'invites:write');

COMMIT;

BEGIN;

DELETE FROM public.auth_role_permissions
WHERE role_id IN (SELECT id FROM public.auth_roles WHERE name = 'tenant_owner');

DELETE FROM public.auth_roles WHERE name = 'tenant_owner';

ALTER TABLE public.users DROP COLUMN IF EXISTS is_platform_admin;

COMMIT;

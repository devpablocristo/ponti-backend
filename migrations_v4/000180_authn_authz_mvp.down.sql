BEGIN;

DROP TABLE IF EXISTS public.auth_memberships;
DROP TABLE IF EXISTS public.auth_role_permissions;
DROP TABLE IF EXISTS public.auth_permissions;
DROP TABLE IF EXISTS public.auth_roles;
DROP TABLE IF EXISTS public.auth_tenants;

DROP INDEX IF EXISTS public.uq_users_idp_sub;

ALTER TABLE public.users
    DROP COLUMN IF EXISTS idp_sub,
    DROP COLUMN IF EXISTS idp_email;

COMMIT;

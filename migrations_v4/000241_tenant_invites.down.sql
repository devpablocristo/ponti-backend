BEGIN;

DROP TRIGGER IF EXISTS trg_min_one_tenant_owner ON public.auth_memberships;
DROP FUNCTION IF EXISTS public.enforce_min_one_tenant_owner();
DROP TABLE IF EXISTS public.tenant_invites;

COMMIT;

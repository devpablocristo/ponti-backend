-- ========================================
-- MIGRATION 000193 AUTH FK INDEXES (DOWN)
-- ========================================

BEGIN;

DROP INDEX IF EXISTS public.idx_auth_memberships_tenant_id;
DROP INDEX IF EXISTS public.idx_auth_memberships_role_id;
DROP INDEX IF EXISTS public.idx_auth_role_permissions_permission_id;

COMMIT;


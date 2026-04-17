-- ========================================
-- MIGRATION 000193 AUTH FK INDEXES (UP)
-- ========================================
-- db_validate exige que toda FK tenga un índice cuyo prefijo coincida con las columnas.
-- Las tablas de auth tienen PK/UNIQUE que cubren algunos prefijos, pero faltan:
-- - auth_memberships.tenant_id
-- - auth_memberships.role_id
-- - auth_role_permissions.permission_id

BEGIN;

CREATE INDEX IF NOT EXISTS idx_auth_memberships_tenant_id
  ON public.auth_memberships (tenant_id);

CREATE INDEX IF NOT EXISTS idx_auth_memberships_role_id
  ON public.auth_memberships (role_id);

CREATE INDEX IF NOT EXISTS idx_auth_role_permissions_permission_id
  ON public.auth_role_permissions (permission_id);

COMMIT;


BEGIN;

-- U0/U2 (Pilar 2): catálogo de permisos FINOS para la superficie admin per-tenant.
-- Aditivo + idempotente. Mapeo elegido para NO cambiar accesos respecto de hoy:
--   users:manage   -> admin              (hoy requireAdmin = role 'admin')
--   invites:write  -> admin, tenant_owner (hoy canManageTenant = admin|tenant_owner)
-- El dual-check en código evalúa el permiso fino y, si el rol todavía no lo tiene,
-- cae al ROL actual (no a api.write, que es más amplio) y loguea fallback_to_coarse.
-- Cuando todos los roles tengan el fino (fallback_to_coarse=0) se puede retirar el rol (U5).

INSERT INTO public.auth_permissions (name)
SELECT v.n FROM (VALUES ('users:manage'), ('invites:write')) AS v(n)
WHERE NOT EXISTS (SELECT 1 FROM public.auth_permissions p WHERE p.name = v.n);

-- users:manage -> admin
INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name = 'users:manage'
WHERE r.name = 'admin'
  AND NOT EXISTS (SELECT 1 FROM public.auth_role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id);

-- invites:write -> admin, tenant_owner
INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name = 'invites:write'
WHERE r.name IN ('admin', 'tenant_owner')
  AND NOT EXISTS (SELECT 1 FROM public.auth_role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id);

COMMIT;

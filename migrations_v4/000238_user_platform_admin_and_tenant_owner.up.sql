BEGIN;

-- U1 (Pilar 2): is_platform_admin como columna persistente en users. Reemplaza al
-- allowlist env (AUTH_PLATFORM_ADMIN_SUBJECTS) como fuente; el código mantiene el env
-- como FALLBACK de transición. Default false -> nadie gana platform-admin por DB hasta
-- setearlo explícitamente (sin cambio de comportamiento respecto del allowlist actual).
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS is_platform_admin boolean NOT NULL DEFAULT false;

-- U1: rol SaaS 'tenant_owner' (rol tope POR-TENANT; lo requiere el invariante >=1 owner
-- de U4). Aditivo: no toca admin/manager/viewer. id/legacy_id/created_at usan defaults.
INSERT INTO public.auth_roles (name)
SELECT 'tenant_owner'
WHERE NOT EXISTS (SELECT 1 FROM public.auth_roles WHERE name = 'tenant_owner');

-- Mapea tenant_owner a los permisos gruesos actuales (api.read/api.write): al menos lo
-- que tiene un admin. Idempotente.
INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name IN ('api.read', 'api.write')
WHERE r.name = 'tenant_owner'
  AND NOT EXISTS (
    SELECT 1 FROM public.auth_role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

COMMIT;

-- Reconciliar clientes archivados según proyectos activos
-- Regla: customer archivado si NO tiene proyectos activos

BEGIN;

-- Archivar customers sin proyectos activos
UPDATE public.customers c
SET deleted_at = NOW(),
    updated_at = NOW(),
    deleted_by = NULL
WHERE c.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM public.projects p
    WHERE p.customer_id = c.id
      AND p.deleted_at IS NULL
  );

-- Restaurar customers con proyectos activos
UPDATE public.customers c
SET deleted_at = NULL,
    updated_at = NOW(),
    deleted_by = NULL
WHERE c.deleted_at IS NOT NULL
  AND EXISTS (
    SELECT 1
    FROM public.projects p
    WHERE p.customer_id = c.id
      AND p.deleted_at IS NULL
  );

COMMIT;

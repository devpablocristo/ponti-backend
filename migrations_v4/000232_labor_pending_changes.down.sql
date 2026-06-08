BEGIN;

ALTER TABLE public.labors DROP COLUMN IF EXISTS is_pending;
ALTER TABLE public.labors ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE public.labors ALTER COLUMN price DROP DEFAULT;

-- Restaurar vista sin fallback de contratista
DROP VIEW IF EXISTS v4_report.workorder_list;
-- (la vista anterior se restaura con la migración 000221)

COMMIT;

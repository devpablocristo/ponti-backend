-- Quita columnas de auditoría en business_parameters
BEGIN;
ALTER TABLE public.business_parameters
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by;
COMMIT;

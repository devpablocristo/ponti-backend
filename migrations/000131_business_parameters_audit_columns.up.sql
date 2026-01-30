-- Agrega columnas de auditoría faltantes en business_parameters
BEGIN;
ALTER TABLE public.business_parameters
    ADD COLUMN IF NOT EXISTS created_by bigint,
    ADD COLUMN IF NOT EXISTS updated_by bigint,
    ADD COLUMN IF NOT EXISTS deleted_by bigint,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
COMMIT;

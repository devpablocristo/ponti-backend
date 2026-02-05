-- ========================================
-- MIGRATION 000030 FIELDS LOTS TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.lot_dates;
DROP TABLE IF EXISTS public.lots;
DROP TABLE IF EXISTS public.fields;
DROP TABLE IF EXISTS public.lease_types;

DROP SEQUENCE IF EXISTS public.lease_types_id_seq;
DROP SEQUENCE IF EXISTS public.fields_id_seq;
DROP SEQUENCE IF EXISTS public.lot_dates_id_seq;
DROP SEQUENCE IF EXISTS public.lots_id_seq;

COMMIT;

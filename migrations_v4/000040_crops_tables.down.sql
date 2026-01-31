-- ========================================
-- MIGRATION 000040 CROPS TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.crops;
DROP SEQUENCE IF EXISTS public.crops_id_seq;

COMMIT;

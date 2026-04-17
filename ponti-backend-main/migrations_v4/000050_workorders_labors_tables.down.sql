-- ========================================
-- MIGRATION 000050 WORKORDERS LABORS TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.workorder_items;
DROP TABLE IF EXISTS public.workorders;
DROP TABLE IF EXISTS public.labors;
DROP TABLE IF EXISTS public.labor_categories;
DROP TABLE IF EXISTS public.labor_types;

DROP SEQUENCE IF EXISTS public.labor_types_id_seq;
DROP SEQUENCE IF EXISTS public.labor_categories_id_seq;
DROP SEQUENCE IF EXISTS public.labors_id_seq;
DROP SEQUENCE IF EXISTS public.workorders_id_seq;
DROP SEQUENCE IF EXISTS public.workorder_items_id_seq;

COMMIT;

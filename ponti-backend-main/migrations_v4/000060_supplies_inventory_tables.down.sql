-- ========================================
-- MIGRATION 000060 SUPPLIES INVENTORY TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.providers;
DROP TABLE IF EXISTS public.supply_movements;
DROP TABLE IF EXISTS public.stocks;
DROP TABLE IF EXISTS public.supplies;
DROP TABLE IF EXISTS public.categories;
DROP TABLE IF EXISTS public.types;

DROP SEQUENCE IF EXISTS public.types_id_seq;
DROP SEQUENCE IF EXISTS public.categories_id_seq;
DROP SEQUENCE IF EXISTS public.supplies_id_seq;
DROP SEQUENCE IF EXISTS public.stocks_id_seq;
DROP SEQUENCE IF EXISTS public.supply_movements_id_seq;
DROP SEQUENCE IF EXISTS public.providers_id_seq;

COMMIT;

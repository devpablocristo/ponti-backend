-- ========================================
-- MIGRATION CORE BASE TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.fx_rates ;
DROP TABLE IF EXISTS public.app_parameters ;
DROP TABLE IF EXISTS public.users ;
DROP SEQUENCE IF EXISTS public.users_id_seq ;
DROP SEQUENCE IF EXISTS public.app_parameters_id_seq ;
DROP SEQUENCE IF EXISTS public.fx_rates_id_seq ;

COMMIT;

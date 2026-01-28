-- ========================================
-- MIGRATION INVESTORS_COMMERCIALIZATION TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.invoices ;
DROP TABLE IF EXISTS public.project_dollar_values ;
DROP TABLE IF EXISTS public.field_investors ;
DROP TABLE IF EXISTS public.admin_cost_investors ;
DROP TABLE IF EXISTS public.crop_commercializations ;
DROP TABLE IF EXISTS public.project_investors ;
DROP TABLE IF EXISTS public.investors ;
DROP SEQUENCE IF EXISTS public.investors_id_seq ;
DROP SEQUENCE IF EXISTS public.project_investors_id_seq ;
DROP SEQUENCE IF EXISTS public.crop_commercializations_id_seq ;
DROP SEQUENCE IF EXISTS public.admin_cost_investors_id_seq ;
DROP SEQUENCE IF EXISTS public.field_investors_id_seq ;
DROP SEQUENCE IF EXISTS public.project_dollar_values_id_seq ;
DROP SEQUENCE IF EXISTS public.invoices_id_seq ;

COMMIT;

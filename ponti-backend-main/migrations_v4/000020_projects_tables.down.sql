-- ========================================
-- MIGRATION 000020 PROJECTS TABLES (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TABLE IF EXISTS public.project_managers;
DROP TABLE IF EXISTS public.managers;
DROP TABLE IF EXISTS public.projects;
DROP TABLE IF EXISTS public.campaigns;
DROP TABLE IF EXISTS public.customers;

DROP SEQUENCE IF EXISTS public.customers_id_seq;
DROP SEQUENCE IF EXISTS public.campaigns_id_seq;
DROP SEQUENCE IF EXISTS public.projects_id_seq;
DROP SEQUENCE IF EXISTS public.managers_id_seq;

COMMIT;

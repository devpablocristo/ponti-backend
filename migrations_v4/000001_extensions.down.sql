-- ========================================
-- MIGRATION 000001 EXTENSIONS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS unaccent;

COMMIT;

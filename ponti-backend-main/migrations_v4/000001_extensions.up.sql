-- ========================================
-- MIGRATION 000001 EXTENSIONS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Extensiones requeridas para búsquedas y normalización
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

COMMIT;

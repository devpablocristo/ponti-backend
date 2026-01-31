-- ========================================
-- MIGRATION 000090 V4 SCHEMAS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP SCHEMA IF EXISTS v4_report;
DROP SCHEMA IF EXISTS v4_calc;
DROP SCHEMA IF EXISTS v4_ssot;
DROP SCHEMA IF EXISTS v4_core;

COMMIT;

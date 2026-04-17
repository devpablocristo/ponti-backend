-- ========================================
-- MIGRATION 000090 V4 SCHEMAS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

CREATE SCHEMA IF NOT EXISTS v4_core;
CREATE SCHEMA IF NOT EXISTS v4_ssot;
CREATE SCHEMA IF NOT EXISTS v4_calc;
CREATE SCHEMA IF NOT EXISTS v4_report;

COMMIT;

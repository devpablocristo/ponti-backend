-- ==========================================================
-- MIGRATION 000126 V4 WORKORDER METRICS RAW (DOWN)
-- ==========================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS v4_calc.workorder_metrics_raw CASCADE;

COMMIT;

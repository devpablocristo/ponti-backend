-- ==========================================
-- MIGRATION 000122 V4 WORKORDER LIST (DOWN)
-- ==========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS v4_report.workorder_list CASCADE;

COMMIT;

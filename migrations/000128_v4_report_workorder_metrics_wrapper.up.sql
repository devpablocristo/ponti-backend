-- ==========================================================
-- MIGRATION 000128 V4 REPORT WORKORDER METRICS WRAPPER (UP)
-- ==========================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Wrapper a la vista de métricas en v4_calc
CREATE OR REPLACE VIEW v4_report.workorder_metrics AS
SELECT * FROM v4_calc.workorder_metrics;

COMMIT;

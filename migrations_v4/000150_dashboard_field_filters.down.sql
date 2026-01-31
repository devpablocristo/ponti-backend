-- ========================================
-- MIGRATION 000150 DASHBOARD FIELD FILTERS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS v4_report.dashboard_operational_indicators_field;
DROP VIEW IF EXISTS v4_report.dashboard_crop_incidence_field;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance_field;
DROP VIEW IF EXISTS v4_report.dashboard_metrics_field;

COMMIT;

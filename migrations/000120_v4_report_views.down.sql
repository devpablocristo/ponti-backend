-- ========================================
-- MIGRATION 000120 V4 REPORT VIEWS.DOWN (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS v4_report.dashboard_contributions_progress;
DROP VIEW IF EXISTS v4_report.investor_contribution_data;
DROP VIEW IF EXISTS v4_report.investor_distributions;
DROP VIEW IF EXISTS v4_report.investor_contribution_categories;
DROP VIEW IF EXISTS v4_report.investor_project_base;
DROP VIEW IF EXISTS v4_report.dashboard_operational_indicators;
DROP VIEW IF EXISTS v4_report.dashboard_crop_incidence;
DROP VIEW IF EXISTS v4_report.dashboard_metrics;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance;
DROP VIEW IF EXISTS v4_report.summary_results;
DROP VIEW IF EXISTS v4_report.field_crop_rentabilidad;
DROP VIEW IF EXISTS v4_report.field_crop_metrics;
DROP VIEW IF EXISTS v4_report.field_crop_labores;
DROP VIEW IF EXISTS v4_report.field_crop_insumos;
DROP VIEW IF EXISTS v4_report.field_crop_economicos;
DROP VIEW IF EXISTS v4_report.field_crop_cultivos;
DROP VIEW IF EXISTS v4_report.labor_metrics;
DROP VIEW IF EXISTS v4_report.labor_list;
DROP VIEW IF EXISTS v4_report.lot_list;
DROP VIEW IF EXISTS v4_report.lot_metrics;

COMMIT;

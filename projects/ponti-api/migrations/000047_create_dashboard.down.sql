-- =========================================================
-- ROLLBACK MIGRACIÓN 000047: Dashboard básico
-- Elimina todas las vistas e índices creados
-- =========================================================

-- Eliminar índices creados
DROP INDEX IF EXISTS idx_lots_field_id;
DROP INDEX IF EXISTS idx_lots_current_crop_id;
DROP INDEX IF EXISTS idx_labors_project_id;
DROP INDEX IF EXISTS idx_supplies_project_id;
DROP INDEX IF EXISTS idx_workorders_lot_id;
DROP INDEX IF EXISTS idx_fields_project_id;
DROP INDEX IF EXISTS idx_projects_campaign_cust;

-- Eliminar vistas del dashboard
DROP VIEW IF EXISTS dashboard_campaign_metrics_view;
DROP VIEW IF EXISTS dashboard_field_metrics_view;
DROP VIEW IF EXISTS dashboard_financial_metrics_view;
DROP VIEW IF EXISTS dashboard_supply_metrics_view;
DROP VIEW IF EXISTS dashboard_labor_metrics_view;
DROP VIEW IF EXISTS dashboard_crop_metrics_view;
DROP VIEW IF EXISTS dashboard_project_metrics_view;
DROP VIEW IF EXISTS dashboard_total_area_view;
DROP VIEW IF EXISTS dashboard_income_by_field_view;
DROP VIEW IF EXISTS dashboard_card_costs_progress_view;
DROP VIEW IF EXISTS dashboard_card_harvest_view;
DROP VIEW IF EXISTS dashboard_card_sowing_view;

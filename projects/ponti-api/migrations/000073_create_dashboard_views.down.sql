-- ========================================
-- ROLLBACK: MIGRACIÓN 000073 - VISTAS PARA DASHBOARD
-- ========================================

-- Eliminar todas las vistas del dashboard
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2;
DROP VIEW IF EXISTS dashboard_management_balance_view_v2;
DROP VIEW IF EXISTS dashboard_operating_result_view_v2;
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2;
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2;
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2;

-- ========================================
-- MIGRATION 000076: DROP ALL LEGACY VIEWS AND SCHEMAS (UP)
-- ========================================
-- 
-- Purpose: Clean slate - remove all legacy views, schemas and functions
-- Date: 2025-09-13
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- Eliminar schemas completos (CASCADE elimina todo su contenido)
DROP SCHEMA IF EXISTS calc CASCADE;
DROP SCHEMA IF EXISTS calc_common CASCADE;
DROP SCHEMA IF EXISTS report CASCADE;

-- Eliminar todas las vistas legacy del dashboard
DROP VIEW IF EXISTS dashboard_contributions_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_costs_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_harvest_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_sowing_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_operating_result_view CASCADE;
DROP VIEW IF EXISTS dashboard_operating_result_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_operational_indicators_view CASCADE;
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_crop_cost_incidence_view CASCADE;
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_balance_management_view CASCADE;
DROP VIEW IF EXISTS dashboard_view CASCADE;

-- Eliminar vistas base (serán reemplazadas por funciones v3_calc)
DROP VIEW IF EXISTS base_admin_costs_view CASCADE;
DROP VIEW IF EXISTS base_direct_costs_view CASCADE;
DROP VIEW IF EXISTS base_income_net_view CASCADE;
DROP VIEW IF EXISTS base_lease_calculations_view CASCADE;
DROP VIEW IF EXISTS base_active_total_view CASCADE;
DROP VIEW IF EXISTS base_operating_result_view CASCADE;
DROP VIEW IF EXISTS base_yield_calculations_view CASCADE;

-- Eliminar vistas de workorders y labor
DROP VIEW IF EXISTS workorder_metrics_view CASCADE;
DROP VIEW IF EXISTS workorder_metrics_view_v2 CASCADE;
DROP VIEW IF EXISTS workorder_list_view CASCADE;
DROP VIEW IF EXISTS labor_cards_cube_view CASCADE;
DROP VIEW IF EXISTS labor_cards_cube_view_v2 CASCADE;
DROP VIEW IF EXISTS labor_metrics_view CASCADE;

-- Eliminar vistas de lotes
DROP VIEW IF EXISTS lot_metrics_view CASCADE;
DROP VIEW IF EXISTS lot_table_view CASCADE;
DROP VIEW IF EXISTS fix_labors_list CASCADE;
DROP VIEW IF EXISTS fix_lot_list CASCADE;
DROP VIEW IF EXISTS fix_lots_metrics CASCADE;

-- Eliminar vistas de reportes
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2 CASCADE;
DROP VIEW IF EXISTS investor_contribution_data_view CASCADE;

-- Eliminar vistas auxiliares
DROP VIEW IF EXISTS views_fixes CASCADE;

-- Eliminar funciones legacy que puedan estar sueltas
DROP FUNCTION IF EXISTS calc.norm_dose(numeric, numeric);

COMMIT;

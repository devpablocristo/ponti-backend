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

-- Eliminar tipos ENUM legacy (serán recreados en v3)
DROP TYPE IF EXISTS movement_type CASCADE;

-- NO eliminar tablas base - las migraciones v3 las necesitan
-- Solo eliminar tablas específicas que no se usan en v3
DROP TABLE IF EXISTS supply_movements CASCADE;
DROP TABLE IF EXISTS stocks CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS project_managers CASCADE;
DROP TABLE IF EXISTS managers CASCADE;
-- Mantener tablas base necesarias para v3:
-- crop_commercializations, workorder_items, workorders, project_investors, 
-- investors, supplies, labors, labor_categories, labor_types, categories, 
-- types, lots, crops, fields, lease_types, projects, campaigns, customers, 
-- users, project_dollar_values, providers, app_parameters, fx_rates
-- NO eliminar schema_migrations - la herramienta de migración la necesita
-- DROP TABLE IF EXISTS schema_migrations CASCADE;

-- NO eliminar secuencias de tablas base - las migraciones v3 las necesitan
-- Solo eliminar secuencias de tablas que se eliminaron arriba
DROP SEQUENCE IF EXISTS managers_id_seq CASCADE;
DROP SEQUENCE IF EXISTS invoices_id_seq CASCADE;
DROP SEQUENCE IF EXISTS stocks_id_seq CASCADE;
DROP SEQUENCE IF EXISTS supply_movements_id_seq CASCADE;

-- Eliminar funciones
DROP FUNCTION IF EXISTS public.update_timestamp() CASCADE;
DROP FUNCTION IF EXISTS public.get_app_parameter(character varying) CASCADE;
DROP FUNCTION IF EXISTS public.get_app_parameter_decimal(character varying) CASCADE;
DROP FUNCTION IF EXISTS public.get_app_parameter_integer(character varying) CASCADE;
DROP FUNCTION IF EXISTS public.get_campaign_closure_days() CASCADE;
DROP FUNCTION IF EXISTS public.get_default_fx_rate() CASCADE;
DROP FUNCTION IF EXISTS public.get_iva_percentage() CASCADE;
DROP FUNCTION IF EXISTS public.get_project_dollar_value(bigint, character varying) CASCADE;
DROP FUNCTION IF EXISTS public.calculate_campaign_closing_date(date) CASCADE;

COMMIT;

-- ========================================
-- MIGRACIÓN 000072: REVERTIR VALORES DE CÁLCULO
-- ========================================

-- Eliminar funciones
DROP FUNCTION IF EXISTS get_iva_percentage();
DROP FUNCTION IF EXISTS get_campaign_closure_days();
DROP FUNCTION IF EXISTS get_default_fx_rate();
DROP FUNCTION IF EXISTS calculate_campaign_closing_date(DATE);

-- Eliminar vistas
DROP VIEW IF EXISTS fix_labors_list;
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;

-- Eliminar índices
DROP INDEX IF EXISTS idx_calc_values_key;

-- Eliminar tabla
DROP TABLE IF EXISTS calc_values;

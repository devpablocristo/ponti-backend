-- =========================================================
-- Rollback de la migración 000046: Crear dashboard completo y optimizado
-- =========================================================

-- Eliminar la función del dashboard
DROP FUNCTION IF EXISTS get_dashboard_payload(BIGINT, BIGINT, BIGINT, BIGINT);

-- Eliminar las vistas del dashboard
DROP VIEW IF EXISTS dashboard_card_sowing_view;
DROP VIEW IF EXISTS dashboard_card_harvest_view;
DROP VIEW IF EXISTS dashboard_card_costs_progress_view;
DROP VIEW IF EXISTS dashboard_card_contributions_view;
DROP VIEW IF EXISTS dashboard_card_operating_result_view;

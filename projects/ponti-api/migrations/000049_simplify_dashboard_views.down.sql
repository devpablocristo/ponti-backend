-- =========================================================
-- Rollback: Eliminar vistas simplificadas del dashboard
-- =========================================================

DROP VIEW IF EXISTS dashboard_card_sowing_view;
DROP VIEW IF EXISTS dashboard_card_harvest_view;
DROP VIEW IF EXISTS dashboard_card_costs_progress_view;
DROP VIEW IF EXISTS dashboard_card_contributions_view;
DROP VIEW IF EXISTS dashboard_card_operating_result_view;

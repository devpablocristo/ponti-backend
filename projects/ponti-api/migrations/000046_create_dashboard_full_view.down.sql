-- ROLLBACK MIGRACIÓN 000046: ELIMINAR DASHBOARD_FULL_VIEW

-- Eliminar la vista dashboard_full_view
DROP VIEW IF EXISTS dashboard_full_view CASCADE;

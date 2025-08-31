-- Rollback: restaurar la vista dashboard_view anterior
-- Esta migración restauró la funcionalidad completa del dashboard basada en 000050
-- corrigiendo solo el problema de wi.lot_id

-- Revertir a la migración anterior
DROP VIEW IF EXISTS dashboard_view;

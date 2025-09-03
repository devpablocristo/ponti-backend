-- =====================================================
-- 000064: WORKORDER - Vistas de Cálculo de Órdenes de Trabajo (ROLLBACK)
-- =====================================================
-- Entidad: workorder (Órdenes de Trabajo)
-- Funcionalidad: Eliminar vistas de cálculo de workorders
-- =====================================================

-- Eliminar vista de cálculos de workorders
DROP VIEW IF EXISTS v_calc_workorders CASCADE;

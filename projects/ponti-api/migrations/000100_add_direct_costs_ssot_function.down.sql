-- ========================================
-- MIGRACIÓN 000100: ADD DIRECT COSTS SSOT FUNCTION (DOWN)
-- ========================================
-- 
-- Propósito: Revertir la función SSOT para costos directos totales
-- Fecha: 2025-01-01
-- Autor: Sistema

-- -------------------------------------------------------------------
-- ELIMINAR FUNCIÓN SSOT PARA COSTOS DIRECTOS TOTALES
-- -------------------------------------------------------------------
DROP FUNCTION IF EXISTS v3_calc.direct_costs_total_for_project(bigint);

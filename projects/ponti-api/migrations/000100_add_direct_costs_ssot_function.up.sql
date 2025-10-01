-- ========================================
-- MIGRACIÓN 000100: ADD DIRECT COSTS SSOT FUNCTION (UP)
-- ========================================
-- 
-- Propósito: Crear función SSOT para costos directos totales
-- Fórmula: SUM(direct_cost_usd) desde v3_workorder_metrics (sin movimientos internos)
-- Fecha: 2025-01-01
-- Autor: Sistema
-- 
-- Nota: Función reutilizable para costos directos en múltiples vistas

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DIRECTOS TOTALES
-- -------------------------------------------------------------------
-- Esta función calcula los costos directos totales usando v3_workorder_metrics (sin movimientos internos)
-- Fórmula: SUM(direct_cost_usd) desde v3_workorder_metrics
CREATE OR REPLACE FUNCTION v3_calc.direct_costs_total_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos: desde v3_workorder_metrics (ya corregidos, sin movimientos internos)
    (SELECT COALESCE(SUM(wm.direct_cost_usd), 0)::double precision
     FROM public.v3_workorder_metrics wm
     WHERE wm.project_id = p_project_id)
  , 0)::double precision
$$;

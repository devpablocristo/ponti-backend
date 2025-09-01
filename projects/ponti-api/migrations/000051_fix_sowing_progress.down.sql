-- ========================================
-- MIGRACIÓN 000051 DOWN: REVERTIR CORRECCIÓN DEL AVANCE DE SIEMBRA
-- ========================================
-- 
-- Esta migración revierte la corrección del avance de siembra
-- y restaura la vista dashboard_view original de la migración 000050
-- 
-- MÓDULOS NO AFECTADOS:
-- ✅ Avance de costos
-- ✅ Avance de cosecha  
-- ✅ Avance de aportes
-- ✅ Resultado operativo
-- ✅ Balance de gestión
-- ✅ Incidencia de costos por cultivo
-- ✅ Indicadores operativos
-- 
-- SOLO SE REVIERTE:
-- 🔧 Avance de siembra (CTE sowing y cálculo de porcentaje)
-- ========================================

-- Eliminar la vista corregida
DROP VIEW IF EXISTS dashboard_view;

-- Restaurar la vista original de la migración 000050
-- (Esto debe ejecutarse manualmente ejecutando la migración 000050)

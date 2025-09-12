-- ========================================
-- MIGRACIÓN 000076: ELIMINAR VISTA v3_workorder_metrics (DOWN)
-- ========================================
-- 
-- Objetivo: Eliminar la vista creada en la migración UP
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

-- -------------------------------------------------------------------
-- v3_workorder_metrics: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_workorder_metrics;

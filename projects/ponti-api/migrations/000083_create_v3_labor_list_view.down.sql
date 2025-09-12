-- ========================================
-- MIGRACIÓN 000082: ELIMINAR VISTA v3_labor_list (DOWN)
-- ========================================
-- 
-- Objetivo: Eliminar la vista creada en la migración UP
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

-- -------------------------------------------------------------------
-- v3_labor_list: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_labor_list;



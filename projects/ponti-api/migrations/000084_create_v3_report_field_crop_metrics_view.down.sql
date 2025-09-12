-- ========================================
-- MIGRACIÓN 000083: ELIMINAR VISTA v3_report_field_crop_metrics_view (DOWN)
-- ========================================
-- 
-- Objetivo: Eliminar la vista creada en la migración UP
-- Fecha: 2025-09-12
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics_view;



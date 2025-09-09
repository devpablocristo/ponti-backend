-- ========================================
-- MIGRACIÓN 000072: ELIMINAR VISTAS BASE DE REPORTES
-- Entidad: report (Eliminar vistas base para reportes)
-- Funcionalidad: Revertir creación de vistas base para reportes
-- ========================================

-- ========================================
-- 1. ELIMINAR VISTA REPORT_FIELD_CROP_METRICS_VIEW_V2
-- ========================================
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

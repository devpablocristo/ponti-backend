-- ========================================
-- MIGRACIÓN 000072: ELIMINAR VISTA BÁSICA DE REPORTES
-- Entidad: report (Eliminar vista básica para reportes)
-- Funcionalidad: Revertir creación de vista básica para reportes
-- ========================================

-- ========================================
-- 1. ELIMINAR VISTA BÁSICA REPORT_FIELD_CROP_METRICS_VIEW_V2
-- ========================================
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

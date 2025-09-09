-- ========================================
-- MIGRACIÓN 000075: REVERTIR CONSISTENCIA DE ÁREAS EN VISTAS
-- Entidad: views (Revertir correcciones de áreas en vistas)
-- Funcionalidad: Revertir correcciones de consistencia de áreas
-- ========================================

-- Revertir la vista a la versión anterior
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

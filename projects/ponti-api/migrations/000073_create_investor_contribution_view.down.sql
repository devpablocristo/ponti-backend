-- ========================================
-- MIGRACIÓN 000073: ELIMINAR VISTAS DE REPORTES COMPLETAS
-- Entidad: report + investor (Eliminar vistas completas para reportes)
-- Funcionalidad: Revertir creación de vistas completas para reportes
-- ========================================

-- ========================================
-- 1. ELIMINAR VISTA DE APORTES DE INVERSORES
-- ========================================
DROP VIEW IF EXISTS investor_contribution_data_view;

-- ========================================
-- 2. ELIMINAR VISTA OPTIMIZADA DE MÉTRICAS
-- ========================================
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;

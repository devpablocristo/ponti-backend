-- ========================================
-- MIGRACIÓN 000074: ELIMINAR VISTA DE APORTES DE INVERSORES
-- Entidad: investor (Eliminar vista para datos de aportes de inversores)
-- Funcionalidad: Revertir creación de vista para reportes de aportes
-- ========================================

-- Drop the unified investor contribution data view
DROP VIEW IF EXISTS investor_contribution_data_view;

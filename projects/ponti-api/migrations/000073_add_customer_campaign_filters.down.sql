-- ========================================
-- MIGRACIÓN 000073: REVERTIR FILTROS CUSTOMER Y CAMPAIGN A VISTAS
-- Entidad: views (Revertir vistas con filtros de cliente y campaña)
-- Funcionalidad: Eliminar customer_id y campaign_id de las vistas
-- ========================================

-- Revertir vista fix_lot_list
DROP VIEW IF EXISTS fix_lot_list;

-- Revertir vista fix_lots_metrics  
DROP VIEW IF EXISTS fix_lots_metrics;

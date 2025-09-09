-- ========================================
-- ROLLBACK: MIGRACIÓN 000074 - CORREGIR VISTAS BASE PARA DASHBOARD
-- ========================================

-- Revertir fix_lot_list a su estado anterior
DROP VIEW IF EXISTS fix_lot_list;

-- Revertir base_lease_calculations_view a su estado anterior
DROP VIEW IF EXISTS base_lease_calculations_view;

-- Revertir base_admin_costs_view a su estado anterior
DROP VIEW IF EXISTS base_admin_costs_view;

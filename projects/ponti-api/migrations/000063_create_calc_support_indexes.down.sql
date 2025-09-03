-- =====================================================
-- 000063: SUPPORT INDEXES - Revertir Índices de Soporte
-- =====================================================

-- Eliminar índices de soporte
DROP INDEX IF EXISTS idx_workorders_labor_notdel;
DROP INDEX IF EXISTS idx_workorders_effarea_notdel;
DROP INDEX IF EXISTS idx_workorder_items_supply_notdel;
DROP INDEX IF EXISTS idx_labors_proj_notdel;
DROP INDEX IF EXISTS idx_supplies_proj_notdel;
DROP INDEX IF EXISTS idx_workorders_lot_id_harvest_notdel;
DROP INDEX IF EXISTS idx_commercializations_p_c_date_notdel;

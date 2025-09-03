-- =====================================================
-- 000063: SUPPORT INDEXES - Índices de Soporte
-- =====================================================
-- Entidad: Database Performance (Rendimiento de Base de Datos)
-- Funcionalidad: Índices para optimizar consultas de cálculos
-- =====================================================

-- Índices de soporte para cálculos (mínimos, parciales donde sea útil)
-- workorders(labor_id) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_workorders_labor_notdel ON workorders(labor_id) WHERE deleted_at IS NULL;

-- workorders(effective_area) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_workorders_effarea_notdel ON workorders(effective_area) WHERE deleted_at IS NULL;

-- workorder_items(supply_id) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_workorder_items_supply_notdel ON workorder_items(supply_id) WHERE deleted_at IS NULL;

-- labors(project_id) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_labors_proj_notdel ON labors(project_id) WHERE deleted_at IS NULL;

-- supplies(project_id) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_supplies_proj_notdel ON supplies(project_id) WHERE deleted_at IS NULL;

-- harvests(lot_id) WHERE deleted_at IS NULL (si existe la tabla)
-- Como no existe tabla harvests, creamos índice en workorders para harvests
CREATE INDEX IF NOT EXISTS idx_workorders_lot_id_harvest_notdel ON workorders(lot_id) WHERE deleted_at IS NULL;

-- crop_commercializations(project_id, crop_id, created_at) WHERE deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_commercializations_p_c_date_notdel ON crop_commercializations(project_id, crop_id, created_at) WHERE deleted_at IS NULL;

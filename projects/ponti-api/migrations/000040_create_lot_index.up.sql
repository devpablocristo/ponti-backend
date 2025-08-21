-- lotes / campos / proyectos
CREATE INDEX IF NOT EXISTS idx_lots_field_deleted ON lots(field_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fields_project ON fields(project_id);

-- cultivos
CREATE INDEX IF NOT EXISTS idx_crops_id ON crops(id);

-- órdenes de trabajo y relaciones utilizadas en la vista
CREATE INDEX IF NOT EXISTS idx_workorders_lot_notdel ON workorders(lot_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workorders_labor_notdel ON workorders(labor_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_labors_id_notdel ON labors(id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_id_name_notdel ON categories(id, name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workorder_items_wo_notdel ON workorder_items(workorder_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_supplies_id_notdel ON supplies(id) WHERE deleted_at IS NULL;

-- comercializaciones (para DISTINCT ON por (project_id, crop_id))
CREATE INDEX IF NOT EXISTS idx_crop_comm_proj_crop_updated
ON crop_commercializations(project_id, crop_id, updated_at DESC) WHERE deleted_at IS NULL;

-- lot_dates: soporte para ON CONFLICT (lot_id, sequence) respetando soft delete
CREATE UNIQUE INDEX IF NOT EXISTS ux_lot_dates_lot_seq_notdel
ON lot_dates(lot_id, sequence) WHERE deleted_at IS NULL; 
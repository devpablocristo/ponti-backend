-- Revertir índices creados en 000026_create_lot_index.up.sql

-- lot_dates: restricción única
DROP INDEX IF EXISTS ux_lot_dates_lot_seq_notdel;

-- comercializaciones
DROP INDEX IF EXISTS idx_crop_comm_proj_crop_updated;

-- insumos
DROP INDEX IF EXISTS idx_supplies_id_notdel;

-- elementos de orden de trabajo
DROP INDEX IF EXISTS idx_workorder_items_wo_notdel;

-- categorías
DROP INDEX IF EXISTS idx_categories_id_name_notdel;

-- labores
DROP INDEX IF EXISTS idx_labors_id_notdel;

-- órdenes de trabajo
DROP INDEX IF EXISTS idx_workorders_labor_notdel;
DROP INDEX IF EXISTS idx_workorders_lot_notdel;

-- cultivos
DROP INDEX IF EXISTS idx_crops_id;

-- campos / proyectos
DROP INDEX IF EXISTS idx_fields_project;

-- lotes
DROP INDEX IF EXISTS idx_lots_field_deleted; 
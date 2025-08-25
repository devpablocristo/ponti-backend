-- =======================
-- RESET DE BASE DE DATOS PARA LIST LOT
-- =======================
-- Este archivo limpia todos los datos de prueba antes de insertar los nuevos seeds

-- Eliminar datos de workorders (dependen de labors, lots, etc.)
DELETE FROM workorder_items WHERE deleted_at IS NULL;
DELETE FROM workorders WHERE deleted_at IS NULL;

-- Eliminar datos de labors (dependen de projects, categories)
DELETE FROM labors WHERE deleted_at IS NULL;

-- Eliminar datos de supplies (dependen de projects, types, categories)
DELETE FROM supplies WHERE deleted_at IS NULL;

-- Eliminar datos de lot_dates (dependen de lots)
DELETE FROM lot_dates WHERE deleted_at IS NULL;

-- Eliminar datos de lots (dependen de fields, crops)
DELETE FROM lots WHERE deleted_at IS NULL;

-- Eliminar datos de fields (dependen de projects, lease_types)
DELETE FROM fields WHERE deleted_at IS NULL;

-- Eliminar datos de crop_commercializations (dependen de projects, crops)
DELETE FROM crop_commercializations WHERE deleted_at IS NULL;

-- Eliminar datos de project_investors (dependen de projects, investors)
DELETE FROM project_investors WHERE deleted_at IS NULL;

-- Eliminar datos de project_managers (dependen de projects, managers)
DELETE FROM project_managers WHERE deleted_at IS NULL;

-- Eliminar datos de projects (dependen de customers, campaigns)
DELETE FROM projects WHERE deleted_at IS NULL;

-- Eliminar datos de campaigns
DELETE FROM campaigns WHERE deleted_at IS NULL;

-- Eliminar datos de customers
DELETE FROM customers WHERE deleted_at IS NULL;

-- Eliminar datos de managers
DELETE FROM managers WHERE deleted_at IS NULL;

-- Eliminar datos de investors
DELETE FROM investors WHERE deleted_at IS NULL;

-- Eliminar datos de categories (dependen de types)
DELETE FROM categories WHERE deleted_at IS NULL;

-- Eliminar datos de types
DELETE FROM types WHERE deleted_at IS NULL;

-- Eliminar datos de lease_types
DELETE FROM lease_types WHERE deleted_at IS NULL;

-- Eliminar datos de crops
DELETE FROM crops WHERE deleted_at IS NULL;

-- Eliminar datos de users
DELETE FROM users WHERE id = 123;

-- Resetear secuencias
ALTER SEQUENCE users_id_seq RESTART WITH 124;
ALTER SEQUENCE projects_id_seq RESTART WITH 1;
ALTER SEQUENCE fields_id_seq RESTART WITH 1;
ALTER SEQUENCE lots_id_seq RESTART WITH 1;
ALTER SEQUENCE labors_id_seq RESTART WITH 1;
ALTER SEQUENCE workorders_id_seq RESTART WITH 1;
ALTER SEQUENCE supplies_id_seq RESTART WITH 1;
ALTER SEQUENCE types_id_seq RESTART WITH 1;
ALTER SEQUENCE categories_id_seq RESTART WITH 1;
ALTER SEQUENCE lease_types_id_seq RESTART WITH 1;
ALTER SEQUENCE crops_id_seq RESTART WITH 1;
ALTER SEQUENCE crop_commercializations_id_seq RESTART WITH 1;

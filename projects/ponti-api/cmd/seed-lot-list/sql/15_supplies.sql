-- =======================
-- SUPPLIES PARA LIST LOT
-- =======================
-- Insumos para el proyecto

INSERT INTO supplies (id, name, project_id, type_id, category_id, unit_id, price, created_at, updated_at) VALUES
(1, 'Semilla Maíz DK 72-10', 1, 1, 1, 1, 1.65, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Supplies creados' as status, COUNT(*) as total FROM supplies;

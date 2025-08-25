-- =======================
-- CROPS PARA LIST LOT
-- =======================
-- Cultivos para el proyecto

INSERT INTO crops (id, name, created_at, updated_at) VALUES
(1, 'Soja', NOW(), NOW()),
(2, 'Maíz', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Crops creados' as status, COUNT(*) as total FROM crops;

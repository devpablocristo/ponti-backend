-- =======================
-- TIPOS DE ARRIENDO PARA LIST LOT
-- =======================
-- Tipos de arriendo para los campos

INSERT INTO lease_types (id, name, created_at, updated_at) VALUES
(1, 'Cantidad Fija', NOW(), NOW()),
(2, 'Porcentaje', NOW(), NOW()),
(3, 'Ambos', NOW(), NOW()),
(4, 'Sin Arriendo', NOW(), NOW());

-- Verificar inserción
SELECT '✅ Tipos de arriendo creados' as status, COUNT(*) as total FROM lease_types;

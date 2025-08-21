-- Crear proveedores de prueba para el sistema
-- Los proveedores se crean para diferentes tipos de insumos agrícolas
-- Incluye proveedores de semillas, agroquímicos, fertilizantes y maquinaria especializada

INSERT INTO providers (name, created_by, updated_by, created_at, updated_at) VALUES
('Proveedor Semillas Premium', 2, 2, NOW(), NOW()),
('Agroquímicos del Sur', 2, 2, NOW(), NOW()),
('Fertilizantes Norte', 2, 2, NOW(), NOW()),
('Maquinaria Agrícola Central', 2, 2, NOW(), NOW()),
('Insumos Agrícolas Express', 2, 2, NOW(), NOW()),
('Semillas Genéticas Avanzadas', 2, 2, NOW(), NOW()),
('Agroquímicos Ecológicos', 2, 2, NOW(), NOW()),
('Fertilizantes Orgánicos', 2, 2, NOW(), NOW()),
('Maquinaria de Precisión', 2, 2, NOW(), NOW()),
('Insumos Tecnológicos', 2, 2, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Proveedores creados:' as status, COUNT(*) as total FROM providers;

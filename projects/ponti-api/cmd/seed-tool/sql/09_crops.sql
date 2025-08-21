-- Crear cultivos de prueba para el sistema
-- Los cultivos se crean para asociar con lotes y órdenes de trabajo
-- Incluye 10 cultivos con nombres únicos para diferentes tipos de producción agrícola

INSERT INTO crops (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Crop 1', 2, 2, NOW(), NOW()),
  ('Crop 2', 2, 2, NOW(), NOW()),
  ('Crop 3', 2, 2, NOW(), NOW()),
  ('Crop 4', 2, 2, NOW(), NOW()),
  ('Crop 5', 2, 2, NOW(), NOW()),
  ('Crop 6', 2, 2, NOW(), NOW()),
  ('Crop 7', 2, 2, NOW(), NOW()),
  ('Crop 8', 2, 2, NOW(), NOW()),
  ('Crop 9', 2, 2, NOW(), NOW()),
  ('Crop 10', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Cultivos creados:' as status, COUNT(*) as total FROM crops; 
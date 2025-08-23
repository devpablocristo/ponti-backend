-- Crear campañas de prueba para el sistema
-- Las campañas se crean para asociar con proyectos agrícolas
-- Incluye 10 campañas con nombres únicos para diferentes temporadas de cultivo

INSERT INTO campaigns (name, created_by, updated_by, created_at, updated_at)
VALUES 
  ('Campaign 1', 2, 2, NOW(), NOW()),
  ('Campaign 2', 2, 2, NOW(), NOW()),
  ('Campaign 3', 2, 2, NOW(), NOW()),
  ('Campaign 4', 2, 2, NOW(), NOW()),
  ('Campaign 5', 2, 2, NOW(), NOW()),
  ('Campaign 6', 2, 2, NOW(), NOW()),
  ('Campaign 7', 2, 2, NOW(), NOW()),
  ('Campaign 8', 2, 2, NOW(), NOW()),
  ('Campaign 9', 2, 2, NOW(), NOW()),
  ('Campaign 10', 2, 2, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Verificar inserción
SELECT '✅ Campañas creadas:' as status, COUNT(*) as total FROM campaigns; 